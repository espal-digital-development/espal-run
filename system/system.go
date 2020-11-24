package system

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

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
		log.Println(err.Error())
		log.Println("You can choose to continue without the base tools.")
		shouldContinue, err := s.continueOrAbort()
		if err != nil {
			return errors.Trace(err)
		}
		if shouldContinue {
			return nil
		}
		os.Exit(2) // nolint:gomnd
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
	if hasErrors {
		errMessage := "Core dependency(/ies) are missing and couldn't be automatically installed"
		log.Println(errMessage + ":")
		for k := range collectErrors {
			if collectErrors[k] == nil {
				continue
			}
			log.Println("\t", collectErrors[k].Error())
		}

		log.Println("You can choose to continue without the dependencies.")
		shouldContinue, err := s.continueOrAbort()
		if err != nil {
			return errors.Trace(err)
		}
		if !shouldContinue {
			os.Exit(2) // nolint:gomnd
		}
	}

	return nil
}

func (s *System) continueOrAbort() (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Do you want to continue or abort? [c/A] ")
		value, err := reader.ReadString('\n')
		if err != nil && err == io.EOF {
			break
		}
		if err != nil {
			return false, errors.Trace(err)
		}
		value = strings.TrimSpace(value)
		if value == "" || value == "A" {
			return false, nil
		} else if value == "c" {
			break
		}
	}
	return true, nil
}

// New returns a new instance of System.
func New() (*System, error) {
	s := &System{}
	return s, nil
}
