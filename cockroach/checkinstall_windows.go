// +build windows

package cockroach

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/espal-digital-development/system/permissions"
	"github.com/juju/errors"
)

func (c *Cockroach) checkInstall() error {
	// TODO :: Need to dynamically detect the latest compatible cockroach version
	// and fetch that (max) version binary.
	// TODO :: Ask the user if saving to the dbDirpath is OK? Otherwise they might
	// provide another path.
	packageName := "cockroach-" + c.version + ".windows-6.2-amd64"
	zipURL := "https://binaries.cockroachdb.com/" + packageName + ".zip"
	dbDirPath, err := filepath.Abs("/CockroachDB")
	if err != nil {
		return errors.Trace(err)
	}
	dbBinPath := dbDirPath + "/bin"
	dbPath := dbBinPath + "/cockroach.exe"
	zipPath := dbDirPath + "/cockroach.zip"
	unzippedPath := dbDirPath + "/" + packageName

	_, err = os.Stat(dbDirPath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Trace(err)
	}
	if os.IsNotExist(err) {
		if err := os.MkdirAll(dbDirPath, permissions.UserReadWrite); err != nil {
			return errors.Trace(err)
		}
	}

	_, err = os.Stat(dbPath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Trace(err)
	}
	if os.IsNotExist(err) {
		if err := c.download(dbBinPath, dbDirPath, zipPath, zipURL, unzippedPath); err != nil {
			return errors.Trace(err)
		}
		if err := c.install(dbPath, zipPath, unzippedPath); err != nil {
			return errors.Trace(err)
		}
	}

	zoneInfoURL := "https://github.com/golang/go/raw/master/lib/time/zoneinfo.zip"
	zoneInfoPath := dbDirPath + "/go-zoneinfo.zip"
	_, err = os.Stat(zoneInfoPath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Trace(err)
	}
	if os.IsNotExist(err) {
		log.Println("Downloading Go Zoneinfo..")
		if err := c.downloadFile(zoneInfoPath, zoneInfoURL); err != nil {
			return errors.Trace(err)
		}
	}

	// TODO :: These are soft-sets. Need a way to hard set these, if even possible
	if !strings.Contains(os.Getenv("PATH"), dbBinPath) {
		os.Setenv("PATH", os.Getenv("PATH")+";"+dbBinPath)
	}
	if os.Getenv("ZONEINFO") == zoneInfoPath {
		os.Setenv("ZONEINFO", zoneInfoPath)
	}

	return nil
}

func (c *Cockroach) download(dbBinPath string, dbDirPath string, zipPath string, zipURL string,
	unzippedPath string) error {
	_, err := os.Stat(zipPath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Trace(err)
	}
	if os.IsNotExist(err) {
		log.Println("Downloading CockroachDB. This may take a while..")
		if err := c.downloadFile(zipPath, zipURL); err != nil {
			return errors.Trace(err)
		}
	}

	_, err = os.Stat(dbBinPath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Trace(err)
	}
	if os.IsNotExist(err) {
		if err := os.MkdirAll(dbBinPath, permissions.UserReadWrite); err != nil {
			return errors.Trace(err)
		}
	}

	_, err = os.Stat(unzippedPath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Trace(err)
	}
	if os.IsNotExist(err) {
		log.Println("Extracting..")
		if err := c.unzip(zipPath, dbDirPath); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func (c *Cockroach) install(dbPath string, zipPath string, unzippedPath string) error {
	log.Println("Installing..")
	unzippedBinaryPath := unzippedPath + "/cockroach.exe"
	if err := os.Rename(unzippedBinaryPath, dbPath); err != nil {
		return errors.Trace(err)
	}

	if err := os.Remove(unzippedPath); err != nil {
		return errors.Trace(err)
	}
	if err := os.Remove(zipPath); err != nil {
		return errors.Trace(err)
	}
	return nil
}
