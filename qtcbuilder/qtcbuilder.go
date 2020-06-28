package qtcbuilder

import (
	"bytes"
	"log"

	"github.com/juju/errors"
)

type QTCBuilder struct {
}

func (b *QTCBuilder) Do() error {
	log.Println("Building templates. Please wait..")
	out, err := b.build()
	if err != nil {
		return errors.Trace(err)
	}
	if bytes.Contains(out, []byte("error")) {
		log.Println(string(out))
		return errors.Errorf("there were errors %s", string(out))
	}
	return nil
}

// New returns a new instance of QTCBuilder.
func New() (*QTCBuilder, error) {
	b := &QTCBuilder{}
	return b, nil
}
