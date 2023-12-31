package main

import "flag"

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

	flag.Parse()

	server := &Server{
		Addr:                  ":8024",
		ServiceDefinitionFile: serviceDefinitionFile,
		Concurrency:           concurrency,
	}

	serverInitErr := server.Init()
	if serverInitErr != nil {
		panic(serverInitErr)
	}

	serverRunErr := server.Run()
	if serverRunErr != nil {
		panic(serverRunErr)
	}
}
