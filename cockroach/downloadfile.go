package cockroach

import (
	"io"
	"net/http"
	"os"

	"github.com/juju/errors"
)

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
