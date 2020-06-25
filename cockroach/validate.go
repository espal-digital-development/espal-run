package cockroach

import "github.com/juju/errors"

func (c *Cockroach) validate() error {
	if c.desiredNodes < 1 {
		return errors.Errorf("desiredNodes has to be at least one. %d given", c.desiredNodes)
	}
	if c.host == "" {
		return errors.Errorf("host cannot be empty")
	}
	if c.rootUser == "" {
		return errors.Errorf("rootUser cannot be empty")
	}
	if err := c.isSafePortRange(c.portStart, "portStart"); err != nil {
		return errors.Trace(err)
	}
	if err := c.isSafePortRange(c.httpPortStart, "httpPortStart"); err != nil {
		return errors.Trace(err)
	}
	if c.httpHost == "" {
		return errors.Errorf("httpHost cannot be empty")
	}
	if c.httpUser == "" {
		return errors.Errorf("httpUser cannot be empty")
	}
	if c.httpPassword == "" {
		return errors.Errorf("httpPassword cannot be empty")
	}
	return nil
}
