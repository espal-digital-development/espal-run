package configchecker

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/espal-digital-development/espal-run/randomstring"
	"github.com/juju/errors"
)

const (
	adminURLLength            = 4
	pprofURLLength            = 4
	defaultAssetsFilesPublic  = "./app/assets/files/public"
	defaultAssetsFilesPrivate = "./app/assets/files/private"
)

type configOption struct {
	Tag           string
	Info          string
	DefaultOption string
	Name          string
	Value         string
	// Only request this option on the full config generation
	RequestOnlyForFull bool
}

type ConfigChecker struct {
	path               string
	generateFullConfig bool
	randomString       *randomstring.RandomString
	textPainter        *textPainter
}

// GetPath gets path.
func (c *ConfigChecker) GetPath() string {
	return c.path
}

// SetPath sets path.
func (c *ConfigChecker) SetPath(path string) {
	c.path = path
}

// TODO :: Show explanation and info about SMTP server (local or services like Mailtrap)

// nolint:funlen,gocognit
func (c *ConfigChecker) Do() error {
	_, err := os.Stat(c.path)
	if err != nil && !os.IsNotExist(err) {
		return nil
	}
	if !os.IsNotExist(err) {
		return nil
	}

	log.Println("No configuration file found. Please answer the following to generate one:")
	configToRequest := c.defaultToRequest()
	reader := bufio.NewReader(os.Stdin)
	for _, configRequest := range configToRequest {
		if !c.generateFullConfig && configRequest.RequestOnlyForFull {
			continue
		}
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

	// Add some defaults without prompting
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

	var output []byte
	if c.generateFullConfig {
		output = configYmlExample
	} else {
		output = simpleConfigYmlExample
	}

	for _, configRequest := range configToRequest {
		output = bytes.Replace(output, []byte(configRequest.Tag), []byte(configRequest.Value), 1)
	}
	if err := ioutil.WriteFile(c.path, output, 0600); err != nil {
		return errors.Trace(err)
	}
	if err := c.generateBasicDirectories(); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (c *ConfigChecker) generateBasicDirectories() error {
	if err := c.generateDirIfNotExist(defaultAssetsFilesPublic); err != nil {
		return errors.Trace(err)
	}
	return errors.Trace(c.generateDirIfNotExist(defaultAssetsFilesPrivate))
}

func (c *ConfigChecker) generateDirIfNotExist(path string) error {
	_, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return errors.Trace(err)
	}
	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0700); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

// New returns a new instance of ConfigChecker.
func New(randomString *randomstring.RandomString, generateFullConfig bool) (*ConfigChecker, error) {
	c := &ConfigChecker{
		randomString: randomString,
		textPainter: &textPainter{
			reset:     "\033[m",
			lightBlue: "\033[0;34m",
			darkBlue:  "\033[0;94m",
		},
		generateFullConfig: generateFullConfig,
	}
	c.textPainter.resolveDefaults()
	return c, nil
}
