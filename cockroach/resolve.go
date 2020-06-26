package cockroach

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	if c.resetDB && os.IsNotExist(err) {
		log.Println(cockroachResettingDatabaseNotRequired)
		return nil
	} else if c.resetDB {
		log.Println(cockroachResettingDatabase)
	} else if os.IsNotExist(err) {
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

	portsNumber := c.portStart
	httpPortsNumber := c.httpPortStart
	for i := 0; i < c.desiredNodes; i++ {
		storeName := fmt.Sprintf("%s%d", "node", i+1)
		if err := c.startNodeNonBlocking(storeName, portsNumber, httpPortsNumber); err != nil {
			return errors.Trace(err)
		}

		// TODO :: This is a wait guess, but might be slower on some devices
		// and might need a better detection mechanism (maybe `lsof -nP -iTCP:26257 | grep LISTEN`?)
		time.Sleep(secondsIntervalBetweenNodesStart * time.Second)
		portsNumber++
		httpPortsNumber++
	}

	// TODO :: Continue from here (startNode function above)
	if true {
		return errors.New("STOP")
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

	return errors.Errorf("STOP")

	if err := c.report(); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (c *Cockroach) report() error {
	fmt.Println("")
	log.Println("All done! You can no login to the http interface:")
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

func (c *Cockroach) setupDirectories() error {
	_, err := os.Stat(c.databasePath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Trace(err)
	}
	if os.IsExist(err) && c.resetDB {
		if c.isUnixOS() {
			out, err := exec.Command("rm", "-rf", c.databasePath).CombinedOutput()
			if err != nil {
				log.Println(string(out))
				return errors.Trace(err)
			}
		} else if c.isWinOS() {
			out, err := exec.Command("rmdir", "/S", c.databasePath).CombinedOutput()
			if err != nil {
				log.Println(string(out))
				return errors.Trace(err)
			}
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

func (c *Cockroach) generateCertificates() error {
	log.Println("Generating ca key..")
	out, err := exec.Command("cockroach", "cert", "create-ca",
		"--certs-dir="+c.certsDir,
		"--ca-key="+filepath.FromSlash(c.safeDir+"/"+c.caKeyName)).CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return errors.Trace(err)
	}

	log.Println("Creating certficate..")
	out, err = exec.Command("cockroach", "cert", "create-client", c.rootUser,
		"--certs-dir="+c.certsDir,
		"--ca-key="+filepath.FromSlash(c.safeDir+"/"+c.caKeyName)).CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return errors.Trace(err)
	}
	return nil
}

func (c *Cockroach) createPrimaryNode() error {
	log.Println("Creating primary node..")
	out, err := exec.Command("cockroach", "cert", "create-node", c.host, "$(hostname)",
		"--certs-dir="+c.certsDir,
		"--ca-key="+filepath.FromSlash(c.safeDir+"/"+c.caKeyName)).CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return errors.Trace(err)
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

func (c *Cockroach) generateDatabaseUsers() error {
	log.Println("Generating database, users, roles and assigning privileges..")
	tmpSQLFile := filepath.FromSlash(os.TempDir() + "/tmp.sql")
	if err := ioutil.WriteFile(tmpSQLFile, []byte(setupDatabaseSQL), 0700); err != nil {
		return errors.Trace(err)
	}
	tmpSHFile := filepath.FromSlash(os.TempDir() + "/tmp.sh")
	if err := ioutil.WriteFile(tmpSHFile,
		[]byte(fmt.Sprintf("#!/bin/sh\n\n"+`cockroach sql --certs-dir=%s --host=%s:%d < %s`,
			c.certsDir, c.host, c.portStart, tmpSQLFile)), 0700); err != nil {
		return errors.Trace(err)
	}
	out, err := exec.Command("/bin/sh", tmpSHFile).CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return errors.Trace(err)
	}
	return nil
}

func (c *Cockroach) generateHTTPInterfaceUser() error {
	log.Println("Generating http interface user..")
	tmpSQLFile := filepath.FromSlash(os.TempDir() + "/tmp.sql")
	if err := ioutil.WriteFile(tmpSQLFile,
		[]byte(fmt.Sprintf(httpUserSQL, c.httpUser, c.httpPassword, c.httpUser)), 0700); err != nil {
		return errors.Trace(err)
	}
	tmpSHFile := filepath.FromSlash(os.TempDir() + "/tmp.sh")
	if err := ioutil.WriteFile(tmpSHFile,
		[]byte(fmt.Sprintf("#!/bin/sh\n\n"+`cockroach sql --certs-dir=%s --host=%s:%d < %s`,
			c.certsDir, c.host, c.portStart, tmpSQLFile)), 0700); err != nil {
		return errors.Trace(err)
	}
	out, err := exec.Command("/bin/sh", tmpSHFile).CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return errors.Trace(err)
	}
	return nil
}

func (c *Cockroach) generateDatabaseUserCertificates() error {
	users := []string{"selecter", "creator", "inserter", "updater", "deletor", "migrator"}
	for k := range users {
		log.Printf("Creating certificate for user `%s`..", users[k])
		out, err := exec.Command("cockroach", "cert", "create-client", users[k],
			"--certs-dir="+c.certsDir,
			"--ca-key="+filepath.FromSlash(c.safeDir+"/"+c.caKeyName)).CombinedOutput()
		if err != nil {
			log.Println(string(out))
			return errors.Trace(err)
		}
	}
	return nil
}
