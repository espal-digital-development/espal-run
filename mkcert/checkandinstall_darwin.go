// +build darwin

package mkcert

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"

	"github.com/juju/errors"
)

// TODO :: Make this mutli-OS compatible

func (m *Mkcert) checkAndInstall() error {
	out, _ := exec.Command("which", "mkcert").CombinedOutput()
	if !bytes.Contains(out, []byte("/mkcert")) {
		out, _ = exec.Command("which", "brew").CombinedOutput()
		if !bytes.Contains(out, []byte("/brew")) {
			return errors.Errorf("Mkcert is not installed and can't access Homebrew. Please install manually")
		}
		log.Println("Mkcert not installed. Installing through Homebrew..")
		out, err := exec.Command("brew", "install", "mkcert", "nss").CombinedOutput()
		if err != nil {
			log.Println(string(out))
			return errors.Trace(err)
		}
	}
	log.Println("Initializing mkcert..")
	out, err := exec.Command("mkcert", "-install").CombinedOutput()
	if len(bytes.TrimSpace(out)) > 0 {
		fmt.Println(string(out))
	}
	if err != nil {
		return errors.Trace(err)
	}
	// TODO :: Should be able to know which domains we want based on the config and/or fixture data
	out, err = exec.Command("mkcert",
		"-cert-file", m.serverPath+"/localhost.crt", "-key-file", m.serverPath+"/localhost.key",
		"*.espal.loc", "espal.loc", "localhost", "127.0.0.1", "::1").CombinedOutput()
	if len(bytes.TrimSpace(out)) > 0 {
		fmt.Println(string(out))
	}
	if err != nil {
		return errors.Trace(err)
	}

	hosts, err := ioutil.ReadFile("/etc/hosts")
	if err != nil && !os.IsNotExist(err) {
		return errors.Trace(err)
	}
	pleaseAddMessage := "Please add `127.0.0.1 www.espal.loc` and `127.0.0.1 espal.loc` to"
	if err != nil && os.IsNotExist(err) {
		log.Printf("Can't load /etc/hosts to check the domain mapping. %s where yours is located", pleaseAddMessage)
	} else {
		reHost, err := regexp.Compile(`(?m)^\s*127\.0\.0\.1\s+(?:\w+\.)?espal\.loc\s*$`)
		if err != nil {
			return errors.Trace(err)
		}
		if !reHost.Match(hosts) {
			log.Printf("%s your /etc/hosts file (you probably need sudo privileges)", pleaseAddMessage)
		}
	}

	return nil
}
