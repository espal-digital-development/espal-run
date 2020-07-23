// +build linux

package system

import (
	"bytes"
	"os/exec"

	"github.com/juju/errors"
)

// Federated List of Linux' and their fallback- and primary package manager
// Debian			dpkg    apt
// Ubuntu			dpkg    apt (snap, flatpack)
// Mint				dpkg    apt (flatpack)
// Zorin OS			dpkg    apt (snap, flatpak)
// Elementary OS	dpkg    apt (snap, flatpak)
// Raspberry Pi OS	dpkg    apt
// MX				dpkg	apt
// Peppermint		dpkg	apt
// Pop!_OS			dpkg	apt
// Deepin			dpkg	apt
// antiX			dpkg	apt
// Sparky			dpkg	apt
// Linux Lite		dpkg	apt
// Kali				dpkg    -
// Tails			dpkg	-
// Bodhi			dpkg	-
// openSUSE			rpm     zypper (yast)
// CentOS			rpm     yum
// RedHat			rpm     yum
// Fedora			rpm     dnf (Dandified yum)
// Arch				libalpm pacman
// Manjaro			libalpm pacman
// Gentoo			emerge  - (name=portage)
// Solus			eopkg   - (based on PiSi)
// Puppy			ppm		- (command? puppy package manager)

const packageNotInstalledErrBlueprint = "%s is not installed. Install it with your package manager via `%s`"

var errNoCompatiblePackageManagerFound = errors.New("no compatible package manager found for this linux distro")

type packageManager struct {
	name        string
	installArgs string
}

func (m *packageManager) installPackageCmd(name string) string {
	installArgs := m.installArgs
	if installArgs == "" {
		installArgs = "install"
	}
	return m.name + " " + installArgs + " " + name
}

func (s *System) checkOSSpecificTools() error {
	_, err := s.determinePackageManager()
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (s *System) determinePackageManager() (*packageManager, error) {
	var pkgManager *packageManager

	// Search order
	search := []*packageManager{
		&packageManager{
			name: "sudo apt",
		},
		// TODO :: Figure out how to download and install via tmpDir
		// &packageManager{ // This one needs downloaded .deb files
		// 	Name:        "dpkg",
		// 	InstallArgs: "--install",
		// },
		&packageManager{
			name: "yum",
		},
		&packageManager{
			name: "dnf",
		},
		&packageManager{
			name: "zypper",
		},
		&packageManager{
			name: "rpm",
		},
		&packageManager{
			name:        "pacman",
			installArgs: "-S",
		},
		// TODO :: Figure out how to implement
		// &packageManager{
		// 	Name:        "libalpm",
		// 	InstallArgs: "???",
		// },
		&packageManager{
			name:        "emerge",
			installArgs: "", // uses no extra parameters, just emerge <package name>
		},
		&packageManager{
			name: "eopkg",
		},
	}

	for k := range search {
		out, _ := exec.Command("which", search[k].name).CombinedOutput()
		if bytes.Contains(out, []byte("/"+search[k].name)) {
			pkgManager = search[k]
			break
		}
	}

	if pkgManager == nil {
		return nil, errors.Trace(errNoCompatiblePackageManagerFound)
	}

	return pkgManager, nil
}

func (s *System) installPngQuantIfNeeded() error {
	packageManager, err := s.determinePackageManager()
	if err != nil {
		return errors.Trace(err)
	}
	return errors.Errorf(packageNotInstalledErrBlueprint, "pngquant", packageManager.installPackageCmd("pngquant"))
}

func (s *System) installJpegoptimIfNeeded() error {
	packageManager, err := s.determinePackageManager()
	if err != nil {
		return errors.Trace(err)
	}
	return errors.Errorf(packageNotInstalledErrBlueprint, "jpegoptim", packageManager.installPackageCmd("jpegoptim"))
}

func (s *System) installGifsicleIfNeeded() error {
	packageManager, err := s.determinePackageManager()
	if err != nil {
		return errors.Trace(err)
	}
	return errors.Errorf(packageNotInstalledErrBlueprint, "gifsicle", packageManager.installPackageCmd("gifsicle"))
}

func (s *System) installSvgoIfNeeded() error {
	packageManager, err := s.determinePackageManager()
	if err != nil {
		return errors.Trace(err)
	}
	return errors.Errorf(packageNotInstalledErrBlueprint, "svgo", packageManager.installPackageCmd("svgo"))
}
