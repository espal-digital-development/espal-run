package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/juju/errors"
)

type configOption struct {
	Tag           string
	Info          string
	DefaultOption string
	Name          string
	Value         string
}

func checkConfig() error {
	configPath := "./app/config.yml"
	_, err := os.Stat(configPath)
	if err != nil && !os.IsNotExist(err) {
		return nil
	}
	if !os.IsNotExist(err) {
		return nil
	}

	// TODO :: 77 Show explanation and info about SMTP server (local or services like Mailtrap)

	configToRequest := []*configOption{
		{
			Tag:           "#DATABASE_HOST",
			Name:          "Database host",
			Info:          "\033[0;34m0.0.0.0\033[m, \033[0;34m127.0.0.1\033[m, \033[0;94mlocalhost\033[m",
			DefaultOption: "localhost",
		},
		{
			Tag:           "#DATABASE_PORT",
			Name:          "Database port",
			Info:          "\033[0;94m26257\033[m",
			DefaultOption: "26257",
		},
		{
			Tag:           "#DATABASE_NAME",
			Name:          "Database name",
			Info:          "\033[0;94mapp\033[m",
			DefaultOption: "app",
		},
		{
			Tag:  "#EMAIL_HOST",
			Name: "Email host",
			Info: "smtp.domain.com",
		},
		{
			Tag:           "#EMAIL_PORT",
			Name:          "Email port",
			DefaultOption: "2525",
		},
		{
			Tag:  "#EMAIL_USERNAME",
			Name: "Email username",
		},
		{
			Tag:  "#EMAIL_PASSWORD",
			Name: "Email password",
		},
		{
			Tag:  "#EMAIL_NO_REPLY_ADDRESS",
			Name: "Email no-reply address",
			Info: "noreply@domain.com",
		},
	}

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
			Value: "_" + randomString(4),
		},
		{
			Tag:   "#PPROF_ADMIN",
			Value: "_" + randomString(4),
		},
	}...)

	for _, configRequest := range configToRequest {
		output = bytes.Replace(output, []byte(configRequest.Tag), []byte(configRequest.Value), 1)
	}

	if err := ioutil.WriteFile(configPath, output, 0644); err != nil {
		return errors.Trace(err)
	}

	return nil
}
