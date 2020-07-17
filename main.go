package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/espal-digital-development/espal-run/cockroach"
	"github.com/espal-digital-development/espal-run/configchecker"
	"github.com/espal-digital-development/espal-run/gopackage"
	"github.com/espal-digital-development/espal-run/openssl"
	"github.com/espal-digital-development/espal-run/projectcreator"
	"github.com/espal-digital-development/espal-run/qtcbuilder"
	"github.com/espal-digital-development/espal-run/randomstring"
	"github.com/espal-digital-development/espal-run/runner"
	"github.com/espal-digital-development/espal-run/sslgenerator"
	"github.com/espal-digital-development/espal-run/storeintegrity"
	"github.com/espal-digital-development/espal-run/system"
	"github.com/juju/errors"
)

const (
	randomPasswordLength    = 32
	defaultServerPath       = "./app/server"
	defaultDatabasePath     = "./app/database"
	defaultStoresPath       = "./stores"
	defaultConfigPath       = "./app/config.yml"
	defaultDatabaseRootUser = "root"
	defaultDatabaseHTTPUser = "espal"
	defaultSoftLimitMax     = 20000
	defaultSoftLimitCur     = 20000
)

// nolint:gochecknoglobals
var (
	cwd               string
	createProjectPath string
	appPath           string
	fullConfigFile    bool
	runChecks         bool
	runQTC            bool
	skipDB            bool
	resetDB           bool
	dbPortStart       int
	dbNodes           int
)

// TODO :: Some problems with the command is the paths that might've been chosen in the config.yml. If they are totally
//   different; it may cause discrepancies for this command.
// TODO :: Security inspections of the area where the espal app is ran.
//   Check mod values and if the environment has dangerous settings set.
// TODO :: Add support for blending xargs parameters and ENV variables.

func parseFlags() {
	flag.StringVar(&createProjectPath, "create-project", "", "Create a new espal app project")
	flag.StringVar(&appPath, "app-path", "", "Target app path")
	flag.BoolVar(&runChecks, "full-config-file", false, "Generate the most complete config file possible with "+
		"default values, unless overridden by the prompter")
	flag.BoolVar(&runChecks, "run-checks", false, "Run the checks with inspectors")
	flag.BoolVar(&runQTC, "run-qtc", false, "Run the QuickTemplate Compiler")
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

// nolint:funlen,gocognit
func main() {
	parseFlags()

	// nolint:nestif
	if createProjectPath != "" {
		var err error
		createProjectPath, err = filepath.Abs(filepath.FromSlash(createProjectPath))
		if err != nil {
			log.Fatal(errors.ErrorStack(err))
		}
		projectCreator, err := projectcreator.New()
		if err != nil {
			log.Fatal(errors.ErrorStack(err))
		}
		if err := projectCreator.Do(createProjectPath); err != nil {
			log.Fatal(errors.ErrorStack(err))
		}
	} else if appPath != "" {
		var err error
		appPath, err = filepath.Abs(appPath)
		if err != nil {
			log.Fatal(errors.ErrorStack(err))
		}
		if err := os.Chdir(appPath); err != nil {
			log.Fatal(errors.ErrorStack(err))
		}
	}

	if err := setCwd(); err != nil {
		log.Fatal(errors.ErrorStack(err))
	}

	ok, err := pathIsAnApp(cwd)
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
	if !ok {
		log.Fatalf("there doesn't seem to be an app at %s", cwd)
	}

	system, err := system.New()
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
	if err := system.SetSoftUlimit(defaultSoftLimitMax, defaultSoftLimitCur); err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
	if err := system.InstallBinaryDependencies(); err != nil {
		log.Fatal(errors.ErrorStack(err))
	}

	randomString, err := randomstring.New()
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
	if err := checkSSL(); err != nil {
		log.Fatal(errors.ErrorStack(err))
	}

	if !skipDB {
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
	if runQTC {
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

	configChecker, err := configchecker.New(randomString, fullConfigFile)
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

func pathIsAnApp(path string) (bool, error) {
	modFilePath := filepath.FromSlash(path + "/go.mod")
	_, err := os.Stat(modFilePath)
	if err != nil && !os.IsNotExist(err) {
		return false, errors.Trace(err)
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	data, err := ioutil.ReadFile(modFilePath)
	if err != nil {
		return false, errors.Trace(err)
	}
	if bytes.Contains(data, []byte("github.com/espal-digital-development/espal-core")) {
		return true, nil
	}
	return false, nil
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
