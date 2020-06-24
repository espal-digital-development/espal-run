package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"time"

	"github.com/juju/errors"
)

const (
	cockroachNotFoundInstalling           = "Did not find `cockroach`. Attempting to installing.."
	cockroachCreatingNewDatabase          = "Creating a new cockroach database.."
	cockroachResettingDatabase            = "Resetting the cockroach database.."
	cockroachResettingDatabaseNotRequired = "No database found. Skipping reset.."
)

func isUnixOS() bool {
	return runtime.GOOS == darwinOS || runtime.GOOS == linuxOS
}

func isWinOS() bool {
	return runtime.GOOS == windowsOS
}

func configureCockroachDB() error {
	// TODO :: Haproxy as well for full fanciness?

	if isUnixOS() {
		out, _ := exec.Command("which", "cockroach").CombinedOutput()
		isInstalled := bytes.Contains(out, []byte("/cockroach"))
		if !isInstalled {
			log.Println(cockroachNotFoundInstalling)
			out, err := exec.Command("brew", "install", "cockroach").CombinedOutput()
			if err != nil {
				log.Println(string(out))
				return errors.Trace(err)
			}
		}
	} else if isWinOS() {
		// TODO :: Needs a Windows variance
		return errors.Errorf("No Windows auto-installation detection implemented yet..")
	}

	// TODO :: Auto-detect info based on config.yml?
	amountOfDesiredNodes := 8
	host := "localhost"
	rootUser := "root"
	portStart := 26257
	httpHost := "localhost"
	httpUser := "espal"
	httpPassword := randomString(32) // TODO :: Something safer, like `openssl rand -hex 16`
	httpPortStart := 26080
	databaseDir, err := filepath.Abs(filepath.FromSlash("./app/database"))
	if err != nil {
		return errors.Trace(err)
	}
	certsDir := filepath.FromSlash(databaseDir + "/certs")
	safeDir := filepath.FromSlash(databaseDir + "/safe")
	nodeBaseName := "node" // Not sure if this can be edited in the build. Default for now
	caKeyName := "ca.key"

	if amountOfDesiredNodes < 1 {
		return errors.Errorf("amountOfDesiredNodes has to be at least one. %d given", amountOfDesiredNodes)
	}

	if isUnixOS() {
		_, err := os.Stat(databaseDir)
		if err != nil && !os.IsNotExist(err) {
			return errors.Trace(err)
		}

		// TODO :: Should also really check if all the content for the
		// database is there and maybe if it's running at all?

		var stopAndDelete bool
		if resetDB && os.IsNotExist(err) {
			log.Println(cockroachResettingDatabaseNotRequired)
		} else if resetDB {
			log.Println(cockroachResettingDatabase)
			stopAndDelete = true
		} else if os.IsNotExist(err) {
			log.Println(cockroachCreatingNewDatabase)
			stopAndDelete = true
		}

		if stopAndDelete {
			// TODO :: If you really mess previous runs up cockroach keeps running as a zombie process.
			// Need to find a way to gracefully stop those as well, or ask the user, because killing
			// all cockroach processes might killed other nodes that the user might have running on their local machine.
			portsNumber := portStart
			for {
				endpoint := fmt.Sprintf("%s:%d", host, portsNumber)
				out, err := exec.Command("cockroach", "quit",
					"--certs-dir="+certsDir, "--host="+endpoint).CombinedOutput()
				if err != nil {
					if !bytes.Contains(out, []byte("cannot dial server")) &&
						!bytes.Contains(out, []byte("Failed to connect to the node")) &&
						!bytes.Contains(out, []byte("node cannot be shut down before it has been initialized")) {
						log.Println(string(out))
						return errors.Trace(err)
					}
				}
				if !bytes.Contains(out, []byte("node is draining")) {
					break
				}
				log.Println("Stopped node `" + endpoint + "`..")
				portsNumber++
				if true {
					// TODO :: This sometimes hangs on multiple nodes. Stopping it for now
					// and letting the auto-killer do it's job further on.
					break
				}
			}

			out, err := exec.Command("rm", "-rf", databaseDir).CombinedOutput()
			if err != nil {
				log.Println(string(out))
				return errors.Trace(err)
			}

			log.Println("Creating certs dir..")
			out, err = exec.Command("mkdir", "-m", "740", "-p", certsDir).CombinedOutput()
			if err != nil {
				log.Println(string(out))
				return errors.Trace(err)
			}

			log.Println("Creating safe dir..")
			out, err = exec.Command("mkdir", "-m", "740", "-p", safeDir).CombinedOutput()
			if err != nil {
				log.Println(string(out))
				return errors.Trace(err)
			}

			// TODO :: Detect errors in all the setup calls below here
			log.Println("Generating ca key..")
			out, err = exec.Command("cockroach", "cert", "create-ca",
				"--certs-dir="+certsDir, "--ca-key="+filepath.FromSlash(safeDir+"/"+caKeyName)).CombinedOutput()
			if err != nil {
				log.Println(string(out))
				return errors.Trace(err)
			}

			log.Println("Creating certficate..")
			out, err = exec.Command("cockroach", "cert", "create-client", rootUser,
				"--certs-dir="+certsDir, "--ca-key="+filepath.FromSlash(safeDir+"/"+caKeyName)).CombinedOutput()
			if err != nil {
				log.Println(string(out))
				return errors.Trace(err)
			}

			log.Println("Creating primary node..")
			out, err = exec.Command("cockroach", "cert", "create-node", host, "$(hostname)",
				"--certs-dir="+certsDir, "--ca-key="+filepath.FromSlash(safeDir+"/"+caKeyName)).CombinedOutput()
			if err != nil {
				log.Println(string(out))
				return errors.Trace(err)
			}

			joinsString := ""
			portsNumber = portStart
			var firstHad = false
			for i := 0; i < amountOfDesiredNodes; i++ {
				if firstHad {
					joinsString += ","
				} else {
					firstHad = true
				}
				joinsString += fmt.Sprintf("%s:%d", host, portsNumber)
				portsNumber++
			}

			rePortListen, err := regexp.Compile(`(?m)^\s*cockroach\s+(\d+).*?\(LISTEN\)\s*$`)
			if err != nil {
				return errors.Trace(err)
			}

			portsNumber = portStart
			httpPortsNumber := httpPortStart
			for i := 0; i < amountOfDesiredNodes; i++ {
				storeName := fmt.Sprintf("%s%d", nodeBaseName, i+1)

				out, err = exec.Command("lsof", "-nP", fmt.Sprintf("-iTCP:%d", portsNumber)).CombinedOutput()
				if err != nil && len(out) > 0 {
					log.Println(string(out))
					return errors.Trace(err)
				} else if len(out) > 0 {
					matches := rePortListen.FindAllSubmatch(out, 1)
					if len(matches) > 0 && len(matches[0]) == 2 {
						log.Println("Node `" + storeName + "` is still running. Trying to stop it..")
						out, err = exec.Command("kill", string(matches[0][1])).CombinedOutput()
						if err != nil && !bytes.Contains(out, []byte("No such process")) {
							log.Println(string(out))
							return errors.Trace(err)
						}
					}
				}

				out, err = exec.Command("lsof", "-nP", fmt.Sprintf("-iTCP:%d", httpPortsNumber)).CombinedOutput()
				if err != nil && len(out) > 0 {
					log.Println(string(out))
					return errors.Trace(err)
				} else if len(out) > 0 {
					matches := rePortListen.FindAllSubmatch(out, 1)
					if len(matches) > 0 && len(matches[0]) == 2 {
						log.Println("Node `" + storeName + "` it's web interface is still running. Trying to stop it..")
						out, err = exec.Command("kill", string(matches[0][1])).CombinedOutput()
						if err != nil && !bytes.Contains(out, []byte("No such process")) {
							log.Println(string(out))
							return errors.Trace(err)
						}
					}
				}

				log.Println("Starting `" + storeName + "`..")

				cmd := exec.Command("cockroach", "start", "--certs-dir="+certsDir,
					"--store="+filepath.FromSlash(databaseDir+"/"+storeName),
					fmt.Sprintf("--listen-addr=%s:%d", host, portsNumber),
					fmt.Sprintf("--http-addr=%s:%d", httpHost, httpPortsNumber),
					"--join="+joinsString, "--background")

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

				// TODO :: This is a wait guess, but might be slower on some devices
				// and might need a better detection mechanism (maybe `lsof -nP -iTCP:26257 | grep LISTEN`?)
				time.Sleep(3 * time.Second)
				portsNumber++
				httpPortsNumber++
			}

			log.Println("Initializing the cluster..")
			out, err = exec.Command("cockroach", "init", "--certs-dir="+certsDir,
				"--host="+fmt.Sprintf("%s:%d", host, portStart)).CombinedOutput()
			if err != nil {
				log.Println(string(out))
				return errors.Trace(err)
			}

			log.Println("Generating database, users, roles and assigning privileges..")
			tmpSQLFile := filepath.FromSlash(os.TempDir() + "/tmp.sql")
			if err := ioutil.WriteFile(tmpSQLFile, []byte(setupDatabaseSQL), 0700); err != nil {
				return errors.Trace(err)
			}
			tmpSHFile := filepath.FromSlash(os.TempDir() + "/tmp.sh")
			if err := ioutil.WriteFile(tmpSHFile,
				[]byte(fmt.Sprintf("#!/bin/sh\n\n"+`cockroach sql --certs-dir=%s --host=%s:%d < %s`,
					certsDir, host, portStart, tmpSQLFile)), 0700); err != nil {
				return errors.Trace(err)
			}
			out, err = exec.Command("/bin/sh", tmpSHFile).CombinedOutput()
			if err != nil {
				log.Println(string(out))
				return errors.Trace(err)
			}

			log.Println("Generating http interface user..")
			if err := ioutil.WriteFile(tmpSQLFile,
				[]byte(fmt.Sprintf(httpUserSQL, httpUser, httpPassword, httpUser)), 0700); err != nil {
				return errors.Trace(err)
			}
			if err := ioutil.WriteFile(tmpSHFile,
				[]byte(fmt.Sprintf("#!/bin/sh\n\n"+`cockroach sql --certs-dir=%s --host=%s:%d < %s`,
					certsDir, host, portStart, tmpSQLFile)), 0700); err != nil {
				return errors.Trace(err)
			}
			out, err = exec.Command("/bin/sh", tmpSHFile).CombinedOutput()
			if err != nil {
				log.Println(string(out))
				return errors.Trace(err)
			}

			users := []string{"selecter", "creator", "inserter", "updater", "deletor", "migrator"}
			for k := range users {
				log.Printf("Creating certificate for user `%s`..", users[k])
				out, err = exec.Command("cockroach", "cert", "create-client", users[k],
					"--certs-dir="+certsDir, "--ca-key="+filepath.FromSlash(safeDir+"/"+caKeyName)).CombinedOutput()
				if err != nil {
					log.Println(string(out))
					return errors.Trace(err)
				}
			}

			fmt.Println("")

			log.Println("All done! You can no login to the http interface:")
			fmt.Println("")
			fmt.Printf("  Address:  https://%s:%d\n", httpHost, httpPortStart)
			fmt.Printf("  User:     %s\n", httpUser)
			fmt.Printf("  Password: %s\n", httpPassword)
			fmt.Println("")
			fmt.Println("  STORE THIS INFORMATION SOMEWHERE SAFE! IT WON'T BE DISPLAYED AGAIN.")
			fmt.Println("")

			// TODO :: There should be a non-interactive mode so this won't block
			// when being executed inside in scripts.
			fmt.Println("Press any key to continue..")
			reader := bufio.NewReader(os.Stdin)
			_, err = reader.ReadString('\n')
			if err != nil && err != io.EOF {
				return errors.Trace(err)
			}
		}
	} else if isWinOS() {
		// TODO :: Needs a Windows variance
		return errors.Errorf("No Windows setup/reset implemented yet..")
	}

	return nil
}

const httpUserSQL = `
CREATE USER %s WITH PASSWORD '%s';
GRANT admin to %s;
`
const setupDatabaseSQL = `CREATE DATABASE app;

CREATE USER selecter;
GRANT SELECT ON DATABASE app TO selecter;

CREATE USER creator;
GRANT SELECT ON DATABASE app TO creator;
GRANT CREATE ON DATABASE app TO creator;

CREATE USER inserter;
GRANT SELECT ON DATABASE app TO inserter;
GRANT INSERT ON DATABASE app TO inserter;

CREATE USER updater;
GRANT SELECT ON DATABASE app TO updater;
GRANT UPDATE ON DATABASE app TO updater;

CREATE USER deletor;
GRANT SELECT ON DATABASE app TO deletor;
GRANT DELETE ON DATABASE app TO deletor;

CREATE USER migrator;
GRANT GRANT ON DATABASE app TO migrator;
GRANT CREATE ON DATABASE app TO migrator;
GRANT DROP ON DATABASE app TO migrator;
GRANT SELECT ON DATABASE app TO migrator;
GRANT INSERT ON DATABASE app TO migrator;
GRANT UPDATE ON DATABASE app TO migrator;
GRANT DELETE ON DATABASE app TO migrator;
`
