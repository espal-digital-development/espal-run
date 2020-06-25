package cockroach

import (
	"bytes"
	"log"
	"os/exec"

	"github.com/juju/errors"
)

func (c *Cockroach) checkInstall() error {
	if c.isUnixOS() {
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
	} else if c.isWinOS() {
		// TODO :: Needs a Windows variance
		return errors.Errorf("No Windows auto-installation detection implemented yet..")
	}
	return nil
}
