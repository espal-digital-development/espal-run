package cockroach

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/juju/errors"
)

// Resolve checks and sets up everything needed for Cockroach DB.
func (c *Cockroach) Resolve() error {
	if err := c.validate(); err != nil {
		return errors.Trace(err)
	}
	if err := c.checkInstall(); err != nil {
		return errors.Trace(err)
	}

	_, err := os.Stat(c.databasePath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Trace(err)
	}
	switch {
	case c.resetDB && os.IsNotExist(err):
		log.Println(cockroachResettingDatabaseNotRequired)
		return nil
	case !c.resetDB && !os.IsNotExist(err):
		if err := c.runNodes(); err != nil {
			return errors.Trace(err)
		}
		return nil
	case c.resetDB:
		log.Println(cockroachResettingDatabase)
	case os.IsNotExist(err):
		log.Println(cockroachCreatingNewDatabase)
	}

	if err := c.setupDirectories(); err != nil {
		return errors.Trace(err)
	}
	if err := c.generateCertificates(); err != nil {
		return errors.Trace(err)
	}
	if err := c.createPrimaryNode(); err != nil {
		return errors.Trace(err)
	}

	if err := c.runNodes(); err != nil {
		return errors.Trace(err)
	}

	if err := c.initializeCluster(); err != nil {
		return errors.Trace(err)
	}

	if err := c.generateDatabaseUsers(); err != nil {
		return errors.Trace(err)
	}
	if err := c.generateHTTPInterfaceUser(); err != nil {
		return errors.Trace(err)
	}

	if err := c.generateDatabaseUserCertificates(); err != nil {
		return errors.Trace(err)
	}

	return errors.Trace(c.report())
}

func (c *Cockroach) setupDirectories() error {
	stat, err := os.Stat(c.databasePath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Trace(err)
	}
	if stat != nil && c.resetDB {
		if err := os.RemoveAll(c.databasePath); err != nil {
			return errors.Trace(err)
		}
	}

	log.Println("Creating certs dir..")
	_, err = os.Stat(c.certsDir)
	if err != nil && !os.IsNotExist(err) {
		return errors.Trace(err)
	}
	if os.IsNotExist(err) {
		if err := os.MkdirAll(c.certsDir, 0740); err != nil {
			return errors.Trace(err)
		}
	}
	_, err = os.Stat(c.safeDir)
	if err != nil && !os.IsNotExist(err) {
		return errors.Trace(err)
	}
	if os.IsNotExist(err) {
		if err := os.MkdirAll(c.safeDir, 0740); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

func (c *Cockroach) getHostsJoin() string {
	joinsString := ""
	portsNumber := c.portStart
	var firstHad = false
	for i := 0; i < c.desiredNodes; i++ {
		if firstHad {
			joinsString += ","
		} else {
			firstHad = true
		}
		joinsString += fmt.Sprintf("%s:%d", c.host, portsNumber)
		portsNumber++
	}
	return joinsString
}

func (c *Cockroach) runNodes() error {
	portsNumber := c.portStart
	httpPortsNumber := c.httpPortStart
	for i := 0; i < c.desiredNodes; i++ {
		storeName := fmt.Sprintf("%s%d", "node", i+1)
		if err := c.startNodeNonBlocking(storeName, portsNumber, httpPortsNumber); err != nil {
			return errors.Trace(err)
		}

		// TODO :: This is a wait guess, but might be slower on some devices
		// and might need a better detection mechanism (maybe `lsof -nP -iTCP:36257 | grep LISTEN`?)
		time.Sleep(secondsIntervalBetweenNodesStart * time.Second)
		portsNumber++
		httpPortsNumber++
	}
	return nil
}

func (c *Cockroach) initializeCluster() error {
	log.Println("Initializing the cluster..")
	out, err := exec.Command("cockroach", "init", "--certs-dir="+c.certsDir,
		"--host="+fmt.Sprintf("%s:%d", c.host, c.portStart)).CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return errors.Trace(err)
	}
	return nil
}

func (c *Cockroach) report() error {
	fmt.Println("")
	log.Println("All done! You can now login to the http interface:")
	fmt.Println("")
	fmt.Printf("  Address:  https://%s:%d\n", c.httpHost, c.httpPortStart)
	fmt.Printf("  User:     %s\n", c.httpUser)
	fmt.Printf("  Password: %s\n", c.httpPassword)
	fmt.Println("")
	fmt.Println("  STORE THIS INFORMATION SOMEWHERE SAFE! IT WON'T BE DISPLAYED AGAIN.")
	fmt.Println("")

	// TODO :: There should be a non-interactive mode so this won't block
	// when being executed inside in scripts.
	fmt.Println("Press any key to continue..")
	reader := bufio.NewReader(os.Stdin)
	_, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return errors.Trace(err)
	}
	return nil
}
