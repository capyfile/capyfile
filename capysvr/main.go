package main

import "flag"

var (
	name    = "capysvr"
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

	var healthCheck bool
	flag.BoolVar(&healthCheck, "health-check", false, "Enable the health check endpoint")
	flag.BoolVar(&healthCheck, "hc", false, "Enable the health check endpoint")

	var healthCheckEndpoint string
	flag.StringVar(&healthCheckEndpoint, "health-check-endpoint", "/health", "Health check endpoint")
	flag.StringVar(&healthCheckEndpoint, "hce", "/health", "Health check endpoint")

	var healthCheckEndpointVerbosity int
	flag.IntVar(&healthCheckEndpointVerbosity, "health-check-verbosity", 0, "Health check endpoint verbosity")
	flag.IntVar(&healthCheckEndpointVerbosity, "hcv", 0, "Health check endpoint verbosity")

	flag.Parse()

	server := &Server{
		Addr:                  ":8024",
		ServiceDefinitionFile: serviceDefinitionFile,
		Concurrency:           concurrency,
		ConcurrencyMode:       concurrencyMode,

		HealthCheck:                  healthCheck,
		HealthCheckEndpoint:          healthCheckEndpoint,
		HealthCheckEndpointVerbosity: healthCheckEndpointVerbosity,
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
