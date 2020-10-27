package openssl

import (
	"github.com/juju/errors"
)

type OpenSSL struct {
}

// Check and install the required binaries.
func (o *OpenSSL) CheckAndInstall() error {
	if err := o.checkAndInstall(); err != nil {
		return errors.Trace(err)
	}
	return nil
}

// New returns a new instance of OpenSSL.
func New() (*OpenSSL, error) {
	o := &OpenSSL{}
	return o, nil
}
