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
	flag.BoolVar(&concurrency, "concurrency", false, "Use concurrency mode")
	flag.BoolVar(&concurrency, "c", false, "Run the pipeline in the concurrent mode")

	var concurrencyMode string
	flag.StringVar(&concurrencyMode, "concurrency-mode", "event", "Concurrency mode to use")
	flag.StringVar(&concurrencyMode, "m", "event", "Concurrency mode to use")

	var sleepTime int
	flag.IntVar(&sleepTime, "sleep", 0, "Sleep time in seconds between each iteration of the worker")
	flag.IntVar(&sleepTime, "s", 0, "Sleep time in seconds between each iteration of the worker")

	var maxIterations int
	flag.IntVar(&maxIterations, "max-iterations", 0, "Maximum number of worker iterations")
	flag.IntVar(&maxIterations, "i", 0, "Maximum number of worker iterations")

	var logFile string
	flag.StringVar(&logFile, "log-file", "", "Log file")
	flag.StringVar(&logFile, "l", "", "Log file")

	flag.Parse()

	args := flag.Args()

	var serviceProcessor string
	if len(args) == 1 {
		serviceProcessor = args[0]
	}

	if serviceProcessor == "" {
		fmt.Println(`capyworker - worker interface for Capyfile

Capyfile is a highly customizable file processing service that allows you to
define and run your own file processing pipelines.

Usage: [-f <service-definition-file>] <service-processor>

Options:
    -f, --service-definition=<service-definition-file> Path to the service definition file
    -c, --concurrency Run the pipeline in the concurrency mode
    -m, --concurrency-mode=<event|lock> Concurrency mode to use
    -s, --sleep=<sleep-time> Sleep time in seconds between each iteration of the worker
    -i, --max-iterations=<max-iterations> Maximum number of worker iterations
    -l, --log-file=<log-file> Log file

Examples:
    $ capyworker -f service-definition.yml sqs:consume
    $ capyworker --service-definition=service-definition.yml --sleep=5 sqs:consume

Version: ` + version + `
Commit: ` + commit + `
Date: ` + date + `
		`)

		return
	}

	worker := &Worker{
		ServiceDefinitionFile: serviceDefinitionFile,
		Concurrency:           concurrency,
		ConcurrencyMode:       concurrencyMode,
		SleepTime:             sleepTime,
		MaxIterations:         maxIterations,
		LogFile:               logFile,
	}

	initErr := worker.Init()
	if initErr != nil {
		panic(initErr)
	}

	runErr := worker.Run(serviceProcessor)
	if runErr != nil {
		panic(runErr)
	}
}
