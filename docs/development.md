## Core concepts

On paper, it supposed to look like this:

![concept diagram](../diagram.drawio.png)

There are three core concepts:
* **Service**. Top layer that has access to the widest context.
* **Processor**. It's responsible for configuring the operations, building operation
  pipeline.
* **Operation**. Do the actual file processing. It can read, write, validate, modify, or store the
  files.

## Development

Capyfile is an open source project so everyone is welcomed to contribute.

### How to create a new operation

There are few things you need to remember:
* Operation should return error only if it is a critical error that does not allow the operation
  to function properly. For example, incorrect configuration, missing dependencies, etc.
* The processable files should not disappear unless you remove the file associated with
  it. Use `pf.ReplaceFile()` method if the file was modified by the operation. This way
  Capyfile can track this change and do proper cleanup if necessary.

### How to run the development environment

What we have so far is a basic dev environment running on Docker.

For the development purposes, we have a docker-compose file with all necessary dependencies
(see `docker-compose.dev.yml`).

Also, we have three service definitions for `capysvr`,`capycmd`, and `capyworker` services:
* `service-definition.capysvr.dev.yml` - prepared service definition for `capysvr`
* `service-definition.capycmd.dev.yml` - prepared service definition for `capycmd`
* `service-definition.capyworker.dev.yml` - prepared service definition for `capyworker`

And the `dev.sh` script that helps to build, run, and stop the services.

What is available for `capysvr`:
```bash
# Build capysvr from the source code and run it with all necessary dependencies
./dev.sh start capysvr

# now capysvr is accessible on http://localhost:8024 or http://capyfile.local:8024
# it use `service-definition.dev.yml` service definition file

# If you have made some changes in the source code, you can rebuild the capysvr
./dev.sh rebuild capysvr

# Stop the capysvr
./dev.sh stop capysvr
```

What is available for `capycmd`:
```bash
# Build capycmd from the source code and run it with all necessary dependencies
# This will open the container's shell where you have access to ./capycmd command
./dev.sh start capycmd
~$ ./capycmd logs:archive

# If you have made some changes in the source code, you can rebuild the capycmd
./dev.sh rebuild capycmd

# Stop the capycmd
./dev.sh stop capycmd
```

And the same stuff is available for `capyworker`:
```bash
# Build capyworker from the source code and run it with all necessary dependencies
# This will open the container's shell where you have access to ./capyworker command
./dev.sh start capyworker
~$ ./capyworker photos:upload

# If you have made some changes in the source code, you can rebuild the capyworker
./dev.sh rebuild capyworker

# Stop the capyworker
./dev.sh stop capyworker
```
