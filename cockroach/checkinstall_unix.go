// +build unix

package cockroach

import (
	"github.com/juju/errors"
)

var errNotImplementedYet = errors.New("not implemented yet")

func (c *Cockroach) checkInstall() error {
	return errors.Wrap(errNotImplementedYet)
}
