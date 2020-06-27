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
			if err := os.MkdirAll(path, f.Mode()); err != nil {
				return errors.Trace(err)
			}
		} else {
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

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}
