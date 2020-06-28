// +build windows

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

	tmpPSFile := filepath.FromSlash(os.TempDir() + "/tmp.ps1")
	if err := ioutil.WriteFile(tmpPSFile,
		[]byte(fmt.Sprintf(`cockroach sql --certs-dir=%s --host=%s:%d --execute="%s"`,
			c.certsDir, c.host, c.portStart, setupDatabaseSQL)), 0600); err != nil {
		return errors.Trace(err)
	}
	out, err := exec.Command("powershell", tmpPSFile).CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return errors.Trace(err)
	}
	return nil
}
