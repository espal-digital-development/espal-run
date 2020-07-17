package gopackage

import (
	"bytes"
	"log"
	"os/exec"
)

// Package represents an object that provides Go package instances and their actions.
type Package interface {
	Name() string
	IsInstalled() bool
	Install() error
	InstallIfNeeded(verbose bool)
}

// GoPackage service object.
type GoPackage struct {
	name string
}

// Name returns the package fully qualified name.
func (p *GoPackage) Name() string {
	return p.name
}

// IsInstalled checks if the current package is installed on the system.
func (p *GoPackage) IsInstalled() bool {
	out, _ := exec.Command("go", "list", p.name).CombinedOutput()
	return string(bytes.Trim(out, "\n")) == p.name
}

// Install attempts to install the current package.
func (p *GoPackage) Install() error {
	return exec.Command("go", "get", "-u", p.name).Run()
}

// InstallIfNeeded will automatically install the package if it not was yet.
// The `verbose` option determines whether to print information about the process.
func (p *GoPackage) InstallIfNeeded(verbose bool) {
	if !p.IsInstalled() {
		if verbose {
			log.Printf("Did not find `%s`. Attempting to installing..\n", p.Name())
		}
		if err := p.Install(); err != nil && verbose {
			log.Printf("Failed to install `%s`\n", p.Name())
		}
	}
}

// New returns a new instance of GoPackage.
func New(name string) *GoPackage {
	return &GoPackage{
		name: name,
	}
}
