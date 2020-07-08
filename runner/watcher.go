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
				// Ignore the MODIFY|ATTRIBUTE combination
				if ev.IsModify() && ev.IsAttrib() {
					continue
				}

				isWatched, err := r.isWatchedFile(ev.Name)
				if err != nil {
					r.fatal(err)
				}
				if !isWatched {
					continue
				}

				// TODO :: Make a buffer and check if the file was really changed (also clean the tmp buffer every boot)
				// if ev.IsModify() {
				// }

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
	return nil
}

func (r *Runner) watch() error {
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
