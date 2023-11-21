# forwarder run

Usage: `forwarder run [--address <host:port>] [--pac <path or url>] [--credentials <username:password@host:port>]... [flags]`

Start HTTP (forward) proxy server.
You can start HTTP or HTTPS server.
If you start an HTTPS server and you don't provide a certificate, the server will generate a self-signed certificate on startup.
The server may be protected by basic authentication.


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

### `--address`

Environment variable: `FORWARDER_ADDRESS`

The server address to listen on.
If the host is empty, the server will listen on all available interfaces.

Default value: `:3128`

### `--basic-auth`

Environment variable: `FORWARDER_BASIC_AUTH`

Basic authentication credentials to protect the server.

### `-s, --credentials`

Environment variable: `FORWARDER_CREDENTIALS`

Site or upstream proxy basic authentication credentials.
The host and port can be set to "*" to match all hosts and ports respectively.
The flag can be specified multiple times to add multiple credentials.

### `--name`

Environment variable: `FORWARDER_NAME`

Name of this proxy instance.
This value is used in the Via header in requests.
The name value in Via header is extended with a random string to avoid collisions when several proxies are chained.

Default value: `forwarder`

### `--protocol`

Environment variable: `FORWARDER_PROTOCOL`

The server protocol.
For https and h2 protocols, if TLS certificate is not specified, the server will use a self-signed certificate.

Default value: `http`

### `--read-header-timeout`

Environment variable: `FORWARDER_READ_HEADER_TIMEOUT`

The amount of time allowed to read request headers.

Default value: `1m0s`

### `--read-limit`

Environment variable: `FORWARDER_READ_LIMIT`

Global read rate limit in bytes per second i.e.
how many bytes per second you can receive from a proxy.
Accepts binary format (e.g.
1.5Ki, 1Mi, 3.6Gi).

Default value: `0`

### `--tls-cert-file`

Environment variable: `FORWARDER_TLS_CERT_FILE`

TLS certificate to use if the server protocol is https or h2.
Can be a path to a file or "data:" followed by base64 encoded certificate.

### `--tls-key-file`

Environment variable: `FORWARDER_TLS_KEY_FILE`

TLS private key to use if the server protocol is https or h2.
Can be a path to a file or "data:" followed by base64 encoded key.

### `--write-limit`

Environment variable: `FORWARDER_WRITE_LIMIT`

Global write rate limit in bytes per second i.e.
how many bytes per second you can send to proxy.
Accepts binary format (e.g.
1.5Ki, 1Mi, 3.6Gi).

Default value: `0`

## Proxy options

### `--deny-domains`

Environment variable: `FORWARDER_DENY_DOMAINS`

Deny requests to the specified domains.
Prefix domains with '-' to exclude requests to certain domains from being denied.

### `--direct-domains`

Environment variable: `FORWARDER_DIRECT_DOMAINS`

Connect directly to the specified domains without using the upstream proxy.
Prefix domains with '-' to exclude requests to certain domains from being directed.
This flag takes precedence over the PAC script.

### `-H, --header`

Environment variable: `FORWARDER_HEADER`

Add or remove HTTP request headers.
Use the format "name: value" to add a header, "name;" to set the header to empty value, "-name" to remove the header, "-name*" to remove headers by prefix.
The header name will be normalized to canonical form.
The header value should not contain any newlines or carriage returns.
The flag can be specified multiple times.
Example: -H "Host: example.com" -H "-User-Agent" -H "-X-*".

### `-p, --pac`

Environment variable: `FORWARDER_PAC`

Proxy Auto-Configuration file to use for upstream proxy selection.
It can be a local file or a URL, you can also use '-' to read from stdin.
The data URI scheme is supported, the format is data:base64,<encoded data>.

### `-x, --proxy`

Environment variable: `FORWARDER_PROXY`

Upstream proxy to use.
The supported protocols are: http, https, socks5.
No protocol specified will be treated as HTTP proxy.
If the port number is not specified, it is assumed to be 1080.
The basic authentication username and password can be specified in the host string e.g.
user:pass@host:port.
Alternatively, you can use the -c, --credentials flag to specify the credentials.
If both are specified, the proxy flag takes precedence.

### `--proxy-header`

Environment variable: `FORWARDER_PROXY_HEADER`

Add or remove HTTP headers on the CONNECT request to the upstream proxy.
See the documentation for the -H, --header flag for more details on the format.

### `--proxy-localhost`

Environment variable: `FORWARDER_PROXY_LOCALHOST`

Setting this to allow enables sending requests to localhost through the upstream proxy.
Setting this to direct sends requests to localhost directly without using the upstream proxy.
By default, requests to localhost are denied.

Default value: `deny`

### `-R, --response-header`

Environment variable: `FORWARDER_RESPONSE_HEADER`

Add or remove HTTP headers on the received response before sending it to the client.
See the documentation for the -H, --header flag for more details on the format.

## MITM options

### `--mitm`

Environment variable: `FORWARDER_MITM`

Enable Man-in-the-Middle (MITM) mode.
It only works with HTTPS requests, HTTP/2 is not supported.
MITM is enabled by default when the --mitm-cacert-file flag is set.
If the CA certificate is not provided MITM uses a generated CA certificate.
The CA certificate used can be retrieved from the API server .

Default value: `false`

### `--mitm-cacert-file`

Environment variable: `FORWARDER_MITM_CACERT_FILE`

CA certificate file to use for generating MITM certificates.
If the file is not specified, a generated CA certificate will be used.
See the documentation for the --mitm flag for more details.

### `--mitm-cakey-file`

Environment variable: `FORWARDER_MITM_CAKEY_FILE`

CA key file to use for generating MITM certificates.

### `--mitm-domains`

Environment variable: `FORWARDER_MITM_DOMAINS`

Limit MITM to the specified domains.
Prefix domains with '-' to exclude requests to certain domains from being MITMed.

### `--mitm-org`

Environment variable: `FORWARDER_MITM_ORG`

Organization name to use in the generated MITM certificates.

Default value: `Sauce Labs Inc.`

### `--mitm-validity`

Environment variable: `FORWARDER_MITM_VALIDITY`

Validity period of the generated MITM certificates.


Default value: `24h0m0s`

## DNS options

### `--dns-round-robin`

Environment variable: `FORWARDER_DNS_ROUND_ROBIN`

If more than one DNS server is specified with the --dns-server flag, passing this flag will enable round-robin selection.


Default value: `false`

### `-n, --dns-server`

Environment variable: `FORWARDER_DNS_SERVER`

DNS server(s) to use instead of system default.
There are two execution policies, when more then one server is specified.
Fallback: the first server in a list is used as primary, the rest are used as fallbacks.
Round robin: the servers are used in a round-robin fashion.
The port is optional, if not specified the default port is 53.

### `--dns-timeout`

Environment variable: `FORWARDER_DNS_TIMEOUT`

Timeout for dialing DNS servers.
Only used if DNS servers are specified.


Default value: `5s`

## HTTP client options

### `--cacert-file`

Environment variable: `FORWARDER_CACERT_FILE`

Add your own CA certificates to verify against.
The system root certificates will be used in addition to any certificates in this list.
Can be a path to a file or "data:" followed by base64 encoded certificate.
Use this flag multiple times to specify multiple CA certificate files.

### `--http-dial-timeout`

Environment variable: `FORWARDER_HTTP_DIAL_TIMEOUT`

The maximum amount of time a dial will wait for a connect to complete.
With or without a timeout, the operating system may impose its own earlier timeout.
For instance, TCP timeouts are often around 3 minutes.


Default value: `10s`

### `--http-idle-conn-timeout`

Environment variable: `FORWARDER_HTTP_IDLE_CONN_TIMEOUT`

The maximum amount of time an idle (keep-alive) connection will remain idle before closing itself.
Zero means no limit.


Default value: `1m30s`

### `--http-response-header-timeout`

Environment variable: `FORWARDER_HTTP_RESPONSE_HEADER_TIMEOUT`

The amount of time to wait for a server's response headers after fully writing the request (including its body, if any).This time does not include the time to read the response body.
Zero means no limit.


Default value: `0s`

### `--http-tls-handshake-timeout`

Environment variable: `FORWARDER_HTTP_TLS_HANDSHAKE_TIMEOUT`

The maximum amount of time waiting to wait for a TLS handshake.
Zero means no limit.

Default value: `10s`

### `--insecure`

Environment variable: `FORWARDER_INSECURE`

Don't verify the server's certificate chain and host name.
Enable to work with self-signed certificates.


Default value: `false`

## API server options

### `--api-address`

Environment variable: `FORWARDER_API_ADDRESS`

The server address to listen on.
If the host is empty, the server will listen on all available interfaces.

Default value: `localhost:10000`

### `--api-basic-auth`

Environment variable: `FORWARDER_API_BASIC_AUTH`

Basic authentication credentials to protect the server.

### `--api-read-header-timeout`

Environment variable: `FORWARDER_API_READ_HEADER_TIMEOUT`

The amount of time allowed to read request headers.

Default value: `1m0s`

### `--prom-namespace`

Environment variable: `FORWARDER_PROM_NAMESPACE`

Prometheus namespace to use for metrics.
The metrics are available at /metrics endpoint in the API server.

Default value: `forwarder`

## Logging options

### `--log-file`

Environment variable: `FORWARDER_LOG_FILE`

Path to the log file, if empty, logs to stdout.

### `--log-http`

Environment variable: `FORWARDER_LOG_HTTP`

HTTP request and response logging mode.
Setting this to none disables logging.
The short-url mode logs [scheme://]host[/path] instead of the full URL.
The error mode logs request line and headers if status code is greater than or equal to 500.

### `--log-http-request-id-header`

Environment variable: `FORWARDER_LOG_HTTP_REQUEST_ID_HEADER`

If the header is present in the request, the proxy will associate the value with the request in the logs.

Default value: `X-Request-Id`

### `--log-level`

Environment variable: `FORWARDER_LOG_LEVEL`

Log level.

Default value: `info`

