package mkcert

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/juju/errors"
)

var (
	errServerPathNotSet = errors.New("serverPath is not set")
)

type Mkcert struct {
	serverPath string
}

// GetServerPath gets serverPath.
func (m *Mkcert) GetServerPath() string {
	return m.serverPath
}

// SetServerPath sets serverPath.
func (m *Mkcert) SetServerPath(serverPath string) {
	m.serverPath = serverPath
}

// Check and install the required binaries.
func (m *Mkcert) CheckAndInstall() error {
	if m.serverPath == "" {
		return errors.Trace(errServerPathNotSet)
	}

	var keyFileFound bool
	var crtFileFound bool
	_, err := os.Stat(m.serverPath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Trace(err)
	}
	if os.IsNotExist(err) {
		if err := os.MkdirAll(m.serverPath, 0700); err != nil {
			return errors.Trace(err)
		}
	}

	files, err := ioutil.ReadDir(m.serverPath)
	if err != nil {
		return errors.Trace(err)
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".key") {
			keyFileFound = true
		} else if strings.HasSuffix(f.Name(), ".crt") || strings.HasSuffix(f.Name(), ".cert") {
			crtFileFound = true
		}
	}

	// If a key/cert file combination is found, just stop. Naive, but ok for now
	if keyFileFound && crtFileFound {
		return nil
	}

	log.Println("No server certificate found. Creating one now..")

	if err := m.checkAndInstall(); err != nil {
		return errors.Trace(err)
	}
	return nil
}

// New returns a new instance of Mkcert.
func New() (*Mkcert, error) {
	m := &Mkcert{}
	return m, nil
}
