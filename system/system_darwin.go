// +build darwin

package system

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"

	"github.com/juju/errors"
)

var errHomebrewNotInstalled = errors.New("Homebrew is required to install required tools. You can install it via " +
	"https://brew.sh")

func (s *System) checkOSSpecificTools() error {
	out, _ := exec.Command("which", "brew").CombinedOutput()
	isInstalled := bytes.Contains(out, []byte("/brew"))
	if !isInstalled {
		if err := s.installHomebrew(); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func (s *System) installHomebrew() error {
	return errors.Trace(errHomebrewNotInstalled)
}

func (s *System) installPngQuantIfNeeded() error {
	out, _ := exec.Command("which", "pngquant").CombinedOutput()
	isInstalled := bytes.Contains(out, []byte("/pngquant"))
	if isInstalled {
		return nil
	}
	log.Println("pngquant is not installed. Attempting to install now..")
	out, err := exec.Command("brew", "install", "pngquant").CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return errors.Trace(err)
	}
	return nil
}

func (s *System) installJpegoptimIfNeeded() error {
	out, _ := exec.Command("which", "jpegoptim").CombinedOutput()
	isInstalled := bytes.Contains(out, []byte("/jpegoptim"))
	if isInstalled {
		return nil
	}
	log.Println("jpegoptim is not installed. Attempting to install now..")
	out, err := exec.Command("brew", "install", "jpegoptim").CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return errors.Trace(err)
	}
	return nil
}

func (s *System) installGifsicleIfNeeded() error {
	out, _ := exec.Command("which", "gifsicle").CombinedOutput()
	isInstalled := bytes.Contains(out, []byte("/gifsicle"))
	if isInstalled {
		return nil
	}
	log.Println("gifsicle is not installed. Attempting to install now..")
	out, err := exec.Command("brew", "install", "gifsicle").CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return errors.Trace(err)
	}
	return nil
}

func (s *System) installSvgoIfNeeded() error {
	out, _ := exec.Command("which", "svgo").CombinedOutput()
	isInstalled := bytes.Contains(out, []byte("/svgo"))
	if isInstalled {
		return nil
	}
	log.Println("svgo is not installed. Attempting to install now..")
	out, err := exec.Command("brew", "install", "svgo").CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return errors.Trace(err)
	}
	return nil
}
