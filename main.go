package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"

	"github.com/espal-digital-development/espal-run/cockroach"
	"github.com/espal-digital-development/espal-run/configchecker"
	"github.com/espal-digital-development/espal-run/gopackage"
	"github.com/espal-digital-development/espal-run/openssl"
	"github.com/espal-digital-development/espal-run/qtcbuilder"
	"github.com/espal-digital-development/espal-run/randomstring"
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
	skipQTC     bool
	resetDB     bool
	dbPortStart int
	dbNodes     int
)

func parseFlags() {
	flag.BoolVar(&runChecks, "run-checks", false, "Run the checks with inspectors")
	flag.BoolVar(&skipQTC, "skip-qtc", false, "Don't run the QuickTemplate Compiler")
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

	setSoftUlimit()

	if err := checkSSL(); err != nil {
		log.Fatal(errors.ErrorStack(err))
	}

	if err := cockroachSetup(randomString); err != nil {
		log.Fatal(errors.ErrorStack(err))
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
	if !skipQTC {
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
	if err := run(); err != nil {
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

func run() error {
	log.Println("Starting the app..")

	cmd := exec.Command("go", "run", cwd+"/main.go")

	// TODO :: When calling TERM on this command, it needs to gracefully stop the espal-core
	// too (prove will be when it reports it's winddown info about how long the server ran for).
	// TODO :: If the output doesn't stop with a newline or throws an error,
	// it probably won't show anything at all. Needs more testing.
	stdOut, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Trace(err)
	}
	stdErr, err := cmd.StderrPipe()
	if err != nil {
		return errors.Trace(err)
	}
	if err := cmd.Start(); err != nil {
		return errors.Trace(err)
	}
	scanner := bufio.NewScanner(stdOut)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}
	errScanner := bufio.NewScanner(stdErr)
	for errScanner.Scan() {
		m := errScanner.Text()
		fmt.Println(m)
	}
	if err := cmd.Wait(); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func installPackages() {
	// TODO :: Maybe just embed them in the vendor and build locally?
	staticCheck := gopackage.New("honnef.co/go/tools/cmd/staticcheck")
	goCheckStyle := gopackage.New("github.com/qiniu/checkstyle/gocheckstyle")
	errCheck := gopackage.New("github.com/kisielk/errcheck")
	qtc := gopackage.New("github.com/valyala/quicktemplate/qtc")
	// TODO :: 77777 The go list calls aren't working correctly due to the Go modules project sub-environment
	staticCheck.InstallIfNeeded(true)
	goCheckStyle.InstallIfNeeded(true)
	errCheck.InstallIfNeeded(true)
	qtc.InstallIfNeeded(true)
}

func setSoftUlimit() {
	// TODO :: Soft limit on Linux needs testing and is probably harder
	// TODO :: Going over the hard limit isn't allowed. Maybe just give a message and continue then
	if runtime.GOOS == "darwin" {
		// TODO :: Could be more efficient to first check if it's already `unlimited` or higher
		// than `10032`. If so; do nothing.
		// TODO :: Because this is a soft limit if might not work as an option for
		// the later-ran espal app?
		// Windows doesn't need this as it already sets it's limit dangerously high by default
		if err := exec.Command("ulimit", "-n", "10032").Run(); err != nil {
			log.Fatal(err)
		}
	}
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
