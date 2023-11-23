# Forwarder Proxy

[![Build Status](https://github.com/saucelabs/forwarder/actions/workflows/go.yml/badge.svg)](https://github.com/saucelabs/forwarder/actions/workflows/go.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/saucelabs/forwarder)](https://goreportcard.com/report/github.com/saucelabs/forwarder) [![GitHub release](https://img.shields.io/github/release/saucelabs/forwarder.svg)](https://github.com/saucelabs/forwarder/releases) ![GitHub all releases](https://img.shields.io/github/downloads/saucelabs/forwarder/total)

Forwarder is a forward HTTP proxy i.e. it supports `CONNECT` requests.

## It can proxy

* HTTP/HTTPS/HTTP2 requests
* WebSockets (both HTTP and HTTPS)
* Server Sent Events (SSE)
* TCP traffic (e.g. SMTP, IMAP, etc.)

## Features

* Supports upstream HTTP(S) and SOCKS5 proxies
* Supports PAC files for upstream proxy configuration
* Supports MITM for HTTPS traffic with automatic certificate generation
* Supports custom DNS servers
* Supports augmenting requests and responses with headers
* Supports basic authentication, for websites and proxies

## Running

You can get Forwarder build from GitHub releases page or a container from Docker Hub.

Forwarder provides the following commands:

- [forwarder run](cli/forwarder_run.md) - Start HTTP (forward) proxy server
- [forwarder pac eval](cli/forwarder_pac_eval.md) - Evaluate a PAC file for given URL (or URLs)
- [forwarder pac server](cli/forwarder_pac_server.md) - Start HTTP server that serves a PAC file
- [forwarder ready](cli/forwarder_ready.md) - Readiness probe for the Forwarder
