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

func (c *Cockroach) generateDatabaseUsers() error {
	log.Println("Generating database, users, roles and assigning privileges..")
	tmpSQLFile := filepath.FromSlash(os.TempDir() + "/tmp.sql")
	if err := ioutil.WriteFile(tmpSQLFile, []byte(setupDatabaseSQL), 0600); err != nil {
		return errors.Trace(err)
	}
	tmpSHFile := filepath.FromSlash(os.TempDir() + "/tmp.sh")
	if err := ioutil.WriteFile(tmpSHFile,
		[]byte(fmt.Sprintf("#!/bin/sh\n\n"+`cockroach sql --certs-dir=%s --host=%s:%d < %s`,
			c.certsDir, c.host, c.portStart, tmpSQLFile)), 0600); err != nil {
		return errors.Trace(err)
	}
	out, err := exec.Command("/bin/sh", tmpSHFile).CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return errors.Trace(err)
	}
	return nil
}
