// +build darwin

package cockroach

import (
	"bytes"
	"log"
	"os/exec"

	"github.com/juju/errors"
)

func (c *Cockroach) checkInstall() error {
	out, _ := exec.Command("which", "cockroach").CombinedOutput()
	isInstalled := bytes.Contains(out, []byte("/cockroach"))
	if !isInstalled {
		log.Println(cockroachNotFoundInstalling)
		out, err := exec.Command("brew", "install", "cockroach").CombinedOutput()
		if err != nil {
			log.Println(string(out))
			return errors.Trace(err)
		}
	}
	return nil
}
