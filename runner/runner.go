package runner

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/juju/errors"
	"github.com/mattn/go-colorable"
)

const startChannelSize = 1000

type Runner struct {
	path   string
	config *configYaml

	// These are the final resolved list of directories (after wildcards are resolved)
	ignoredDirectories   []string
	exclusiveDirectories []string

	startChannel chan string
	stopChannel  chan bool

	logger     *log.Logger
	mainLog    logFunc
	watcherLog logFunc
	runnerLog  logFunc
	buildLog   logFunc
	appLog     logFunc
	logColors  map[string]string
}

// SetPath sets the runner's config path.
func (r *Runner) SetPath(path string) error {
	r.path = path
	return errors.Trace(r.resolveConfig())
}

func (r *Runner) run() (bool, error) {
	r.runnerLog("Running...")

	cmd := exec.Command(r.buildPath())
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return false, errors.Trace(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return false, errors.Trace(err)
	}
	if err := cmd.Start(); err != nil {
		return false, errors.Trace(err)
	}

	go io.Copy(appLogWriter{r}, stderr) // nolint:errcheck
	go io.Copy(appLogWriter{r}, stdout) // nolint:errcheck

	go func(r *Runner) {
		<-r.stopChannel
		pid := cmd.Process.Pid
		r.runnerLog("Killing PID %d", pid)
		if err := cmd.Process.Kill(); err != nil {
			r.runnerLog("Process kill failed %s", err.Error())
		}
	}(r)

	return true, nil
}

func (r *Runner) flushEvents() {
	for {
		select {
		case eventName := <-r.startChannel:
			r.mainLog("receiving event %s", eventName)
		default:
			return
		}
	}
}

func (r *Runner) buildPath() string {
	p := filepath.Join(r.config.TmpPath, r.config.BuildName)
	if runtime.GOOS == "windows" && filepath.Ext(p) != ".exe" {
		p += ".exe"
	}
	return p
}

func (r *Runner) buildErrorsFilePath() string {
	return filepath.Join(r.config.TmpPath, r.config.BuildLog)
}

func (r *Runner) build() (string, bool, error) {
	r.buildLog("Building...")
	cmd := exec.Command("go", "build", "-o", r.buildPath(), r.config.Root)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", false, errors.Trace(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", false, errors.Trace(err)
	}
	if err := cmd.Start(); err != nil {
		return "", false, errors.Trace(err)
	}
	if _, err := io.Copy(os.Stdout, stdout); err != nil {
		return "", false, errors.Trace(err)
	}
	errBuf, _ := ioutil.ReadAll(stderr)
	if err := cmd.Wait(); err != nil {
		return string(errBuf), false, nil
	}
	return "", true, nil
}

// nolint:gocognit
func (r *Runner) start() {
	loopIndex := 0
	buildDelay := r.config.BuildDelay
	started := false

	go func(r *Runner) {
		for {
			loopIndex++
			r.mainLog("Waiting (loop %d)...", loopIndex)
			eventName := <-r.startChannel

			r.mainLog("receiving first event %s", eventName)
			r.mainLog("sleeping for %d milliseconds", buildDelay/1e6) // nolint:gomnd
			time.Sleep(buildDelay)
			r.mainLog("flushing events")

			r.flushEvents()

			r.mainLog("Started! (%d Goroutines)", runtime.NumGoroutine())
			err := r.removeBuildErrorsLog()
			if err != nil {
				r.mainLog(err.Error())
			}

			buildFailed := false
			if r.shouldRebuild(eventName) { // nolint:nestif
				errorMessage, ok, err := r.build()
				if !ok || err != nil {
					buildFailed = true
					if err != nil {
						r.mainLog("Build Failed: \n %s %s", errorMessage, err.Error())
					} else {
						r.mainLog("Build Failed: \n %s", errorMessage)
					}
					if !started {
						os.Exit(1)
					}
					r.createBuildErrorsLog(errorMessage)
				}
			}

			if !buildFailed {
				if started {
					r.stopChannel <- true
				}
				if _, err := r.run(); err != nil {
					r.mainLog("Run Failed: \n %s", err.Error())
				}
			}

			started = true
			r.mainLog(strings.Repeat("-", 20))
		}
	}(r)
}

// Watches for file changes in the root directory.
// After each file system event it builds and (re)starts the application.
func (r *Runner) Start() error {
	if err := r.initFolders(); err != nil {
		return errors.Trace(err)
	}
	if err := r.watch(); err != nil {
		return errors.Trace(err)
	}

	r.start()
	r.startChannel <- "/"
	<-make(chan int)

	return nil
}

// New returns a new instance of Runner.
func New() (*Runner, error) {
	r := &Runner{
		path: "espal-run.yml",

		startChannel: make(chan string, startChannelSize),
		stopChannel:  make(chan bool),

		logger: log.New(colorable.NewColorableStderr(), "", 0),
	}
	if err := r.resolveConfig(); err != nil {
		return nil, errors.Trace(err)
	}
	r.loadColors()
	r.mainLog = r.newLogFunc("main")
	r.watcherLog = r.newLogFunc("watcher")
	r.runnerLog = r.newLogFunc("runner")
	r.buildLog = r.newLogFunc("build")
	r.appLog = r.newLogFunc("app")
	return r, nil
}
