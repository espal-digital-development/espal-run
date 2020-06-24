package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/juju/errors"
)

// TODO :: Some problems with the command is the paths that might've been
// chosen in the config.yml. If they are totally different; it may cause
// discrepancies for this command.

// TODO :: Detect not being in a project directory. Or maybe give flag option to target the project directory/directories

const (
	linuxOS   = "linux"
	darwinOS  = "darwin"
	windowsOS = "windows"
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
	// TODO :: Generate localhost.crt/localhost.key on-the-fly (even if openssh etc. is still needed) (every OS probably needs a total custom variance here too)
	if err := generateTLSFiles(); err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
	if err := configureCockroachDB(); err != nil {
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
