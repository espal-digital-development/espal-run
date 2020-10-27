// +build linux

package mkcert

import (
	"bytes"
	"os/exec"

	"github.com/juju/errors"
)

var errNotImplemented = errors.New("not implemented yet")

func (m *Mkcert) checkAndInstall() error {
	out, _ := exec.Command("which", "mkcert").CombinedOutput()
	if bytes.Contains(out, []byte("/mkcert")) {
		return nil
	}
	return errors.Trace(errNotImplemented)
}
