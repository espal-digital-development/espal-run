package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/juju/errors"
)

func generateTLSFiles() error {
	// TODO :: This won't be needed for online versions with LetsEncrypt or external certificates
	// TODO :: This is really unstable. Files can be named anything,
	// thus this fully relies on files being .key and .crt/.cert.
	var keyFileFound bool
	var crtFileFound bool
	serverPath := filepath.FromSlash("app/server")
	_, err := os.Stat(serverPath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Trace(err)
	}
	if os.IsNotExist(err) {
		if err := os.Mkdir(serverPath, 0700); err != nil {
			return errors.Trace(err)
		}
	}

	files, err := ioutil.ReadDir(serverPath)
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

	if runtime.GOOS == darwinOS || runtime.GOOS == linuxOS {
		out, _ := exec.Command("which", "openssl").CombinedOutput()
		if !bytes.Contains(out, []byte("/openssl")) {
			if runtime.GOOS == darwinOS {
				if !homebrewInstalled {
					return errors.Errorf("OpenSSL is not installed and can't access Homebrew. Please install manually")
				}
				log.Println("OpenSSL not installed. Installing through Homebrew..")
				out, err := exec.Command("brew", "install", "openssl").CombinedOutput()
				if err != nil {
					log.Println(string(out))
					return errors.Trace(err)
				}
			} else if runtime.GOOS == linuxOS {
				if !linuxbrewInstalled {
					log.Fatal("OpenSSL is not installed and can't access Linuxbrew. Please install manually")
				}
				log.Println("OpenSSL not installed. Installing through Linuxbrew..")
				out, err := exec.Command("brew", "install", "openssl").CombinedOutput()
				if err != nil {
					log.Println(string(out))
					return errors.Trace(err)
				}
			}
		}
	} else if runtime.GOOS == windowsOS {
		out, _ := exec.Command("which", "openssl").CombinedOutput()
		if !bytes.Contains(out, []byte("/openssl")) {
			return errors.Errorf("OpenSSL is not installed. Please install Git or OpenSSL directly")
		}
	}

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
