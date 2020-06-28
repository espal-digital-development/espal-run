// +build windows

package cockroach

// import "github.com/juju/errors"

// func (c *Cockroach) generateHTTPInterfaceUser() error {
// 	return errors.New("not implemented yet")
// }

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

	SQLString := fmt.Sprintf(httpUserSQL, c.httpUser, c.httpPassword, c.httpUser)

	tmpPSFile := filepath.FromSlash(os.TempDir() + "/tmp.ps1")
	if err := ioutil.WriteFile(tmpPSFile,
		[]byte(fmt.Sprintf(`cockroach sql --certs-dir=%s --host=%s:%d --execute="%s"`,
			c.certsDir, c.host, c.portStart, SQLString)), 0600); err != nil {
		return errors.Trace(err)
	}
	out, err := exec.Command("powershell", tmpPSFile).CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return errors.Trace(err)
	}
	return nil
}
