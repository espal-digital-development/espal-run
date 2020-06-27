// +build windows

package cockroach

import (
	"github.com/juju/errors"
)

func (c *Cockroach) startNodeNonBlocking(storeName string, portsNumber int, httpPortsNumber int) error {
	return errors.Trace(errNotImplementedYet)
}
