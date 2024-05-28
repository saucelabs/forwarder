---
id: run
title: forwarder run
weight: 101
---

# Forwarder Run

Usage: `forwarder run [--address <host:port>] [--pac <path or url>] [--credentials <username:password@host:port>]... [flags]`

Start HTTP (forward) proxy server.
You can start HTTP or HTTPS server.
If you start an HTTPS server and you don't provide a certificate, the server will generate a self-signed certificate on startup.
The server may be protected by basic authentication.


**Note:** You can also specify the options as YAML, JSON or TOML file using `--config-file` flag.
You can generate a config file by running `forwarder run config-file` command.


## Examples

```
  # HTTP proxy with upstream proxy
  forwarder run --proxy http://localhost:8081

  # Start HTTP proxy with PAC script
  forwarder run --address localhost:3128 --pac https://example.com/pac.js

  # HTTPS proxy server with basic authentication
  forwarder run --protocol https --address localhost:8443 --basic-auth user:password

```

## Server options

### `--address` {#address}

* Environment variable: `FORWARDER_ADDRESS`
* Value Format: `<host:port>`
* Default value: `:3128`

The server address to listen on.
If the host is empty, the server will listen on all available interfaces.

### `--basic-auth` {#basic-auth}

* Environment variable: `FORWARDER_BASIC_AUTH`
* Value Format: `<username[:password]>`

Basic authentication credentials to protect the server.

### `-s, --credentials` {#credentials}

* Environment variable: `FORWARDER_CREDENTIALS`
* Value Format: `<username[:password]@host:port,...>`

Site or upstream proxy basic authentication credentials.
The host and port can be set to "*" to match all hosts and ports respectively.
The flag can be specified multiple times to add multiple credentials.

### `--idle-timeout` {#idle-timeout}

* Environment variable: `FORWARDER_IDLE_TIMEOUT`
* Value Format: `<duration>`
* Default value: `1h0m0s`

The maximum amount of time to wait for the next request before closing connection.

### `--name` {#name}

* Environment variable: `FORWARDER_NAME`
* Value Format: `<string>`
* Default value: `forwarder`

Name of this proxy instance.
This value is used in the Via header in requests.
The name value in Via header is extended with a random string to avoid collisions when several proxies are chained.

### `--protocol` {#protocol}

* Environment variable: `FORWARDER_PROTOCOL`
* Value Format: `<http|https>`
* Default value: `http`

The server protocol.
For https and h2 protocols, if TLS certificate is not specified, the server will use a self-signed certificate.

### `--read-header-timeout` {#read-header-timeout}

* Environment variable: `FORWARDER_READ_HEADER_TIMEOUT`
* Value Format: `<duration>`
* Default value: `1m0s`

The amount of time allowed to read request headers.

### `--read-limit` {#read-limit}

* Environment variable: `FORWARDER_READ_LIMIT`
* Value Format: `<bandwidth>`
* Default value: `0`

Global read rate limit in bytes per second i.e.
how many bytes per second you can receive from a proxy.
Accepts binary format (e.g.
1.5Ki, 1Mi, 3.6Gi).

### `--tls-cert-file` {#tls-cert-file}

* Environment variable: `FORWARDER_TLS_CERT_FILE`
* Value Format: `<path or base64>`

TLS certificate to use if the server protocol is https or h2.

Syntax:

- File: `/path/to/file.pac`
- Embed: `data:base64,<base64 encoded data>`

### `--tls-handshake-timeout` {#tls-handshake-timeout}

* Environment variable: `FORWARDER_TLS_HANDSHAKE_TIMEOUT`
* Value Format: `<duration>`
* Default value: `10s`

The maximum amount of time to wait for a TLS handshake before closing connection.
Zero means no limit.

### `--tls-key-file` {#tls-key-file}

* Environment variable: `FORWARDER_TLS_KEY_FILE`
* Value Format: `<path or base64>`

TLS private key to use if the server protocol is https or h2.

Syntax:

- File: `/path/to/file.pac`
- Embed: `data:base64,<base64 encoded data>`

### `--write-limit` {#write-limit}

* Environment variable: `FORWARDER_WRITE_LIMIT`
* Value Format: `<bandwidth>`
* Default value: `0`

Global write rate limit in bytes per second i.e.
how many bytes per second you can send to proxy.
Accepts binary format (e.g.
1.5Ki, 1Mi, 3.6Gi).

## Proxy options

### `--connect-header` {#connect-header}

* Environment variable: `FORWARDER_CONNECT_HEADER`
* Value Format: `<header>`

Add or remove CONNECT request headers.
See the documentation for the -H, --header flag for more details on the format.

### `--deny-domains` {#deny-domains}

* Environment variable: `FORWARDER_DENY_DOMAINS`
* Value Format: `[-]<regexp>,...`

Deny requests to the specified domains.
Prefix domains with '-' to exclude requests to certain domains from being denied.

### `--direct-domains` {#direct-domains}

* Environment variable: `FORWARDER_DIRECT_DOMAINS`
* Value Format: `[-]<regexp>,...`

Connect directly to the specified domains without using the upstream proxy.
Prefix domains with '-' to exclude requests to certain domains from being directed.
This flag takes precedence over the PAC script.

### `-H, --header` {#header}

* Environment variable: `FORWARDER_HEADER`
* Value Format: `<header>`

Add or remove HTTP request headers.
Use the format "name: value" to add a header, "name;" to set the header to empty value, "-name" to remove the header, "-name*" to remove headers by prefix.
The header name will be normalized to canonical form.
The header value should not contain any newlines or carriage returns.
The flag can be specified multiple times.
Example: -H "Host: example.com" -H "-User-Agent" -H "-X-*".

### `-p, --pac` {#pac}

* Environment variable: `FORWARDER_PAC`
* Value Format: `<path or URL>`

Proxy Auto-Configuration file to use for upstream proxy selection.

Syntax:

- File: `/path/to/file.pac`
- URL: `http://example.com/proxy.pac`
- Embed: `data:base64,<base64 encoded data>`
- Stdin: `-`

### `-x, --proxy` {#proxy}

* Environment variable: `FORWARDER_PROXY`
* Value Format: `<[protocol://]host:port>`

Upstream proxy to use.
The supported protocols are: http, https, socks5.
No protocol specified will be treated as HTTP proxy.
The basic authentication username and password can be specified in the host string e.g.
user:pass@host:port.
Alternatively, you can use the -c, --credentials flag to specify the credentials.
If both are specified, the proxy flag takes precedence.

### `--proxy-header` {#proxy-header}

* Environment variable: `FORWARDER_PROXY_HEADER`
* Value Format: `<header>`

DEPRECATED: use --connect-header flag instead

### `--proxy-localhost` {#proxy-localhost}

* Environment variable: `FORWARDER_PROXY_LOCALHOST`
* Value Format: `<allow|deny|direct>`
* Default value: `deny`

Setting this to allow enables sending requests to localhost through the upstream proxy.
Setting this to direct sends requests to localhost directly without using the upstream proxy.
By default, requests to localhost are denied.

### `-R, --response-header` {#response-header}

* Environment variable: `FORWARDER_RESPONSE_HEADER`
* Value Format: `<header>`

Add or remove HTTP headers on the received response before sending it to the client.
See the documentation for the -H, --header flag for more details on the format.

## MITM options

### `--mitm` {#mitm}

* Environment variable: `FORWARDER_MITM`
* Value Format: `<value>`
* Default value: `false`

Enable Man-in-the-Middle (MITM) mode.
It only works with HTTPS requests, HTTP/2 is not supported.
MITM is enabled by default when the --mitm-cacert-file flag is set.
If the CA certificate is not provided MITM uses a generated CA certificate.
The CA certificate used can be retrieved from the API server.

### `--mitm-cacert-file` {#mitm-cacert-file}

* Environment variable: `FORWARDER_MITM_CACERT_FILE`
* Value Format: `<path or base64>`

CA certificate file to use for generating MITM certificates.
If the file is not specified, a generated CA certificate will be used.
See the documentation for the --mitm flag for more details.

Syntax:

- File: `/path/to/file.pac`
- Embed: `data:base64,<base64 encoded data>`

### `--mitm-cakey-file` {#mitm-cakey-file}

* Environment variable: `FORWARDER_MITM_CAKEY_FILE`
* Value Format: `<path or base64>`

CA key file to use for generating MITM certificates.

### `--mitm-domains` {#mitm-domains}

* Environment variable: `FORWARDER_MITM_DOMAINS`
* Value Format: `[-]<regexp>,...`

Limit MITM to the specified domains.
Prefix domains with '-' to exclude requests to certain domains from being MITMed.

### `--mitm-org` {#mitm-org}

* Environment variable: `FORWARDER_MITM_ORG`
* Value Format: `<name>`
* Default value: `Forwarder Proxy MITM`

Organization name to use in the generated MITM certificates.

### `--mitm-validity` {#mitm-validity}

* Environment variable: `FORWARDER_MITM_VALIDITY`
* Value Format: `<duration>`
* Default value: `24h0m0s`

Validity period of the generated MITM certificates.

## DNS options

### `--dns-round-robin` {#dns-round-robin}

* Environment variable: `FORWARDER_DNS_ROUND_ROBIN`
* Value Format: `<value>`
* Default value: `false`

If more than one DNS server is specified with the --dns-server flag, passing this flag will enable round-robin selection.

### `-n, --dns-server` {#dns-server}

* Environment variable: `FORWARDER_DNS_SERVER`
* Value Format: `<ip>[:<port>]`

DNS server(s) to use instead of system default.
There are two execution policies, when more then one server is specified.
Fallback: the first server in a list is used as primary, the rest are used as fallbacks.
Round robin: the servers are used in a round-robin fashion.
The port is optional, if not specified the default port is 53.

### `--dns-timeout` {#dns-timeout}

* Environment variable: `FORWARDER_DNS_TIMEOUT`
* Value Format: `<duration>`
* Default value: `5s`

Timeout for dialing DNS servers.
Only used if DNS servers are specified.

## HTTP client options

### `--cacert-file` {#cacert-file}

* Environment variable: `FORWARDER_CACERT_FILE`
* Value Format: `<path or base64>`

Add your own CA certificates to verify against.
The system root certificates will be used in addition to any certificates in this list.
Use this flag multiple times to specify multiple CA certificate files.

Syntax:

- File: `/path/to/file.pac`
- Embed: `data:base64,<base64 encoded data>`

### `--http-dial-timeout` {#http-dial-timeout}

* Environment variable: `FORWARDER_HTTP_DIAL_TIMEOUT`
* Value Format: `<duration>`
* Default value: `30s`

The maximum amount of time a dial will wait for a connect to complete.
With or without a timeout, the operating system may impose its own earlier timeout.
For instance, TCP timeouts are often around 3 minutes.

### `--http-idle-conn-timeout` {#http-idle-conn-timeout}

* Environment variable: `FORWARDER_HTTP_IDLE_CONN_TIMEOUT`
* Value Format: `<duration>`
* Default value: `1m30s`

The maximum amount of time an idle (keep-alive) connection will remain idle before closing itself.
Zero means no limit.

### `--http-response-header-timeout` {#http-response-header-timeout}

* Environment variable: `FORWARDER_HTTP_RESPONSE_HEADER_TIMEOUT`
* Value Format: `<duration>`
* Default value: `0s`

The amount of time to wait for a server's response headers after fully writing the request (including its body, if any).This time does not include the time to read the response body.
Zero means no limit.

### `--http-tls-handshake-timeout` {#http-tls-handshake-timeout}

* Environment variable: `FORWARDER_HTTP_TLS_HANDSHAKE_TIMEOUT`
* Value Format: `<duration>`
* Default value: `10s`

The maximum amount of time waiting to wait for a TLS handshake.
Zero means no limit.

### `--insecure` {#insecure}

* Environment variable: `FORWARDER_INSECURE`
* Value Format: `<value>`
* Default value: `false`

Don't verify the server's certificate chain and host name.
Enable to work with self-signed certificates.

## API server options

### `--api-address` {#api-address}

* Environment variable: `FORWARDER_API_ADDRESS`
* Value Format: `<host:port>`
* Default value: `localhost:10000`

The server address to listen on.
If the host is empty, the server will listen on all available interfaces.

### `--api-basic-auth` {#api-basic-auth}

* Environment variable: `FORWARDER_API_BASIC_AUTH`
* Value Format: `<username[:password]>`

Basic authentication credentials to protect the server.

### `--api-idle-timeout` {#api-idle-timeout}

* Environment variable: `FORWARDER_API_IDLE_TIMEOUT`
* Value Format: `<duration>`
* Default value: `1h0m0s`

The maximum amount of time to wait for the next request before closing connection.

### `--api-read-header-timeout` {#api-read-header-timeout}

* Environment variable: `FORWARDER_API_READ_HEADER_TIMEOUT`
* Value Format: `<duration>`
* Default value: `1m0s`

The amount of time allowed to read request headers.

## Logging options

### `--log-file` {#log-file}

* Environment variable: `FORWARDER_LOG_FILE`
* Value Format: `<path>`

Path to the log file, if empty, logs to stdout.

### `--log-http` {#log-http}

* Environment variable: `FORWARDER_LOG_HTTP`
* Value Format: `[api|proxy:]<none|short-url|url|headers|body|errors>,...`
* Default value: `errors`

HTTP request and response logging mode.

Modes: 

- none: no logging
- short-url: logs [scheme://]host[/path] instead of the full URL
- url: logs the full URL including query parameters
- headers: logs request line and headers
- body: logs request line, headers, and body
- errors: logs request line and headers if status code is greater than or equal to 500

Modes for different modules can be specified separated by commas.
The following example specifies that the API module logs errors, the proxy module logs headers, and anything else logs full URL.

```
--log-http=api:errors,proxy:headers,url
```

### `--log-http-request-id-header` {#log-http-request-id-header}

* Environment variable: `FORWARDER_LOG_HTTP_REQUEST_ID_HEADER`
* Value Format: `<name>`
* Default value: `X-Request-Id`

If the header is present in the request, the proxy will associate the value with the request in the logs.

### `--log-level` {#log-level}

* Environment variable: `FORWARDER_LOG_LEVEL`
* Value Format: `<error|info|debug>`
* Default value: `info`

Log level.

