package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var serviceProcessor string
	if len(os.Args) > 1 {
		serviceProcessor = os.Args[1]
	}

	if serviceProcessor == "" {
		fmt.Println(`capycmd - commandline interface for Capyfile

Capyfile is a highly customizable file processing service that allows you to
define and run your own file processing pipelines.

Usage: <service-processor> [--service-definition=<service-definition-file>]

Options:
	--service-definition=<service-definition-file>    Path to the service definition file

Examples:
    $ capycmd logs:archive --service-definition=/etc/capyfile/logs.service-definition.json
    $ CAPYFILE_SERVICE_DEFINITION_FILE=/etc/capyfile/images.service-definition.json capycmd images:clear_metadata
    $ CAPYFILE_SERVICE_DEFINITION_URL=https://example.com/photos.service-definition.json capycmd photos:upload

Version: ` + version + `
Commit: ` + commit + `
Date: ` + date + `
		`)

		return
	}

	var serviceDefinitionFile string
	flag.StringVar(&serviceDefinitionFile, "service-definition", "", "Service definition file")

	cli := &Cli{
		ServiceDefinitionFile: serviceDefinitionFile,
	}

	initErr := cli.Init()
	if initErr != nil {
		panic(initErr)
	}

	runErr := cli.Run(serviceProcessor)
	if runErr != nil {
		panic(runErr)
	}
}
