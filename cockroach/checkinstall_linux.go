// +build linux

package cockroach

import (
	"bytes"
	"log"
	"os"
	"os/exec"

	"github.com/juju/errors"
)

const cockroachNotFoundInstalling = "Did not find `cockroach`. Attempting to install.."

func (c *Cockroach) checkInstall() error {
	out, _ := exec.Command("which", "cockroach").CombinedOutput()
	isInstalled := bytes.Contains(out, []byte("/cockroach"))
	if isInstalled {
		return nil
	}
	log.Println(cockroachNotFoundInstalling)
	tmpDir := os.TempDir()
	tarFileName := tmpDir + "/cockroach.tgz"
	if err := c.downloadFile(tarFileName,
		"https://binaries.cockroachdb.com/cockroach-v20.1.3.linux-amd64.tgz"); err != nil {
		return errors.Trace(err)
	}
	out, err := exec.Command("tar", "zxvf", tarFileName).CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return errors.Trace(err)
	}
	out, err = exec.Command("cp", "cockroach-v20.1.3.linux-amd64/cockroach", os.Getenv("GOPATH")+"/bin/").
		CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return errors.Trace(err)
	}
	return nil
}
