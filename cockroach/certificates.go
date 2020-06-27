package cockroach

import (
	"log"
	"os/exec"
	"path/filepath"

	"github.com/juju/errors"
)

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
