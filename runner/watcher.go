package runner

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/howeyc/fsnotify"
	"github.com/juju/errors"
)

const windowsOS = "windows"

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

	var sumBytes []byte
	var err error

	if runtime.GOOS == windowsOS {
		sumBytes, err = exec.Command("certutil", "-hashfile", path, "MD5").CombinedOutput()
	} else {
		sumBytes, err = exec.Command("md5", path).CombinedOutput()
	}

	if err != nil {
		return false, errors.Trace(err)
	}

	var fileSum string
	if runtime.GOOS == windowsOS {
		fileSum = strings.Split(string(sumBytes), "\n")[1]
	} else {
		fileSum = strings.TrimSpace(strings.Split(string(sumBytes), " = ")[1])
	}

	if ok && sum == fileSum {
		return false, nil
	}
	r.checksumsMutex.Lock()
	r.fileChecksums[path] = fileSum
	r.checksumsMutex.Unlock()
	return true, nil
}

func (r *Runner) handleEvent(ev *fsnotify.FileEvent) bool {
	if ev.IsAttrib() {
		return false
	}
	isWatched, err := r.isWatchedFile(ev.Name)
	if err != nil {
		r.buildLog("watch check failed for %s: %s", ev.Name, err.Error())
		return false
	}
	if !isWatched {
		return false
	}

	if ev.IsModify() {
		isChanged, err := r.validateChecksum(ev.Name)
		if err != nil {
			r.buildLog("md5 checksum failed for %s: %s", ev.Name, err.Error())
		}
		if !isChanged {
			return false
		}
	}
	if ev.IsCreate() {
		stat, err := os.Stat(ev.Name)
		if err != nil {
			r.buildLog("filesize check failed for %s: %s", ev.Name, err.Error())
		} else if stat.Size() == 0 {
			// Add to the checksum buffer, so modify won't re-trigger if empty is re-saved
			_, err := r.validateChecksum(ev.Name)
			if err != nil {
				r.buildLog("md5 checksum failed for %s: %s", ev.Name, err.Error())
			}
			return false
		}
	}
	if ev.IsDelete() {
		r.checksumsMutex.RLock()
		_, ok := r.fileChecksums[ev.Name]
		r.checksumsMutex.RUnlock()
		if ok {
			r.checksumsMutex.Lock()
			delete(r.fileChecksums, ev.Name)
			r.checksumsMutex.Unlock()
		}
	}

	if r.config.SmartRebuildQtpl && strings.HasSuffix(ev.Name, ".qtpl") {
		ok, err := r.rebuildQtpl(ev.Name)
		if err != nil {
			r.watcherLog("QuickTemplate compilation failed: %s", err.Error())
			return false
		}
		if !ok {
			return false
		}
		if r.verbosity >= verbosityNormal {
			r.watcherLog("Recompiled %s", ev.Name)
		}
	}
	return true
}

func (r *Runner) watchFolder(path string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Trace(err)
	}

	go func(r *Runner) {
		for {
			select {
			case ev := <-watcher.Event:
				if propagate := r.handleEvent(ev); !propagate {
					continue
				}
				if r.verbosity >= verbosityQuiet {
					r.watcherLog("sending event %s", ev)
				}
				r.startChannel <- ev.String()
			case err := <-watcher.Error:
				r.watcherLog("error: %s", err)
			}
		}
	}(r)

	if r.verbosity >= verbosityVerbose {
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
		if err != nil {
			return errors.Trace(err)
		}
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
			if r.verbosity >= verbosityVerbose {
				r.watcherLog("Ignoring %s", path)
			}
			return filepath.SkipDir
		}

		return r.watchFolder(path)
	})
}
