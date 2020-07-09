package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"syscall"

	"github.com/espal-digital-development/espal-run/cockroach"
	"github.com/espal-digital-development/espal-run/configchecker"
	"github.com/espal-digital-development/espal-run/gopackage"
	"github.com/espal-digital-development/espal-run/openssl"
	"github.com/espal-digital-development/espal-run/qtcbuilder"
	"github.com/espal-digital-development/espal-run/randomstring"
	"github.com/espal-digital-development/espal-run/runner"
	"github.com/espal-digital-development/espal-run/sslgenerator"
	"github.com/espal-digital-development/espal-run/storeintegrity"
	"github.com/juju/errors"
)

// TODO :: Some problems with the command is the paths that might've been
// chosen in the config.yml. If they are totally different; it may cause
// discrepancies for this command.

// TODO :: Detect not being in a project directory. Or maybe give flag
// option to target the project directory/directories.

// TODO :: Security inspections of the area where the espal app is ran.
// Check mod values and if the environment has dangerous settings set.

// TODO :: Add support for blending xargs parameters and ENV variables.

// TODO :: Check macOS Homebrew installed

const (
	randomPasswordLength    = 32
	defaultServerPath       = "./app/server"
	defaultDatabasePath     = "./app/database"
	defaultStoresPath       = "./stores"
	defaultConfigPath       = "./app/config.yml"
	defaultDatabaseRootUser = "root"
	defaultDatabaseHTTPUser = "espal"
)

// nolint:gochecknoglobals
var (
	cwd         string
	runChecks   bool
	allSkips    bool
	skipQTC     bool
	skipDB      bool
	resetDB     bool
	dbPortStart int
	dbNodes     int
)

func parseFlags() {
	flag.BoolVar(&runChecks, "run-checks", false, "Run the checks with inspectors")
	flag.BoolVar(&allSkips, "all-skips", false, "Enable all available skips: skip-qtc, skip-db")
	flag.BoolVar(&skipQTC, "skip-qtc", false, "Don't run the QuickTemplate Compiler")
	flag.BoolVar(&skipDB, "skip-db", false, "Don't run the Cockroach checks, installer and starter")
	flag.BoolVar(&resetDB, "reset-db", false, "Reset the database")
	flag.IntVar(&dbPortStart, "db-port-start", 36257, "Port start range")
	flag.IntVar(&dbNodes, "db-nodes", 1, "Desired amount of nodes")
	flag.Parse()
}

func setCwd() error {
	var err error
	cwd, err = os.Getwd()
	return errors.Trace(err)
}

func main() {
	parseFlags()
	if err := setCwd(); err != nil {
		log.Fatal(errors.ErrorStack(err))
	}

	randomString, err := randomstring.New()
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
	if err := setSoftUlimit(); err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
	if err := checkSSL(); err != nil {
		log.Fatal(errors.ErrorStack(err))
	}

	if !allSkips && !skipDB {
		if err := cockroachSetup(randomString); err != nil {
			log.Fatal(errors.ErrorStack(err))
		}
	}
	storeIntegrity, err := storeintegrity.New()
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
	storeIntegrity.SetPath(defaultStoresPath)
	if err := storeIntegrity.Check(); err != nil {
		log.Fatal(errors.ErrorStack(err))
	}

	installPackages()
	if !allSkips && !skipQTC {
		qtcBuilder, err := qtcbuilder.New()
		if err != nil {
			log.Fatal(errors.ErrorStack(err))
		}
		if err := qtcBuilder.Do(); err != nil {
			log.Fatal(errors.ErrorStack(err))
		}
	}
	if runChecks {
		runAllChecks()
	}

	configChecker, err := configchecker.New(randomString)
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
	configChecker.SetPath(defaultConfigPath)
	if err := configChecker.Do(); err != nil {
		log.Fatal(errors.ErrorStack(err))
	}

	appRunner, err := runner.New()
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
	if err := appRunner.Start(); err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
}

func checkSSL() error {
	openSSL, err := openssl.New()
	if err != nil {
		return errors.Trace(err)
	}
	if err := openSSL.CheckAndInstall(); err != nil {
		return errors.Trace(err)
	}

	sslGenerator, err := sslgenerator.New()
	if err != nil {
		return errors.Trace(err)
	}
	sslGenerator.SetServerPath(defaultServerPath)
	if err := sslGenerator.Do(); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func cockroachSetup(randomString *randomstring.RandomString) error {
	// TODO :: Haproxy as well for full fanciness?
	cockroach, err := cockroach.New()
	if err != nil {
		return errors.Trace(err)
	}
	// TODO :: Auto-detect info based on existing config.yml?
	cockroach.SetDesiredNodes(dbNodes)
	cockroach.SetPortStart(dbPortStart)
	// TODO :: Random generate user
	cockroach.SetRootUser(defaultDatabaseRootUser)
	// TODO :: Random generate user
	cockroach.SetHTTPUser(defaultDatabaseHTTPUser)
	// TODO :: Something safer, like `openssl rand -hex 16`
	cockroach.SetHTTPPassword(randomString.Simple(randomPasswordLength))
	if err := cockroach.SetDatabasePath(defaultDatabasePath); err != nil {
		return errors.Trace(err)
	}
	cockroach.SetResetDB(resetDB)
	if err := cockroach.Resolve(); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func installPackages() {
	qtc := gopackage.New("github.com/valyala/quicktemplate/qtc")
	qtc.InstallIfNeeded(true)
}

func setSoftUlimit() error {
	if runtime.GOOS == "darwin" {
		var rLimit syscall.Rlimit
		rLimit.Max = 10000
		rLimit.Cur = 10000
		err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		if err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func runAllChecks() {
	out, _ := exec.Command("staticcheck", "./...", "|", "grep", "-v", "bindata.go").CombinedOutput()
	if bytes.Contains(out, []byte("\n")) {
		log.Println(string(out))
	}

	removeCoreChecks := regexp.MustCompile(`(?m)^.*?local[\/\\]opt.*?\n`)
	out, _ = exec.Command("errcheck", "./...").CombinedOutput()
	out = bytes.Trim(removeCoreChecks.ReplaceAll(out, []byte("")), "\n")
	// Silly check if there's more than the normal complain-line
	if bytes.Contains(out, []byte("\n")) {
		log.Println(string(out))
	}

	out, _ = exec.Command("gocheckstyle", "-config=.go_style", ".").CombinedOutput()
	if !bytes.Contains(out, []byte("There are no problems")) {
		log.Println(string(out))
	}
}
