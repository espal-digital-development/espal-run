package system

import (
	"log"

	"github.com/juju/errors"
)

// System provides tools to communicate with the core/OS system.
type System struct {
}

// SetSoftUlimit sets the OS's native soft max and cur limit.
func (s *System) SetSoftUlimit(max uint64, cur uint64) error {
	return s.setSoftUlimit(max, cur)
}

// InstallBinaryDependencies checks if binaries that the system is dependent on are installed and will do so if needed.
func (s *System) InstallBinaryDependencies() error {
	if err := s.checkOSSpecificTools(); err != nil {
		return errors.Trace(err)
	}
	collectErrors := []error{
		errors.Trace(s.installPngQuantIfNeeded()),
		errors.Trace(s.installJpegoptimIfNeeded()),
		errors.Trace(s.installGifsicleIfNeeded()),
		errors.Trace(s.installSvgoIfNeeded()),
	}

	var hasErrors bool
	for k := range collectErrors {
		if collectErrors[k] != nil {
			hasErrors = true
			break
		}
	}
	if !hasErrors {
		return nil
	}

	errMessage := "Core dependency(/ies) are missing and couldn't be automatically installed"
	log.Println(errMessage + ":")
	for k := range collectErrors {
		if collectErrors[k] == nil {
			continue
		}
		log.Println("\t", collectErrors[k].Error())
	}

	// TODO :: When all real installation attempts are ready, make this an error again

	// return errors.New(errMessage)
	return nil
}

// New returns a new instance of System.
func New() (*System, error) {
	s := &System{}
	return s, nil
}
