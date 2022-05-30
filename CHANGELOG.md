# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Roadmap

- Abstract the underlying proxy implementation - currently bound to GoProxy
- Load config from Config file, and from env vars. Use viper for that
- Automatically alocates a random port, if the specified one is in-use

## [0.2.0] - 2022-05-30
## Changed
- Upgraded goproxy library to the latest master
- Upgraded pacman
- Renamed FORWARDER_*_CREDENTIAL to FORWARDER_*_AUTH env vars

## [0.1.20] - 2022-04-21
## Changed
- Upgrade goproxy library to the latest master
- Check that the response struct is allocated before logging its details

## [0.1.19] - 2022-04-18
## Changed
- Revert the changes introduced in 0.1.18
- Log request and response at DEBUG level

## [0.1.18] - 2022-04-13
## Changed
- Redirect standard output to configured logger

## [0.1.17] - 2022-04-12
## Changed
- Log request and response at INFO level
- Log request headers at DEBUG level

## [0.1.16] - 2022-04-11
## Changed
- Reset upstream proxy configuration when PAC is used and a URL doesn't require proxy
- Upgraded golang-ci version (CI pipeline)

## [0.1.15] - 2022-02-22
## Changed
- Fix message logging level

## [0.1.14] - 2022-02-22
## Changed
- Updated dependencies

## [0.1.13] - 2022-02-14
### Changed
- Upgraded PACMan version
- Upgraded golang-ci version (CI pipeline)
- Standardized Go version to 1.16

## [0.1.12] - 2021-11-02
### Changed
- Better error message

## [0.1.11] - 2021-11-02
### Added
- Added support to specify multiple DNS server.

## [0.1.10] - 2021-11-01
### Added
- Added support to specify the DNS server.
- Added `dnsURI` the DNS URI custom validator.

### Changed
- Fix bug in `ProxyLogger.Printf` where `v` wasn't passed as a slice.
- Rename PAC `textOrURI` terminology to `source`.

## [0.1.9] - 2021-10-29
### Added
- Added the ability to specify an external logger.
- Added state which properly handles multiple calls to `Run`.
- Added `ProxyLogger` which is passed to the underlying proxy implementation (`GoProxy`) to format its message accordingly to Forwarder logger format.
- Proper handles cases where `Run` may be called without proxy being setup.
- Covered all changes with tests.

### Changed
- Upgraded to Sypl v1.5.5

## [0.1.8] - 2021-10-27
### Changed
- Removed unused/leftover file.

## [0.1.7] - 2021-10-27
### Added
- Added more tests covering passing credentials via env var.

## [0.1.6] - 2021-10-27
### Changed
- Valid proxy schemes enum are simple strings.

## [0.1.5] - 2021-10-27
### Added
- More logging.

### Changed
- Upgraded PACMan dependency.

## [0.1.4] - 2021-10-19
### Added
- Added valid proxy schemes enum.
- Added the ability to check if the specified local proxy port is available. If not, it'll automatically allocate a set of new ones and test each until it finds an available one. If the pool is exhausted, it fails.

## [0.1.3] - 2021-10-15
### Added
- Ability to route `localhost`/`127.0.0.1` requests thru any upstream proxy - if any.

## [0.1.2] - 2021-10-13
### Changed
- Updated dependencies.

## [0.1.1] - 2021-10-4
### Added
- Integrated PAC (PACMan).
- More tests.
- Added benchmark.

## [0.1.0] - 2021-09-21
### Added
- [x] Ability to proxy connections.
- [x] Ability to protect local proxy with basic auth.
- [x] Ability to forward connections to a parent proxy.
- [x] Ability to forward connections to a parent proxy protected with basic auth.

### Checklist

- [x] CI Pipeline:
  - [x] Lint
  - [x] Tests
  - [x] Integration tests
  - [x] Automatic release (git tag)
- [x] Proxy as a Package.
- [x] Proxy as a CLI:
  - [x] Command has `short` description
  - [x] Command has `long` description
  - [x] Command has `example`
  - [x] Command flags have meaningful, clear names, and when needed - short names
  - [x] Print version with injected information from pipeline such as commit, build data, and tag version
- [x] Extensive logging options:
  - [x] Logging level
  - [x] Log file name
  - [x] Log file logging level
  - [x] Print logging setup information at `debug` level such as path to the filename
- [x] Documentation:
  - [x] Package's documentation (`doc.go`)
  - [x] Meaningful code comments, and symbol names (`const`, `var`, `func`)
  - [x] `GoDoc` server tested
  - [x] `README.md`
  - [x] `LICENSE`
    - [x] Files has LICENSE in the header
  - [x] Useful `CHANGELOG.md`
  - [x] Clear `CONTRIBUTION.md`
- Automation:
  - [x] `Makefile`
  - [x] Watch for changes - hot reload, build binary at `bin/forwarder` (`make dev`)
- Testing:
  - [x] Coverage 80%+
  - [x] Unit test
  - [x] Integration test
  - [x] Real testing
- Examples:
  - [x] Example's test file
- Errors:
  - [x] Consistent, and standardized errors (powered by `CustomError`)
- Logging:
  - [x] Consistent, and standardized logging (powered by `Sypl`)
  - [x] Output to `stdout`
  - [x] Output to `stderr`
  - [x] Output to file
