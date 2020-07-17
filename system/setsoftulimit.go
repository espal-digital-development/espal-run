// +build !windows

package system

import (
	"syscall"

	"github.com/juju/errors"
)

func (s *System) setSoftUlimit(max uint64, cur uint64) error {
	var rLimit syscall.Rlimit
	rLimit.Max = max
	rLimit.Cur = cur
	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}
