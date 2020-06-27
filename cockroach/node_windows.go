// +build windows

package cockroach

import (
	"github.com/juju/errors"
)

func (c *Cockroach) startNodeNonBlocking(storeName string, portsNumber int, httpPortsNumber int) error {
	// TODO :: WINDOWS :: Continue from here (startNode function above)
	return errors.New("not implemented yet")
}
