// +build darwin

package openssl

import (
	"bytes"
	"log"
	"os/exec"

	"github.com/juju/errors"
)

func (o *OpenSSL) checkAndInstall() error {
	out, _ := exec.Command("which", "openssl").CombinedOutput()
	if bytes.Contains(out, []byte("/openssl")) {
		return nil
	}
	out, _ = exec.Command("which", "brew").CombinedOutput()
	if !bytes.Contains(out, []byte("/brew")) {
		return errors.Errorf("OpenSSL is not installed and can't access Homebrew. Please install manually")
	}
	log.Println("OpenSSL not installed. Installing through Homebrew..")
	out, err := exec.Command("brew", "install", "openssl").CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return errors.Trace(err)
	}
	return nil
}
