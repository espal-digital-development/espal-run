// +build windows

package system

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"

	"github.com/juju/errors"
)

const notInstalledErrBlueprint = "%s is not installed. Please install it at %s"

func (s *System) checkOSSpecificTools() error {
	return nil
}

func (s *System) installPngQuantIfNeeded() error {
	out, _ := exec.Command("which", "pngquant").CombinedOutput()
	isInstalled := bytes.Contains(out, []byte("/pngquant"))

	if isInstalled {
		return nil
	}

	return errors.Errorf(notInstalledErrBlueprint, "pngquant", "https://pngquant.org")
}

func (s *System) installJpegoptimIfNeeded() error {
	return nil
}

func (s *System) installGifsicleIfNeeded() error {
	out, _ := exec.Command("which", "gifsicle").CombinedOutput()
	isInstalled := bytes.Contains(out, []byte("/gifsicle"))

	if isInstalled {
		return nil
	}

	return errors.Errorf(notInstalledErrBlueprint, "gifsicle", "https://www.lcdf.org/gifsicle/")
}

func (s *System) installSvgoIfNeeded() error {
	out, _ := exec.Command("which", "svgo").CombinedOutput()
	isInstalled := bytes.Contains(out, []byte("/svgo"))

	if isInstalled {
		return nil
	}

	log.Println("svgo is not installed. Attempting to install now..")
	out, err := exec.Command("npm", "install", "-g", "svgo").CombinedOutput()

	if err != nil {
		fmt.Println(string(out))
		return errors.Trace(err)
	}

	return errors.Errorf(notInstalledErrBlueprint, "svgo", "https://github.com/svg/svgo")
}
