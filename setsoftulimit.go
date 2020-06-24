package main

import (
	"log"
	"os/exec"
	"runtime"
)

func setSoftUlimit() {
	// TODO :: Soft limit on Linux needs testing and is probably harder
	// TODO :: Going over the hard limit isn't allowed. Maybe just give a message and continue then
	if runtime.GOOS == darwinOS {
		// TODO :: Could be more efficient to first check if it's already `unlimited` or higher than `10032`. If so; do nothing
		// TODO :: Because this is a soft limit if might not work as an option for
		// the later-ran espal app?
		// Windows doesn't need this as it already sets it's limit dangerously high by default
		if err := exec.Command("ulimit", "-n", "10032").Run(); err != nil {
			log.Fatal(err)
		}
	}
}
