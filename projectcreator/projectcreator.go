package projectcreator

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/juju/errors"
)

type ProjectCreator struct {
}

func (c *ProjectCreator) Do(path string) error {
	stat, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return errors.Trace(err)
	}
	switch {
	case os.IsNotExist(err):
		if err := os.MkdirAll(path, 0700); err != nil {
			return errors.Trace(err)
		}
	case stat.IsDir():
		return errors.Errorf("%s already exists", path)
	default:
		return errors.Errorf("%s already exists, and is not a directory", path)
	}

	if err := ioutil.WriteFile(filepath.FromSlash(path+"/.gitignore"), gitIgnoreFile, 0600); err != nil {
		return errors.Trace(err)
	}
	if err := ioutil.WriteFile(filepath.FromSlash(path+"/espal-run.yml"), espalRunFile, 0600); err != nil {
		return errors.Trace(err)
	}
	if err := ioutil.WriteFile(filepath.FromSlash(path+"/main.go"), mainGoFile, 0600); err != nil {
		return errors.Trace(err)
	}
	if err := ioutil.WriteFile(filepath.FromSlash(path+"/main_test.go"), mainGoTestFile, 0600); err != nil {
		return errors.Trace(err)
	}

	if err := os.Chdir(path); err != nil {
		return errors.Trace(err)
	}

	out, err := exec.Command("go", "mod", "init", filepath.Base(path)).CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return errors.Trace(err)
	}
	out, err = exec.Command("go", "mod", "tidy").CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return errors.Trace(err)
	}

	return nil
}

// New returns a new instance of ProjectCreator.
func New() (*ProjectCreator, error) {
	c := &ProjectCreator{}
	return c, nil
}
