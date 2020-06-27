package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/espal-digital-development/espal-run/cockroach"
	"github.com/espal-digital-development/espal-run/gopackage"
	"github.com/espal-digital-development/espal-run/openssl"
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

const (
	randomPasswordLength = 32
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

func main() {
	flag.BoolVar(&runChecks, "run-checks", false, "Run the checks with inspectors")
	flag.BoolVar(&skipQTC, "skip-qtc", false, "Don't run the QuickTemplate Compiler")
	flag.BoolVar(&resetDB, "reset-db", false, "Reset the database")
	flag.IntVar(&dbPortStart, "db-port-start", 26259, "Port start range")
	flag.IntVar(&dbNodes, "db-nodes", 1, "Desired amount of nodes")
	flag.Parse()

	var err error
	cwd, err = os.Getwd()
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	}

	// TODO :: Check macOS Homebrew installed

	setSoftUlimit()

	openSSL, err := openssl.New()
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
	if err := openSSL.CheckAndInstall(); err != nil {
		log.Fatal(errors.ErrorStack(err))
	}

	sslGenerator, err := sslgenerator.New()
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
	sslGenerator.SetServerPath("./app/server")
	if err := sslGenerator.Do(); err != nil {
		log.Fatal(errors.ErrorStack(err))
	}

	// TODO :: Haproxy as well for full fanciness?
	cockroach, err := cockroach.New()
	if err != nil {
		log.Fatal(errors.Trace(err))
	}
	// TODO :: Auto-detect info based on existing config.yml?
	cockroach.SetDesiredNodes(dbNodes)
	cockroach.SetPortStart(dbPortStart)
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

	storeIntegrity, err := storeintegrity.New()
	if err != nil {
		log.Fatal(errors.Trace(err))
	}
	storeIntegrity.SetPath("./stores")
	if err := storeIntegrity.Check(); err != nil {
		log.Fatal(errors.Trace(err))
	}

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

func run() error {
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

func buildQTC() {
	log.Println("Building templates. Please wait..")
	out, err := exec.Command("cd", "pages", "&&", "qtc").CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	if bytes.Contains(out, []byte("error")) {
		log.Println(string(out))
	}
}
