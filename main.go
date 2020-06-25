package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/espal-digital-development/espal-run/cockroach"
	"github.com/juju/errors"
)

// TODO :: Some problems with the command is the paths that might've been
// chosen in the config.yml. If they are totally different; it may cause
// discrepancies for this command.

// TODO :: Detect not being in a project directory. Or maybe give flag
// option to target the project directory/directories.

const (
	linuxOS  = "linux"
	darwinOS = "darwin"
	// windowsOS = "windows"

	defaultDesiredNodes  = 1
	randomPasswordLength = 32
)

var (
	cwd                string
	runChecks          bool
	skipQTC            bool
	resetDB            bool
	homebrewInstalled  bool
	linuxbrewInstalled bool
)

func main() {
	flag.BoolVar(&runChecks, "run-checks", false, "Run the checks with inspectors")
	flag.BoolVar(&skipQTC, "skip-qtc", false, "Don't run the QuickTemplate Compiler")
	flag.BoolVar(&resetDB, "reset-db", false, "Reset the database")
	flag.Parse()

	var err error
	cwd, err = os.Getwd()
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	}

	if !strings.Contains("linux,darwin,windows", runtime.GOOS) {
		log.Printf("Unsupported OS `%s` detected. Assuming Linux-style actions from this point on.\n", runtime.GOOS)
	}

	if runtime.GOOS == darwinOS {
		out, _ := exec.Command("which", "brew").CombinedOutput()
		homebrewInstalled = bytes.Contains(out, []byte("/brew"))
	} else if runtime.GOOS == linuxOS {
		out, _ := exec.Command("which", "brew").CombinedOutput()
		linuxbrewInstalled = bytes.Contains(out, []byte("/brew"))
	}

	setSoftUlimit()
	// TODO :: Generate localhost.crt/localhost.key on-the-fly (even if openssh etc. is still needed)
	// (every OS probably needs a total custom variance here too).
	if err := generateTLSFiles(); err != nil {
		log.Fatal(errors.ErrorStack(err))
	}

	// TODO :: Haproxy as well for full fanciness?
	cockroach, err := cockroach.New()
	if err != nil {
		log.Fatal(errors.Trace(err))
	}
	// TODO :: Auto-detect info based on existing config.yml?
	cockroach.SetDesiredNodes(defaultDesiredNodes)
	cockroach.SetRootUser("root")                                 // TODO :: Random generate user
	cockroach.SetHTTPUser("espal")                                // TODO :: Random generate user
	cockroach.SetHTTPPassword(randomString(randomPasswordLength)) // TODO :: Something safer, like `openssl rand -hex 16`
	if err := cockroach.SetDatabasePath("./app/database"); err != nil {
		log.Fatal(errors.Trace(err))
	}
	cockroach.SetResetDB(resetDB)
	if err := cockroach.Resolve(); err != nil {
		log.Fatal(errors.ErrorStack(err))
	}

	checkStoresIntegrity()
	installPackages()

	if !skipQTC {
		buildQTC()
	}

	if runChecks {
		runAllChecks()
	}

	if err := checkConfig(); err != nil {
		log.Fatal(errors.ErrorStack(err))
	}

	if err := run(); err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
}
