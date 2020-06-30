package configchecker

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/espal-digital-development/espal-run/randomstring"
	"github.com/juju/errors"
)

const (
	adminURLLength = 4
	pprofURLLength = 4
)

type configOption struct {
	Tag           string
	Info          string
	DefaultOption string
	Name          string
	Value         string
}

type ConfigChecker struct {
	randomString *randomstring.RandomString
	path         string
	textPainter  *textPainter
}

// GetPath gets path.
func (c *ConfigChecker) GetPath() string {
	return c.path
}

// SetPath sets path.
func (c *ConfigChecker) SetPath(path string) {
	c.path = path
}

// TODO :: 77 Show explanation and info about SMTP server (local or services like Mailtrap)

func (c *ConfigChecker) Do() error {
	_, err := os.Stat(c.path)
	if err != nil && !os.IsNotExist(err) {
		return nil
	}
	if !os.IsNotExist(err) {
		return nil
	}

	configToRequest := c.defaultToRequest()
	reader := bufio.NewReader(os.Stdin)
	for _, configRequest := range configToRequest {
		for {
			fmt.Printf(configRequest.Name)
			if configRequest.Info != "" {
				fmt.Printf(" <%s>", configRequest.Info)
			}
			fmt.Printf(" : ")
			value, err := reader.ReadString('\n')
			if err != nil && err == io.EOF {
				break
			}
			if err != nil {
				return errors.Trace(err)
			}
			value = strings.Trim(value, "\n")
			if value == "" {
				if configRequest.DefaultOption == "" {
					continue
				} else {
					value = configRequest.DefaultOption
				}
			}
			configRequest.Value = value
			break
		}
	}

	output := configYmlExample

	// Add some defaults without
	configToRequest = append(configToRequest, []*configOption{
		{
			Tag:   "#URLS_ADMIN",
			Value: "_" + c.randomString.Simple(adminURLLength),
		},
		{
			Tag:   "#PPROF_ADMIN",
			Value: "_" + c.randomString.Simple(pprofURLLength),
		},
	}...)

	for _, configRequest := range configToRequest {
		output = bytes.Replace(output, []byte(configRequest.Tag), []byte(configRequest.Value), 1)
	}

	if err := ioutil.WriteFile(c.path, output, 0600); err != nil {
		return errors.Trace(err)
	}
	return nil
}

// New returns a new instance of ConfigChecker.
func New(randomString *randomstring.RandomString) (*ConfigChecker, error) {
	c := &ConfigChecker{
		randomString: randomString,
		textPainter: &textPainter{
			reset:     "\033[m",
			lightBlue: "\033[0;34m",
			darkBlue:  "\033[0;94m",
		},
	}
	c.textPainter.resolveDefaults()
	return c, nil
}
