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

const defaultBuildDelay = 600 * time.Millisecond

type configYaml struct {
	Root                 string
	TmpPath              string        `yaml:"tmpPath"`
	BuildName            string        `yaml:"buildName"`
	BuildLog             string        `yaml:"buildLog"`
	Verbosity            string        `yaml:"verbosity"`
	SmartRebuildQtpl     bool          `yaml:"smartRebuildQtpl"`
	ValidExtensions      []string      `yaml:"validExtensions"`
	InvalidExtensions    []string      `yaml:"invalidExtensions"`
	IgnoredFiles         []string      `yaml:"ignoredFiles"`
	IgnoredDirectories   []string      `yaml:"ignoredDirectories"`
	ExclusiveDirectories []string      `yaml:"exclusiveDirectories"`
	BuildDelay           time.Duration `yaml:"buildDelay"`
	Colorize             bool          `yaml:"colorize"`
	LogColors            logColorsYaml `yaml:"logColors"`
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
		Root:                 ".",
		TmpPath:              "./tmp",
		BuildName:            "espal-core",
		BuildLog:             "errors.log",
		Verbosity:            "normal",
		SmartRebuildQtpl:     true,
		ValidExtensions:      []string{"go", "qtpl", "js", "css"},
		InvalidExtensions:    []string{"tmp", "lock", "log", "yml", "json"},
		IgnoredDirectories:   []string{"tmp", "node_modules"},
		IgnoredFiles:         []string{"*_test.go"},
		ExclusiveDirectories: []string{},
		BuildDelay:           defaultBuildDelay,
		Colorize:             true,
		LogColors: logColorsYaml{
			Main:    "cyan",
			Build:   "yellow",
			Runner:  "green",
			Watcher: "magenta",
		},
	}
}

func (r *Runner) wildcardDirRootExists(path string) (string, bool, error) {
	rootDir := strings.TrimSuffix(path, "/**/*")
	stat, err := os.Stat(rootDir)
	if err != nil && !os.IsNotExist(err) {
		return "", false, errors.Trace(err)
	}
	return rootDir, !os.IsNotExist(err) && stat.IsDir(), nil
}

func (r *Runner) buildIgnoredDirectories() {
	r.ignoredDirectories = []string{}
	for _, dir := range r.config.IgnoredDirectories {
		dir := strings.TrimSpace(dir)
		if strings.Contains(dir, "*") { // nolint:nestif
			dirs, err := zglob.Glob(strings.TrimSpace(dir))
			if err != nil {
				r.runnerLog(err.Error())
			}
			if len(dirs) > 0 {
				for _, matchedDir := range dirs {
					stat, err := os.Stat(matchedDir)
					if err != nil {
						r.runnerLog(err.Error())
					}
					if stat.IsDir() {
						r.ignoredDirectories = append(r.ignoredDirectories, matchedDir)
					}
				}
			}
			rootDir, ok, err := r.wildcardDirRootExists(dir)
			if err != nil {
				r.runnerLog(err.Error())
			}
			if ok {
				r.ignoredDirectories = append(r.ignoredDirectories, rootDir)
			}
		} else {
			r.ignoredDirectories = append(r.ignoredDirectories, dir)
		}
	}
}

func (r *Runner) buildExclusiveDirectories() {
	r.exclusiveDirectories = []string{}
	for _, dir := range r.config.ExclusiveDirectories {
		dir := strings.TrimSpace(dir)
		if strings.Contains(dir, "*") { // nolint:nestif
			dirs, err := zglob.Glob(strings.TrimSpace(dir))
			if err != nil {
				r.runnerLog(err.Error())
			}
			if len(dirs) > 0 {
				for _, matchedDir := range dirs {
					stat, err := os.Stat(matchedDir)
					if err != nil {
						r.runnerLog(err.Error())
					}
					if stat.IsDir() {
						r.exclusiveDirectories = append(r.exclusiveDirectories, matchedDir)
					}
				}
			}
			rootDir, ok, err := r.wildcardDirRootExists(dir)
			if err != nil {
				r.runnerLog(err.Error())
			}
			if ok {
				r.exclusiveDirectories = append(r.exclusiveDirectories, rootDir)
			}
		} else {
			r.exclusiveDirectories = append(r.exclusiveDirectories, dir)
		}
	}
}

func (r *Runner) resolveVerbosity() error {
	supported := map[string]uint8{
		"verbose": verbosityVerbose,
		"normal":  verbosityNormal,
		"quiet":   verbosityQuiet,
		"silent":  verbositySilent,
	}
	var ok bool
	r.verbosity, ok = supported[r.config.Verbosity]
	if !ok {
		mapsKeys := []string{}
		for k := range supported {
			mapsKeys = append(mapsKeys, k)
		}
		return errors.Errorf("unsupported verbostiy `%s`. Supported are `%s`",
			r.config.Verbosity, strings.Join(mapsKeys, ", "))
	}
	return nil
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

	r.buildIgnoredDirectories()
	r.buildExclusiveDirectories()

	if err := r.resolveVerbosity(); err != nil {
		return errors.Trace(err)
	}

	r.config.TmpPath, err = filepath.Abs(r.config.TmpPath)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}
