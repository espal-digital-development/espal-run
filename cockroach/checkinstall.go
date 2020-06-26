package cockroach

import (
	"archive/zip"
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/juju/errors"
)

func (c *Cockroach) checkInstall() error {
	if c.isUnixOS() {
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
	} else if c.isWinOS() {
		// TODO :: Need to dynamically detect the latest compatible cockroach version
		// and fetch that (max) version binary.
		// TODO :: Ask the user if saving to the dbDirpath is OK? Otherwise they might
		// give another path.
		packageName := "cockroach-v20.1.2.windows-6.2-amd64"
		zipURL := "https://binaries.cockroachdb.com/" + packageName + ".zip"
		dbDirPath, err := filepath.Abs("/CockroachDB")
		if err != nil {
			return errors.Trace(err)
		}
		dbBinPath := dbDirPath + filepath.FromSlash("/bin")
		dbPath := dbBinPath + filepath.FromSlash("/cockroach.exe")
		zipPath := dbDirPath + filepath.FromSlash("/cockroach.zip")
		unzippedPath := dbDirPath + filepath.FromSlash("/"+packageName)

		_, err = os.Stat(dbDirPath)
		if err != nil && !os.IsNotExist(err) {
			return errors.Trace(err)
		}
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dbDirPath, 0600); err != nil {
				return errors.Trace(err)
			}
		}

		_, err = os.Stat(dbPath)
		if err != nil && !os.IsNotExist(err) {
			return errors.Trace(err)
		}
		if os.IsNotExist(err) {
			_, err = os.Stat(zipPath)
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
				if err := os.MkdirAll(dbBinPath, 0600); err != nil {
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

			log.Println("Installing..")
			unzippedBinaryPath := unzippedPath + filepath.FromSlash("/cockroach.exe")
			if err := os.Rename(unzippedBinaryPath, dbPath); err != nil {
				return errors.Trace(err)
			}

			if err := os.Remove(unzippedPath); err != nil {
				return errors.Trace(err)
			}
			if err := os.Remove(zipPath); err != nil {
				return errors.Trace(err)
			}
		}

		zoneInfoURL := "https://github.com/golang/go/raw/master/lib/time/zoneinfo.zip"
		zoneInfoPath := dbDirPath + filepath.FromSlash("/go-zoneinfo.zip")
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
	}
	return nil
}

func (c *Cockroach) downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return errors.Trace(err)
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return errors.Trace(err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return errors.Trace(err)
}

func (c *Cockroach) unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return errors.Trace(err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			log.Println(err)
		}
	}()

	if err := os.MkdirAll(dest, 0600); err != nil {
		return errors.Trace(err)
	}

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return errors.Trace(err)
		}
		defer func() {
			if err := rc.Close(); err != nil {
				log.Println(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					log.Println(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return errors.Trace(err)
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}
