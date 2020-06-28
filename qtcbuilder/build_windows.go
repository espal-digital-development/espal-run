// +build windows

package qtcbuilder

import (
	"os/exec"

	"github.com/juju/errors"
)

func (b *QTCBuilder) build() ([]byte, error) {
	out, err := exec.Command("cd", "pages", "-AND", "qtc").CombinedOutput()
	return out, errors.Trace(err)
}
