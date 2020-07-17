package projectcreator

// nolint:gochecknoglobals
var (
	gitIgnoreFile = []byte(`/app/database
/app/server
/tmp
`)
	espalRunFile = []byte(`verbosity: quiet
ignoredDirectories: ['tmp', 'app/assets/vue/**/*', 'app/database/**/*', 'node_modules']
exclusiveDirectories: ['app/assets/**/*'] # This will ignore other inclusion rules
`)
	mainGoFile = []byte(`package main

import (
	"log"

	"github.com/espal-digital-development/espal-core/runner"
	core "github.com/espal-digital-development/espal-module-core"
	"github.com/juju/errors"
	_ "github.com/lib/pq"
)

func main() {
	app, err := runner.New()
	if err != nil {
		log.Fatal(errors.Trace(err))
	}

	coreModule, err := core.New()
	if err != nil {
		log.Fatal(err)
	}
	if err := app.RegisterModule(coreModule); err != nil {
		log.Fatal(err)
	}

	if err := app.Run(); err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
}
`)
	mainGoTestFile = []byte("package main_test\n")
)
