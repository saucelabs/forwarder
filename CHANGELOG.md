# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Roadmap

- Abstract the underlying proxy implementation - currently bound to GoProxy
- Load config from Config file, and from env vars. Use viper for that

## [0.1.3] - 2021-10-15
### Added
- Ability to route `localhost`/`127.0.0.1` requests thru any upstream proxy - if any.

## [0.1.2] - 2021-10-13
### Changed
- Updated dependencies.

## [0.1.1] - 2021-10-4
### Added
- Integrated PAC (Pacman).
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
