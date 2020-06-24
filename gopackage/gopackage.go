package gopackage

import (
	"bytes"
	"log"
	"os/exec"
)

// Package represents an object that provides Go package
// instances and their actions.
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
func (goPackage *GoPackage) Name() string {
	return goPackage.name
}

// IsInstalled checks if the current package is installed on the system.
func (goPackage *GoPackage) IsInstalled() bool {
	out, _ := exec.Command("go", "list", goPackage.name).CombinedOutput()
	return string(bytes.Trim(out, "\n")) == goPackage.name
}

// Install attempts to install the current package.
func (goPackage *GoPackage) Install() error {
	return exec.Command("go", "get", "-u", goPackage.name).Run()
}

// InstallIfNeeded will automatically install the package if it not was yet.
// The `verbose` option determines whether to print information about the process.
func (goPackage *GoPackage) InstallIfNeeded(verbose bool) {
	if !goPackage.IsInstalled() {
		if verbose {
			log.Printf("Did not find `%s`. Attempting to installing..\n", goPackage.Name())
		}
		if err := goPackage.Install(); err != nil && verbose {
			log.Printf("Failed to install `%s`\n", goPackage.Name())
		}
	}
}

// New returns a new instance of GoPackage.
func New(name string) *GoPackage {
	return &GoPackage{
		name: name,
	}
}
