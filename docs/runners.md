# Runners

Runners defines the way how the pipeline is being run.

There are three types of runners:
* [capycmd](#capycmd) - command line interface to run the pipeline
* [capysvr](#capysvr) - HTTP server to run the pipeline for every request
* [capyworker](#capyworker) - worker to run the pipeline periodically

## capycmd

`capycmd` is a command line interface to run the pipeline.

### Usage

```bash
capycmd [-f <service-definition-file> -c] <service-processor>

Options:
	-f, --service-definition=<service-definition-file> Path to the service definition file
    -c, --concurrency Run the pipeline in the concurrent mode
```

### Examples

```bash
capycmd -f pipeline.images.yaml -c images:convert
```

## capysvr

`capysvr` is a HTTP server to run the pipeline for every request. By default, the server listens on :8024 address.

### Usage

```bash
capysvr [-f <service-definition-file> -p <port>] <service-processor>

Options:
    -f, --service-definition=<service-definition-file> Path to the service definition file
    -c, --concurrency Run the pipelines in the concurrent mode
```

### Examples

```bash
capysvr -f pipeline.images.yaml
```

## capyworker

`capyworker` is a worker to run the pipeline periodically. By default, the worker runs without any delay.

### Usage

```bash
capyworker [-f <service-definition-file> -s <sleep-time> -i <max-iterations> -l <log-file> -c] <service-processor>

Options:
    -f, --service-definition=<service-definition-file> Path to the service definition file
	-c, --concurrency Run the pipeline in the concurrency mode
    -s, --sleep=<sleep-time> Sleep time in seconds between each iteration of the worker
    -i, --max-iterations=<max-iterations> Maximum number of worker iterations
    -l, --log-file=<log-file> Log file
```

### Examples

```bash
capyworker -f pipeline.images.yaml -s 5 -i 1000 -l /var/log/images.capyfile.log -c images:convert
```
