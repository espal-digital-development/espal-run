package runner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/juju/errors"
	"github.com/mattn/go-zglob"
)

func (r *Runner) initFolders() error {
	r.runnerLog("InitFolders")
	_, err := os.Stat(r.config.TmpPath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Trace(err)
	}
	if os.IsNotExist(err) {
		r.runnerLog("mkdir %s", r.config.TmpPath)
		return errors.Trace(os.Mkdir(r.config.TmpPath, 0700))
	}
	r.runnerLog("tmp dir already exists")
	return nil
}

func (r *Runner) isTmpDir(path string) (bool, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return false, errors.Trace(err)
	}

	return absolutePath == r.config.TmpPath, nil
}

func (r *Runner) isIgnoredFolder(path string) bool {
	for _, dir := range r.ignoredDirectories {
		if dir == path {
			return true
		}
	}
	return false
}

func (r *Runner) isWatchedFile(path string) (bool, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return false, errors.Trace(err)
	}

	if strings.HasPrefix(absolutePath, r.config.TmpPath) {
		return false, nil
	}

	ext := filepath.Ext(path)
	for k := range r.config.InvalidExtensions {
		if r.config.InvalidExtensions[k] == ext {
			return false, nil
		}
	}

	if len(r.config.IgnoredFiles) > 0 { // nolint:nestif
		fileName := filepath.Base(path)
		for k := range r.config.IgnoredFiles {
			if strings.Contains(r.config.IgnoredFiles[k], "*") {
				ok, err := zglob.Match(r.config.IgnoredFiles[k], fileName)
				if err != nil {
					return false, errors.Trace(err)
				}
				if ok {
					return false, nil
				}
			} else if r.config.IgnoredFiles[k] == fileName {
				return false, nil
			}
		}
	}

	for k := range r.config.ValidExtensions {
		if r.config.ValidExtensions[k] == ext {
			return true, nil
		}
	}

	return false, nil
}

func (r *Runner) shouldRebuild(eventName string) bool {
	for _, e := range r.config.InvalidExtensions {
		fileName := strings.Replace(strings.Split(eventName, ":")[0], `"`, "", -1)
		if strings.HasSuffix(fileName, e) {
			return false
		}
	}
	return true
}
