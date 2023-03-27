# 🚜 Forwarder [![Build Status](https://github.com/saucelabs/forwarder/actions/workflows/go.yml/badge.svg)](https://github.com/saucelabs/forwarder/actions/workflows/go.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/saucelabs/forwarder)](https://goreportcard.com/report/github.com/saucelabs/forwarder) [![GoDoc](https://godoc.org/github.com/saucelabs/forwarder?status.svg)](https://godoc.org/github.com/saucelabs/forwarder) [![GitHub release](https://img.shields.io/github/release/saucelabs/forwarder.svg)](https://github.com/saucelabs/forwarder/releases)

Forwarder is HTTP proxy.
It can be used to proxy HTTP/HTTPS/HTTP2 requests, Server Sent Events, WebSockets, TCP traffic and more.
It supports downstream HTTP(S) and SOCKS5 proxies and basic authentication.

## Installation

The release binaries are available on [GitHub](https://github.com/saucelabs/forwarder/releases).

You can also install Forwarder from source:

```go
go get -u github.com/saucelabs/forwarder/cmd/forwarder
```

## Usage

```text
HTTP (forward) proxy server with PAC support and PAC testing tools

Commands:
  run           Start HTTP (forward) proxy server
  pac           Tools for working with PAC files

Other Commands:
  completion    Generate the autocompletion script for the specified shell
  version       Print version information

The following options can be passed to any subcommand:

Options:
    --config-file <path> (env FORWARDER_CONFIG_FILE)
        Configuration file to load options from. The supported formats are: JSON, YAML, TOML, HCL, and Java
        properties. The file format is determined by the file extension, if not specified the default format is YAML.
        The following precedence order of configuration sources is used: command flags, environment variables, config
        file, default values.
```

### Run

```text
Start HTTP (forward) proxy server. You can start HTTP or HTTPS server. If you start an HTTPS server and you don't
provide a certificate, the server will generate a self-signed certificate on startup. The server may be protected by
basic authentication. Whenever applicable, username and password are URL decoded. This allows you to pass in special
characters such as @ by using %%40 or pass in a colon with %%3a.

Examples:
  # HTTP proxy with upstream proxy
  forwarder run --proxy http://localhost:8081

  # Start HTTP proxy with PAC script
  forwarder run --address localhost:3128 --pac https://example.com/pac.js

  # HTTPS proxy server with basic authentication
  forwarder run --protocol https --address localhost:8443 --basic-auth user:password

Options:
    --address <host:port> (default ':3128') (env FORWARDER_ADDRESS)
        The server address to listen on. If the host is empty, the server will listen on all available interfaces.

    --basic-auth <username:password> (env FORWARDER_BASIC_AUTH)
        Basic authentication credentials to protect the server. Username and password are URL decoded. This allows you
        to pass in special characters such as @ by using %40 or pass in a colon with %3a.

    -c, --credentials <username:password@host:port> (env FORWARDER_CREDENTIALS)
        Site or upstream proxy basic authentication credentials. The host and port can be set to "*" to match all
        hosts and ports respectively. The flag can be specified multiple times to add multiple credentials.

    -H, --header <header> (env FORWARDER_HEADER)
        Add or remove HTTP request headers. Use the format "name: value" to add a header, "name;" to set the header to
        empty value, "-name" to remove the header, "-name*" to remove headers by prefix. The header name will be
        normalized to canonical form. The header value should not contain any newlines or carriage returns. The flag
        can be specified multiple times. Example: -H "Host: example.com" -H "-User-Agent" -H "-X-*".

    -p, --pac <path or URL> (env FORWARDER_PAC)
        Proxy Auto-Configuration file to use for upstream proxy selection. It can be a local file or a URL, you can
        also use '-' to read from stdin.

    --protocol <http|https> (default http) (env FORWARDER_PROTOCOL)
        The server protocol. For https and h2 protocols, if TLS certificate is not specified, the server will use a
        self-signed certificate.

    -x, --proxy [protocol://]host[:port] (env FORWARDER_PROXY)
        Upstream proxy to use. The supported protocols are: http, https, socks, socks5. No protocol specified will be
        treated as HTTP proxy. If the port number is not specified, it is assumed to be 1080. The basic authentication
        username and password can be specified in the host string e.g. user:pass@host:port. Alternatively, you can use
        the -c, --credentials flag to specify the credentials.

    --proxy-localhost <allow|deny|direct> (default deny) (env FORWARDER_PROXY_LOCALHOST)
        Setting this to allow enables sending requests to localhost through the upstream proxy. Setting this to direct
        sends requests to localhost directly without using the upstream proxy. By default, requests to localhost are
        denied.

    --read-header-timeout <duration> (default 1m0s) (env FORWARDER_READ_HEADER_TIMEOUT)
        The amount of time allowed to read request headers.

    -R, --response-header <header> (env FORWARDER_RESPONSE_HEADER)
        Add or remove HTTP headers on the received response before sending it to the client. See the documentation for
        the -H, --header flag for more details on the format.

    --tls-cert-file <path> (env FORWARDER_TLS_CERT_FILE)
        TLS certificate to use if the server protocol is https or h2.

    --tls-key-file <path> (env FORWARDER_TLS_KEY_FILE)
        TLS private key to use if the server protocol is https or h2.

API server options:
    --api-address <host:port> (default 'localhost:10000') (env FORWARDER_API_ADDRESS)
        The server address to listen on. If the host is empty, the server will listen on all available interfaces.

    --api-basic-auth <username:password> (env FORWARDER_API_BASIC_AUTH)
        Basic authentication credentials to protect the server. Username and password are URL decoded. This allows you
        to pass in special characters such as @ by using %40 or pass in a colon with %3a.

    --api-log-http <none|url|headers|body|errors> (default errors) (env FORWARDER_API_LOG_HTTP)
        HTTP request and response logging mode. By default, request line and headers are logged if response status
        code is greater than or equal to 500. Setting this to none disables logging.

    --api-read-header-timeout <duration> (default 1m0s) (env FORWARDER_API_READ_HEADER_TIMEOUT)
        The amount of time allowed to read request headers.

DNS options:
    -n, --dns-server <ip>[:<port>] (env FORWARDER_DNS_SERVER)
        DNS server(s) to use instead of system default. If specified multiple times, the first one is used as primary
        server, the rest are used as fallbacks. The port is optional, if not specified the default port is 53.

    --dns-timeout <duration> (default 5s) (env FORWARDER_DNS_TIMEOUT)
        Timeout for dialing DNS servers.

HTTP client options:
    --http-dial-timeout <duration> (default 30s) (env FORWARDER_HTTP_DIAL_TIMEOUT)
        The maximum amount of time a dial will wait for a connect to complete. With or without a timeout, the
        operating system may impose its own earlier timeout. For instance, TCP timeouts are often around 3 minutes. 

    --http-idle-conn-timeout <duration> (default 1m30s) (env FORWARDER_HTTP_IDLE_CONN_TIMEOUT)
        The maximum amount of time an idle (keep-alive) connection will remain idle before closing itself. Zero means
        no limit. 

    --http-response-header-timeout <duration> (default 0s) (env FORWARDER_HTTP_RESPONSE_HEADER_TIMEOUT)
        The amount of time to wait for a server's response headers after fully writing the request (including its
        body, if any).This time does not include the time to read the response body. Zero means no limit. 

    --http-tls-handshake-timeout <duration> (default 10s) (env FORWARDER_HTTP_TLS_HANDSHAKE_TIMEOUT)
        The maximum amount of time waiting to wait for a TLS handshake. Zero means no limit.

    -k, --insecure <value> (default false) (env FORWARDER_INSECURE)
        Don't verify the server's certificate chain and host name. Enable to work with self-signed certificates. 

Logging options:
    --log-file <path> (env FORWARDER_LOG_FILE)
        Path to the log file, if empty, logs to stdout.

    --log-http <none|url|headers|body|errors> (default errors) (env FORWARDER_LOG_HTTP)
        HTTP request and response logging mode. By default, request line and headers are logged if response status
        code is greater than or equal to 500. Setting this to none disables logging.

    --log-level <error|info|debug> (default info) (env FORWARDER_LOG_LEVEL)
        Log level.

Usage:
  forwarder run [--address <host:port>] [--pac <path or url>] [--credentials <username:password@host:port>]... [flags]
```

## Proxy Auto-Configuration (PAC)

Forwarder implements [Proxy Auto-Configuration (PAC)](https://developer.mozilla.org/en-US/docs/Web/HTTP/Proxy_servers_and_tunneling/Proxy_Auto-Configuration_PAC_file) file support.
PAC file is a JavaScript file that returns a proxy URL for a given URL.
It can be used to implement complex proxy rules.

Forwarder also implements [Microsoft's PAC extensions for IPv6](https://learn.microsoft.com/en-us/windows/win32/winhttp/ipv6-aware-proxy-helper-api-definitions).
So you can use `FindProxyForURL` and `FindProxyForURLEx` functions to implement IPv6-aware proxy rules.

Forwarder can be used as a PAC file server (see `forwarder pac server` command), or as a PAC file test util (see `forwarder pac eval` command).

### Available functions

The following functions are available in PAC files:

- alert
- convert_addr
- dateRange
- dnsDomainIs
- dnsDomainLevels
- dnsResolve
- dnsResolveEx
- getClientVersion
- getDay
- getMonth
- isInNet
- isInNetEx
- isPlainHostName
- isResolvable
- isResolvableEx
- isValidIpAddress
- localHostOrDomainIs
- myIpAddress
- myIpAddressEx
- shExpMatch
- sortIpAddressList
- timeRange
- weekdayRange

## License

Forwarder is licensed under the [Mozilla Public License 2.0](https://www.mozilla.org/en-US/MPL/2.0/).
