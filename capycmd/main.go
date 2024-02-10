package main

import (
	"flag"
	"fmt"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var serviceDefinitionFile string
	flag.StringVar(&serviceDefinitionFile, "service-definition", "", "Service definition file")
	flag.StringVar(&serviceDefinitionFile, "f", "", "Service definition file")

	var concurrency bool
	flag.BoolVar(&concurrency, "concurrency", false, "Run the pipeline in the concurrent mode")
	flag.BoolVar(&concurrency, "c", false, "Run the pipeline in the concurrent mode")

	var concurrencyMode string
	flag.StringVar(&concurrencyMode, "concurrency-mode", "event", "Concurrency mode to use")
	flag.StringVar(&concurrencyMode, "m", "event", "Concurrency mode to use")

	flag.Parse()

	args := flag.Args()

	var serviceProcessor string
	if len(args) == 1 {
		serviceProcessor = args[0]
	}

	if serviceProcessor == "" {
		fmt.Println(`capycmd - commandline interface for Capyfile

Capyfile is a highly customizable file processing service that allows you to
define and run your own file processing pipelines.

Usage: [-f <service-definition-file>] <service-processor>

Options:
    -f, --service-definition=<service-definition-file> Path to the service definition file
    -c, --concurrency Run the pipeline in the concurrent mode
    -m, --concurrency-mode=<event|lock> Concurrency mode to use

Examples:
    $ capycmd -f service-definition.yml logs:compress
    $ capycmd --service-definition=/etc/capyfile/logs.service-definition.json logs:archive
    $ CAPYFILE_SERVICE_DEFINITION_FILE=/etc/capyfile/images.service-definition.json capycmd images:clear_metadata
    $ CAPYFILE_SERVICE_DEFINITION_URL=https://example.com/photos.service-definition.json capycmd photos:upload

Version: ` + version + `
Commit: ` + commit + `
Date: ` + date + `
		`)

		return
	}

	cli := &Cli{
		ServiceDefinitionFile: serviceDefinitionFile,
		Concurrency:           concurrency,
		ConcurrencyMode:       concurrencyMode,
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
