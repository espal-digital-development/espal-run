// +build unix darwin

package cockroach

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/juju/errors"
)

func (c *Cockroach) startNodeNonBlocking(storeName string, portsNumber int, httpPortsNumber int) error {
	if err := c.stopRunningNodes(storeName, portsNumber, "instance"); err != nil {
		return errors.Trace(err)
	}
	if err := c.stopRunningNodes(storeName, httpPortsNumber, "http interface"); err != nil {
		return errors.Trace(err)
	}

	log.Println("Starting `" + storeName + "`..")

	cmd := exec.Command("cockroach", "start", "--certs-dir="+c.certsDir,
		"--store="+filepath.FromSlash(c.databasePath+"/"+storeName),
		fmt.Sprintf("--listen-addr=%s:%d", c.host, portsNumber),
		fmt.Sprintf("--http-addr=%s:%d", c.httpHost, httpPortsNumber),
		"--join="+c.getHostsJoin(), "--background")

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
	go func() {
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
			log.Println(err)
		}
	}()
	return nil
}

func (c *Cockroach) stopRunningNodes(storeName string, portsNumber int, subject string) error {
	out, err := exec.Command("lsof", "-nP", fmt.Sprintf("-iTCP:%d", portsNumber)).CombinedOutput()
	if err != nil && len(out) > 0 {
		log.Println(string(out))
		return errors.Trace(err)
	} else if len(out) > 0 {
		matches := c.rePortListen.FindAllSubmatch(out, 1)
		if len(matches) > 0 && len(matches[0]) == 2 {
			log.Printf("Node `"+storeName+"` it's %s is still running. Trying to stop it..", subject)
			out, err = exec.Command("kill", string(matches[0][1])).CombinedOutput()
			if err != nil && !bytes.Contains(out, []byte("No such process")) {
				log.Println(string(out))
				return errors.Trace(err)
			}
		}
	}
	return nil
}
