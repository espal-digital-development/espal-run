package projectcreator

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/juju/errors"
)

type ProjectCreator struct {
	devMode bool
}

// nolint:funlen
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

	if err := ioutil.WriteFile(path+"/.gitignore", gitIgnoreFile, 0600); err != nil {
		return errors.Trace(err)
	}
	if err := ioutil.WriteFile(path+"/espal-run.yml", runFile, 0600); err != nil {
		return errors.Trace(err)
	}
	if err := ioutil.WriteFile(path+"/main.go", mainGoFile, 0600); err != nil {
		return errors.Trace(err)
	}
	if err := ioutil.WriteFile(path+"/main_test.go", mainGoTestFile, 0600); err != nil {
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

	// TODO :: Need to find wherever the espal-core and espal-module-core are
	if c.devMode {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return errors.Trace(err)
		}
		modFilePath := absPath + "/go.mod"
		modFile, err := ioutil.ReadFile(modFilePath)
		if err != nil {
			return errors.Trace(err)
		}
		modFile = bytes.Replace(modFile, []byte("require ("), []byte("replace (\n"+
			"	github.com/espal-digital-development/espal-core => ../espal-core\n"+
			"	github.com/espal-digital-development/espal-module-core => ../espal-module-core\n"+
			")\n\n"+
			"require ("), 1)

		if err := ioutil.WriteFile(modFilePath, modFile, 0600); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

// New returns a new instance of ProjectCreator.
func New(devMode bool) (*ProjectCreator, error) {
	c := &ProjectCreator{
		devMode: devMode,
	}
	return c, nil
}
