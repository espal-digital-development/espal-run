// +build linux

package system

// Federated List of Linux' and their fallback and primary package manager
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

type pkg struct {
	name         string // Global shared name
	managersName string // Name for the specific package manager
}

type packageManager struct {
	name              string
	installArgs       string
	availablePackages []*pkg
}

func (m *packageManager) installPackageCmd(name string) string {
	installArgs := m.installArgs
	if installArgs == "" {
		installArgs = "install"
	}
	return m.name + " " + installArgs + " " + name
}

func (s *System) checkOSSpecificTools() error {
	return nil
}

func (s *System) determinePackageManager() (string, error) {
	name := ""

	// Search order
	search := []*packageManager{
		&packageManager{
			Name: "apt",
		},
		// TODO :: Figure out how to download and install via tmpDir
		// &packageManager{ // This one needs downloaded .deb files
		// 	Name:        "dpkg",
		// 	InstallArgs: "--install",
		// },
		&packageManager{
			Name: "yum",
		},
		&packageManager{
			Name: "dnf",
		},
		&packageManager{
			Name: "zypper",
		},
		&packageManager{
			Name: "rpm",
		},
		&packageManager{
			Name:        "pacman",
			InstallArgs: "-S",
		},
		// TODO :: Figure out how to implement
		// &packageManager{
		// 	Name:        "libalpm",
		// 	InstallArgs: "???",
		// },
		&packageManager{
			Name:        "emerge",
			InstallArgs: "", // uses no extra parameters, just emerge <package name>
		},
		&packageManager{
			Name: "eopkg",
		},
	}

	return name, nil
}

func (s *System) installPngQuantIfNeeded() error {
	return nil
}

func (s *System) installJpegoptimIfNeeded() error {
	return nil
}

func (s *System) installGifsicleIfNeeded() error {
	return nil
}

func (s *System) installSvgoIfNeeded() error {
	return nil
}
