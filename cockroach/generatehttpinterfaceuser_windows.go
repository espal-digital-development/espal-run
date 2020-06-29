// +build windows

package cockroach

// import "github.com/juju/errors"

// func (c *Cockroach) generateHTTPInterfaceUser() error {
// 	return errors.New("not implemented yet")
// }

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/juju/errors"
)

func (c *Cockroach) generateHTTPInterfaceUser() error {
	log.Println("Generating http interface user..")

	SQLString := fmt.Sprintf(httpUserSQL, c.httpUser, c.httpPassword, c.httpUser)

	out, err := exec.Command("cockroach", "sql", "--certs-dir="+c.certsDir, fmt.Sprintf("--host=%s:%d", c.host, c.portStart), `--execute=`+SQLString).CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return errors.Trace(err)
	}
	return nil
}
