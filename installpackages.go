package main

import "github.com/espal-digital-development/espal-run/gopackage"

func installPackages() {
	// TODO :: Maybe just embed them in the vendor and build locally?
	staticCheck := gopackage.New("honnef.co/go/tools/cmd/staticcheck")
	goCheckStyle := gopackage.New("github.com/qiniu/checkstyle/gocheckstyle")
	errCheck := gopackage.New("github.com/kisielk/errcheck")
	qtc := gopackage.New("github.com/valyala/quicktemplate/qtc")
	// TODO :: 77777 The go list calls aren't working correctly due to the Go modules project sub-environment
	staticCheck.InstallIfNeeded(true)
	goCheckStyle.InstallIfNeeded(true)
	errCheck.InstallIfNeeded(true)
	qtc.InstallIfNeeded(true)
}
