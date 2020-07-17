// +build windows

package system

import "github.com/pkg/errors"

const notInstalledErrBlueprint = "%s is not installed. Please install it at %s"

func (s *System) checkOSSpecificTools() error {
	return nil
}

func (s *System) installPngQuantIfNeeded() error {
	return errors.Errorf(notInstalledErrBlueprint, "pngquant", "https://pngquant.org")
}

func (s *System) installJpegoptimIfNeeded() error {
	return nil
}

func (s *System) installGifsicleIfNeeded() error {
	return errors.Errorf(notInstalledErrBlueprint, "gifsicle", "https://www.lcdf.org/gifsicle/")
}

func (s *System) installSvgoIfNeeded() error {
	return errors.Errorf(notInstalledErrBlueprint, "svgo", "https://github.com/svg/svgo")
}
