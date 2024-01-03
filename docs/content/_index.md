+++
description = "Forwarder is a production-ready, fast MITM proxy with PAC support. It's suitable for debugging, intercepting and manipulating HTTP traffic. It's used as a core component of Sauce Labs Sauce Connect Proxy."
+++

# Forwarder Proxy

[![Build Status](https://github.com/saucelabs/forwarder/actions/workflows/go.yml/badge.svg)](https://github.com/saucelabs/forwarder/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/saucelabs/forwarder)](https://goreportcard.com/report/github.com/saucelabs/forwarder)
[![GitHub Repo stars](https://img.shields.io/github/stars/saucelabs/forwarder)](https://github.com/saucelabs/forwarder/)
[![GitHub release](https://img.shields.io/github/release/saucelabs/forwarder.svg)](https://github.com/saucelabs/forwarder/releases)

Forwarder is a production-ready, fast MITM proxy with PAC support.
It's suitable for debugging, intercepting and manipulating HTTP traffic.
It's used as a core component of Sauce Labs [Sauce Connect Proxy](https://docs.saucelabs.com/secure-connections/sauce-connect/).
It is a forward proxy, which means it proxies traffic from clients to servers (e.g. browsers to websites), and supports `CONNECT` requests.
It can proxy:

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

See the Install instructions for your platform or use the Docker image.
When you have Forwarder installed, you can run it with the following command:

- [forwarder run](cli/forwarder_run.md) - Start HTTP (forward) proxy server
- [forwarder pac eval](cli/forwarder_pac_eval.md) - Evaluate a PAC file for given URL (or URLs)
- [forwarder pac server](cli/forwarder_pac_server.md) - Start HTTP server that serves a PAC file
- [forwarder ready](cli/forwarder_ready.md) - Readiness probe for the Forwarder

## Asking for help

If you have any questions about Forwarder, please feel free to ask them on the
[Forwarder Discussions](https://github.com/saucelabs/forwarder/discussions) page.