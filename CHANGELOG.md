# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.2] 2024-01-21

### Added

- `input_forget` operation to forget the input
- parameter to remove original file for `filesystem_input_remove` operation

### Changed

- new documentation structure

## [1.2.1] 2024-01-01

### Added

- concurrency flag for all pipeline runners
- parameter to manage operation max packet size limit
- parameters to manage operation tick delays

### Fixed

- goroutines synchronization issue that could lead to empty final output being returned 

## [1.2.0] 2023-12-16

### Added

- `capyworker` to run the pipeline every N seconds
- More stable concurrency algorithm

## [1.1.4] 2023-11-25

### Added

- `capycmd` and `capysvr` now do the final cleanup

### Changed

- `commandArgs` parameter for `command_exec` operation is now optional

### Fixed

- Fix `capycmd` arguments parsing
- Fix cleanup inconsistency if there are original files should be preserved

## [1.1.3] 2023-11-22

### Added

- `command_exec` operation now can accept empty input
- `cleanupPolicy` operation parameter

### Changed

- Operations are now opening and closing the files when needed
- `exiftool_metadata_cleanup` now does not overwrite the original file

## [1.1.2] - 2023-11-18

### Added

- `command_exec` operation to execute commands

## [1.1.1] - 2023-11-12

### Added

- YAML pipeline config support

## [1.1.0] - 2023-11-11

### Added

- Concurrent operations support for `capysvr`
- New `http_*` operations to read HTTP input

### Changed

- `capysvr` now expects `http_*` operations to read HTTP input

### Fixed

- Concurrent operations no longer stuck when empty input is received
- `filesystem_input_write` operation now always reads the file from the beginning
- `filesystem_input_write` operation now uses correct file extension when if it was transformed