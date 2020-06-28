package sslgenerator

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/juju/errors"
)

// TODO :: This won't be needed for online versions with LetsEncrypt or external certificates

var (
	errServerPathNotSet = errors.New("serverPath is not set")
)

type SSLGenerator struct {
	serverPath string
}

// GetServerPath gets serverPath.
func (g *SSLGenerator) GetServerPath() string {
	return g.serverPath
}

// SetServerPath sets serverPath.
func (g *SSLGenerator) SetServerPath(serverPath string) {
	g.serverPath = filepath.FromSlash(serverPath)
}

func (g *SSLGenerator) Do() error {
	if g.serverPath == "" {
		return errors.Trace(errServerPathNotSet)
	}

	var keyFileFound bool
	var crtFileFound bool
	_, err := os.Stat(g.serverPath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Trace(err)
	}
	if os.IsNotExist(err) {
		if err := os.MkdirAll(g.serverPath, 0700); err != nil {
			return errors.Trace(err)
		}
	}

	files, err := ioutil.ReadDir(g.serverPath)
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

	log.Println("No server certificate found. Creating one now. This might take a while..")

	// TODO :: Use makecert here instead? Or try both with makecert as a first option?

	openSSLCreation := []string{
		`req`, `-new`, `-newkey`, `rsa:4096`, `-days`, `365`, `-nodes`, `-x509`,
		`-subj`, `/C=US/ST=State/L=Town/O=Office/CN=localhost`,
		`-keyout`, `app/server/localhost.key`,
		`-out`, `app/server/localhost.crt`,
	}
	out, err := exec.Command("openssl", openSSLCreation...).CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return errors.Trace(err)
	}

	return nil
}

// New returns a new instance of SSLGenerator.
func New() (*SSLGenerator, error) {
	g := &SSLGenerator{}
	return g, nil
}
