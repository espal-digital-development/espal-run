// +build linux

package openssl

import (
	"bytes"
	"os/exec"

	"github.com/juju/errors"
)

var errNotImplemented = errors.New("not implemented yet")

func (o *OpenSSL) checkAndInstall() error {
	out, _ := exec.Command("which", "openssl").CombinedOutput()
	if bytes.Contains(out, []byte("/openssl")) {
		return nil
	}
	return errors.Trace(errNotImplemented)
}
