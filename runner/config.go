package runner

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/juju/errors"
	"github.com/mattn/go-zglob"
	"gopkg.in/yaml.v2"
)

const defaultBuildDelay = 300 * time.Millisecond

type configYaml struct {
	Root               string
	TmpPath            string        `yaml:"tmpPath"`
	BuildName          string        `yaml:"buildName"`
	BuildLog           string        `yaml:"buildLog"`
	VerboseWatching    bool          `yaml:"verboseWatching"`
	SmartRebuildQtpl   bool          `yaml:"smartRebuildQtpl"`
	ValidExtensions    []string      `yaml:"validExtensions"`
	InvalidExtensions  []string      `yaml:"invalidExtensions"`
	IgnoredFiles       []string      `yaml:"ignoredFiles"`
	IgnoredDirectories []string      `yaml:"ignoredDirectories"`
	BuildDelay         time.Duration `yaml:"buildDelay"`
	Colorize           bool          `yaml:"colorize"`
	LogColors          logColorsYaml `yaml:"logColors"`
}

type logColorsYaml struct {
	Main    string `yaml:"main"`
	Build   string `yaml:"build"`
	Runner  string `yaml:"runner"`
	Watcher string `yaml:"watcher"`
	App     string `yaml:"app"`
}

func (r *Runner) fillDefaultConfig() {
	r.config = &configYaml{
		Root:               ".",
		TmpPath:            "./tmp",
		BuildName:          "espal-core",
		BuildLog:           "errors.log",
		SmartRebuildQtpl:   true,
		ValidExtensions:    []string{"go", "qtpl", "js", "css"},
		InvalidExtensions:  []string{"tmp", "lock", "log", "yml", "json"},
		IgnoredDirectories: []string{"tmp", "node_modules"},
		IgnoredFiles:       []string{"*_test.go"},
		BuildDelay:         defaultBuildDelay,
		Colorize:           true,
		LogColors: logColorsYaml{
			Main:    "cyan",
			Build:   "yellow",
			Runner:  "green",
			Watcher: "magenta",
		},
	}
}

func (r *Runner) resolveConfig() error {
	if _, err := os.Stat(r.path); err != nil {
		return errors.Trace(err)
	}
	bytes, err := ioutil.ReadFile(r.path)
	if err != nil {
		return errors.Trace(err)
	}

	r.fillDefaultConfig()

	if err := yaml.Unmarshal(bytes, r.config); err != nil {
		return errors.Trace(err)
	}

	for k := range r.config.ValidExtensions {
		if !strings.HasPrefix(r.config.ValidExtensions[k], ".") {
			r.config.ValidExtensions[k] = "." + r.config.ValidExtensions[k]
		}
	}
	for k := range r.config.InvalidExtensions {
		if !strings.HasPrefix(r.config.InvalidExtensions[k], ".") {
			r.config.InvalidExtensions[k] = "." + r.config.InvalidExtensions[k]
		}
	}

	r.ignoredDirectories = []string{}
	for _, dir := range r.config.IgnoredDirectories {
		dir := strings.TrimSpace(dir)
		if strings.Contains(dir, "*") {
			dirs, err := zglob.Glob(strings.TrimSpace(dir))
			if err != nil {
				r.runnerLog(err.Error())
			}
			if len(dirs) > 0 {
				r.ignoredDirectories = append(r.ignoredDirectories, dirs...)
			}
		} else {
			r.ignoredDirectories = append(r.ignoredDirectories, dir)
		}
	}

	r.config.TmpPath, err = filepath.Abs(r.config.TmpPath)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}
