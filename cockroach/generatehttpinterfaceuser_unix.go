// +build unix

package cockroach

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/juju/errors"
)

func (c *Cockroach) generateHTTPInterfaceUser() error {
	log.Println("Generating http interface user..")
	tmpSQLFile := filepath.FromSlash(os.TempDir() + "/tmp.sql")
	if err := ioutil.WriteFile(tmpSQLFile,
		[]byte(fmt.Sprintf(httpUserSQL, c.httpUser, c.httpPassword, c.httpUser)), 0700); err != nil {
		return errors.Trace(err)
	}
	tmpSHFile := filepath.FromSlash(os.TempDir() + "/tmp.sh")
	if err := ioutil.WriteFile(tmpSHFile,
		[]byte(fmt.Sprintf("#!/bin/sh\n\n"+`cockroach sql --certs-dir=%s --host=%s:%d < %s`,
			c.certsDir, c.host, c.portStart, tmpSQLFile)), 0700); err != nil {
		return errors.Trace(err)
	}
	out, err := exec.Command("/bin/sh", tmpSHFile).CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return errors.Trace(err)
	}
	return nil
}
