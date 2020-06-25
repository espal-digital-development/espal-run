package main

import (
	"bufio"
	"fmt"
	"os/exec"

	"github.com/juju/errors"
)

func run() error {
	cmd := exec.Command("go", "run", cwd+"/main.go")

	// TODO :: When calling TERM on this command, it needs to gracefully stop the espal-core
	// too (prove will be when it reports it's winddown info about how long the server ran for).
	// TODO :: If the output doesn't stop with a newline or throws an error,
	// it probably won't show anything at all. Needs more testing.
	stdOut, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Trace(err)
	}
	stdErr, err := cmd.StderrPipe()
	if err != nil {
		return errors.Trace(err)
	}
	if err := cmd.Start(); err != nil {
		return errors.Trace(err)
	}
	scanner := bufio.NewScanner(stdOut)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}
	errScanner := bufio.NewScanner(stdErr)
	for errScanner.Scan() {
		m := errScanner.Text()
		fmt.Println(m)
	}
	if err := cmd.Wait(); err != nil {
		return errors.Trace(err)
	}
	return nil
}
