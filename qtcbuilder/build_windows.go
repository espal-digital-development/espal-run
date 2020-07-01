// +build windows

package qtcbuilder

import (
	"os"
	"os/exec"

	"github.com/juju/errors"
)

func (b *QTCBuilder) build() ([]byte, error) {
	_, err := os.Stat("./pages")
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.Trace(err)
	}
	if os.IsNotExist(err) {
		return nil, nil
	}
	out, err := exec.Command("cd", "pages", "-AND", "qtc").CombinedOutput()
	return out, errors.Trace(err)
}
