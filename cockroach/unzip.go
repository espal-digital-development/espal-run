package cockroach

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/juju/errors"
)

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

	for _, f := range r.File {
		if err := c.extractZipAndWriteFile(f, dest); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

func (c *Cockroach) extractZipAndWriteFile(f *zip.File, dest string) error {
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
		return errors.Trace(os.MkdirAll(path, f.Mode()))
	}

	if !f.FileInfo().IsDir() {
		if err := os.MkdirAll(filepath.Dir(path), f.Mode()); err != nil {
			return errors.Trace(err)
		}
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
