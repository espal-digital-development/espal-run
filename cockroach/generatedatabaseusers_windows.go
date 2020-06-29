// +build windows

package cockroach

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/juju/errors"
)

func (c *Cockroach) generateDatabaseUsers() error {
	log.Println("Generating database, users, roles and assigning privileges..")

	out, err := exec.Command("cockroach", "sql", "--certs-dir="+c.certsDir, fmt.Sprintf("--host=%s:%d", c.host, c.portStart), `--execute=`+setupDatabaseSQL).CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return errors.Trace(err)
	}
	return nil
}
