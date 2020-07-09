package runner

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/howeyc/fsnotify"
	"github.com/juju/errors"
)

func (r *Runner) rebuildQtpl(path string) (bool, error) {
	cmd := exec.Command("qtc", "-file", path)
	out, err := cmd.CombinedOutput()
	if err != nil || bytes.Contains(out, []byte("error when parsing file")) {
		r.watcherLog("Building QuickTemplate `%s` failed:\n%s", path, out)
		return false, errors.Trace(err)
	}
	return true, nil
}

// returned bool indicates if the file at the given path has changed.
func (r *Runner) validateChecksum(path string) (bool, error) {
	r.checksumsMutex.RLock()
	sum, ok := r.fileChecksums[path]
	r.checksumsMutex.RUnlock()
	sumBytes, err := exec.Command("md5", path).CombinedOutput()
	if err != nil {
		return false, errors.Trace(err)
	}
	fileSum := strings.Trim(strings.Split(string(sumBytes), " = ")[1], "\n")
	if ok && sum == fileSum {
		return false, nil
	}
	r.checksumsMutex.Lock()
	r.fileChecksums[path] = fileSum
	r.checksumsMutex.Unlock()
	return true, nil
}

// nolint:gocognit
func (r *Runner) watchFolder(path string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Trace(err)
	}

	go func(r *Runner) {
		for {
			select {
			case ev := <-watcher.Event:
				if ev.IsAttrib() {
					continue
				}

				if ev.IsModify() {
					isChanged, err := r.validateChecksum(ev.Name)
					if err != nil {
						r.fatal(err)
					}
					if !isChanged {
						continue
					}
				}

				isWatched, err := r.isWatchedFile(ev.Name)
				if err != nil {
					r.fatal(err)
				}
				if !isWatched {
					continue
				}

				if r.config.SmartRebuildQtpl && strings.HasSuffix(ev.Name, ".qtpl") {
					ok, err := r.rebuildQtpl(ev.Name)
					if err != nil {
						r.fatal(err)
					}
					if !ok {
						continue
					}
					r.watcherLog("Rebuild %s", ev.Name)
				}

				r.watcherLog("sending event %s", ev)
				r.startChannel <- ev.String()
			case err := <-watcher.Error:
				r.watcherLog("error: %s", err)
			}
		}
	}(r)

	if r.config.VerboseWatching {
		r.watcherLog("Watching %s", path)
	}

	if err := watcher.Watch(path); err != nil {
		return errors.Trace(err)
	}
	r.totalWatchedFolders++
	return nil
}

func (r *Runner) watch() error {
	if len(r.exclusiveDirectories) > 0 {
		for _, dir := range r.exclusiveDirectories {
			if r.isIgnoredFolder(dir) {
				continue
			}
			if err := r.watchFolder(dir); err != nil {
				return errors.Trace(err)
			}
		}
		return nil
	}

	return filepath.Walk(r.config.Root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		isTmpDir, tmpDirErr := r.isTmpDir(path)
		if tmpDirErr != nil {
			return errors.Trace(tmpDirErr)
		}
		if isTmpDir {
			return filepath.SkipDir
		}

		if len(path) > 1 && strings.HasPrefix(filepath.Base(path), ".") {
			return filepath.SkipDir
		}

		if r.isIgnoredFolder(path) {
			if r.config.VerboseWatching {
				r.watcherLog("Ignoring %s", path)
			}
			return filepath.SkipDir
		}

		return r.watchFolder(path)
	})
}
