// +build !windows

package main

import (
	"syscall"

	"github.com/juju/errors"
)

func setSoftUlimit() error {
	var rLimit syscall.Rlimit
	rLimit.Max = 20000
	rLimit.Cur = 20000
	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}
