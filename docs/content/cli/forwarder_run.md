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

### `--proxy-protocol-listener` {#proxy-protocol-listener}

* Environment variable: `FORWARDER_PROXY_PROTOCOL_LISTENER`
* Value Format: `<value>`
* Default value: `false`

The PROXY protocol is used to correctly read the client's IP address.
When enabled the proxy will expect the client to send the PROXY protocol header before the actual request.
PROXY protocol version 1 and 2 are supported.

### `--proxy-protocol-read-header-timeout` {#proxy-protocol-read-header-timeout}

* Environment variable: `FORWARDER_PROXY_PROTOCOL_READ_HEADER_TIMEOUT`
* Value Format: `<duration>`
* Default value: `5s`

The amount of time to wait for PROXY protocol header.
Zero means no limit.

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

### `--shutdown-timeout` {#shutdown-timeout}

* Environment variable: `FORWARDER_SHUTDOWN_TIMEOUT`
* Value Format: `<duration>`
* Default value: `30s`

The maximum amount of time to wait for the server to drain connections before closing.
Zero means no limit.

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

Use the format:

- name:value to add a header
- name; to set the header to empty value
- -name to remove the header
- -name* to remove headers by prefix

The header name will be normalized to canonical form.
The header value should not contain any newlines or carriage returns.
The flag can be specified multiple times.
The following example removes the User-Agent header and all headers starting with X-.

```
-H "-User-Agent" -H "-X-*"
```

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

### `--mitm-cache-size` {#mitm-cache-size}

* Environment variable: `FORWARDER_MITM_CACHE_SIZE`
* Value Format: `<size>`
* Default value: `1024`

Maximum number of certificates to cache.
If the cache is full, the least recently used certificate is removed.

### `--mitm-cache-ttl` {#mitm-cache-ttl}

* Environment variable: `FORWARDER_MITM_CACHE_TTL`
* Value Format: `<duration>`
* Default value: `6h0m0s`

Expiration time of the cached certificates.

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

### `--connect-to` {#connect-to}

* Environment variable: `FORWARDER_CONNECT_TO`
* Value Format: `<HOST1:PORT1:HOST2:PORT2>,...`

For a request to the given HOST1:PORT1 pair, connect to HOST2:PORT2 instead.
This option is suitable to direct requests at a specific server, e.g.
at a specific cluster node in a cluster of servers.
This option is only used to establish the network connection and does not work when request is routed using an upstream proxy.
It does NOT affect the hostname/port that is used for TLS/SSL (e.g.
SNI, certificate verification) or for the application protocols.
HOST1 and PORT1 may be the empty string, meaning any host/port.
HOST2 and PORT2 may also be the empty string, meaning use the request's original host/port.

### `--http-dial-attempts` {#http-dial-attempts}

* Environment variable: `FORWARDER_HTTP_DIAL_ATTEMPTS`
* Value Format: `<int>`
* Default value: `3`

The number of attempts to dial the network address.

### `--http-dial-backoff` {#http-dial-backoff}

* Environment variable: `FORWARDER_HTTP_DIAL_BACKOFF`
* Value Format: `<duration>`
* Default value: `1s`

The amount of time to wait between dial attempts.

### `--http-dial-timeout` {#http-dial-timeout}

* Environment variable: `FORWARDER_HTTP_DIAL_TIMEOUT`
* Value Format: `<duration>`
* Default value: `25s`

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

### `--http-tls-keylog-file` {#http-tls-keylog-file}

* Environment variable: `FORWARDER_HTTP_TLS_KEYLOG_FILE`
* Value Format: `<path>`

File to log TLS master secrets in NSS key log format.
By default, the value is taken from the SSLKEYLOGFILE environment variable.
It can be used to allow external programs such as Wireshark to decrypt TLS connections.

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

### `--api-read-limit` {#api-read-limit}

* Environment variable: `FORWARDER_API_READ_LIMIT`
* Value Format: `<bandwidth>`
* Default value: `0`

Global read rate limit in bytes per second i.e.
how many bytes per second you can receive from a proxy.
Accepts binary format (e.g.
1.5Ki, 1Mi, 3.6Gi).

### `--api-shutdown-timeout` {#api-shutdown-timeout}

* Environment variable: `FORWARDER_API_SHUTDOWN_TIMEOUT`
* Value Format: `<duration>`
* Default value: `30s`

The maximum amount of time to wait for the server to drain connections before closing.
Zero means no limit.

### `--api-write-limit` {#api-write-limit}

* Environment variable: `FORWARDER_API_WRITE_LIMIT`
* Value Format: `<bandwidth>`
* Default value: `0`

Global write rate limit in bytes per second i.e.
how many bytes per second you can send to proxy.
Accepts binary format (e.g.
1.5Ki, 1Mi, 3.6Gi).

## Logging options

### `--log-file` {#log-file}

* Environment variable: `FORWARDER_LOG_FILE`
* Value Format: `<path>`

Path to the log file, if empty, logs to stdout.
The file is reopened on SIGHUP to allow log rotation using external tools.

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


## Time-based access control

Forwarder allows to impose time restrictions on its operations. User can configure day of week + hour ranges when traffic forwarding is allowed.
Outside specified time ranges, all outgoing HTTP/HTTPS requests will be denied using HTTP 451 error code.

All hour ranges are specified in local time zone specific for the machine running forwarder. Hours are specified using 24h format.

If there is no time frame specified - no time-based access control will be enforced.

Limitations: Access control is enforced only on newly opened connections. Long lasting sessions opened during allowed time-frame will not be forcibly closed if allowed time frame ends.

### `--allow-time-frame` {#allow-time-frame}

* Environment variable: `FORWARDER_ALLOW_TIME_FRAME`
* Value Format: `{day-of-week}/{hour_start}-{hour_end},...`

Examples of value format:

* allow forwarding on Mondays between 11:00 and 13:00: `mon/11-13`
* allow forwarding on Tuesdays and Fridays between 18:00 and 23:59:59: `tue/18-24,fri/18-24`
* allow forwarding on Mondays between 00:00 and 23:59:59: `mon/00-24`

## Kerberos options

Forwarder supports Kerberos authentication - SPNEGO tokens passed in `Authorization` or `Proxy-Authorization` HTTP headers. Implemented method is called "opportunistic authentication", it means forwarder does not try to detect `401` or `407` HTTP error codes and negotiate Kerberos authentication - it uses predefined host names needing Kerberos authentication and it's up to the user to know those hosts in advance.

Forwarder can authenticate using Kerberos both to the configured upstream proxy (using `Proxy-Authorization` header) and inject `Authorization` header to the proxied requests - by fetching relevant Kerberos token for particular SPN (Service Principal Name) and converting it to SPNEGO token.

Current implementation uses HTTP request host (or configured proxy hostname) as SPN. For example `app.example.com` is converted to SPN `HTTP/app.example.com` and such SPN is expected to be present in Kerberos KDC server as service name.

To use Kerberos authentication mechanism you need to have `krb5.conf` file which points to proper realms and servers and keytab files accessible to forwarder.
Below are links to sample files configured for `example.com` domain:

* [krb5.conf](../config/kerberos/krb5.conf)
* [keytab](../config/kerberos/keytab) 

Keytab is a binary file storing saved usernames (principals) and passwords used to authenticate against Kerberos KDC server so there is no need for a password to be passed as command line arguments in forwarder. Sample file contains two entries (user1 and user2) with password `password`.

To create a custom keytab/keytab entry run:

```
root@host:/etc/krb5kdc# ktutil
ktutil:  add_entry -password -p user3@example.com -k 1 -e aes256-cts-hmac-sha1-96
Password for user3@example.com: 
ktutil: write_kt keytab
ktutil: exit
```

file will be saved as `keytab` in current directory.
(ktutil is often part of `krb5-user` linux package)

### Important! 

Keytab file needs separate user entry for each supported client encryption type, otherwise if encryption type entry negotiated between client
and KDC server is not found in keytab for particular user - forwarder will fail with following error:

```
msg="fatal error exiting" error="kerberos KDC login: [Root cause: Encrypting_Error] KRBMessage_Handling_Error: AS Exchange Error: failed setting AS_REQ PAData for pre-authentication required < Encrypting_Error: error getting key from credentials: matching key not found in keytab. Looking for \"user3\" realm: example.com kvno: 0 etype: 18"
```

### --kerberos-cfg-file {#kerberos-cfg-file}

* Environment variable: `FORWARDER_KERBEROS_CFG_FILE`
* Value Format: `<path>`

Path to krb5.conf configuration file with kerberos connection settings. 
File format reference:

https://web.mit.edu/kerberos/krb5-1.12/doc/admin/conf_files/krb5_conf.html

See this file for example: [krb5.conf](../config/kerberos/krb5.conf)


### --kerberos-keytab-file {#kerberos-cfg-file}

* Environment variable: `FORWARDER_KERBEROS_KEYTAB_FILE`
* Value Format: `<path>`

Path to keytab file holding credentials to a user which authenticates to a Kerberos KDC server.

Keytab files are in binary format and can be created and managed for example using `ktutil` tool distributed with MIT Kerberos software:
https://web.mit.edu/kerberos/krb5-latest/doc/admin/admin_commands/ktutil.html#ktutil-1


### --kerberos-user-name {#kerberos-user-name}

* Environment variable: `FORWARDER_KERBEROS_USER_NAME`
* Value Format: `<username>`

Name of the user (principal name using Kerberos nomenclature) as which forwarder will authenticate to a Kerberos KDC server. User and its password (hashed) must be present in keytab file specified in `--kerberos-keytab-file`

### --kerberos-user-realm {#kerberos-user-realm}

* Environment variable: `FORWARDER_KERBEROS_USER_REALM`
* Value Format: `<domain name>`

Kerberos realm of the user specified in `--kerberos-user-name`. It depends on Kerberos settings in organisation but in most cases it's the company domain name.


### --kerberos-enabled-hosts {#kerberos-enabled-hosts}

* Environment variable: `FORWARDER_KERBEROS_ENABLED_HOSTS`
* Value Format: `host1,host2,host3....`

List of hosts for which Kerberos (SPNEGO) authorization tokens will be injected as `Authorization` header. If HTTP request already has such header (or header is added by Forwarder by means of other settings, like custom headers or basic auth) - this header value will be overwritten by SPNEGO token.


### --kerberos-auth-upstream-proxy {#kerberos-auth-upstream-proxy}

* Environment variable: `FORWARDER_KERBEROS_AUTH_UPSTREAM_PROXY`
* Value Format: `<value>` (you can use empty command line switch to enable)
* Default Value: `false`

Authenticate to a configured upstream proxy with Kerberos (using `Proxy-Authorization` HTTP header). Please note that if forwarder configuration results in multiple proxies available (like PAC for example), forwarder will try to authenticate to each one of them.


### --kerberos-run-diagnostics {#kerberos-run-diagnostics}

* Environment variable: `FORWARDER_KERBEROS_RUN_DIAGNOSTICS`
* Value Format: `<value>` (you can use empty commant switch to enable)
* Default Value: `false`


Running forwarder with `--kerberos-run-diagnostics` switch will print debugging information about Kerberos connection or known configuration erros - for example an error when there are discrepancies between supported encryption types and keytab entry:

```
 msg="fatal error exiting" error="kerberos configuration potential problems: default_tkt_enctypes specifies 17 but this enctype is not available in the client's keytab\ndefault_tkt_enctypes specifies 23 but this enctype is not available in the client's keytab\npreferred_preauth_types specifies 17 but this enctype is not available in the client's keytab\npreferred_preauth_types specifies 15 but this enctype is not available in the client's keytab\npreferred_preauth_types specifies 14 but this enctype is not available in the client's keytab"
```

Diagnostics printout will allow you to match enctype number to string:

```
"DefaultTGSEnctypes": [
      "aes256-cts-hmac-sha1-96",
      "aes128-cts-hmac-sha1-96",
      "des3-cbc-sha1",
      "arcfour-hmac-md5",
      "camellia256-cts-cmac",
      "camellia128-cts-cmac",
      "des-cbc-crc",
      "des-cbc-md5",
      "des-cbc-md4"
    ],
    "DefaultTGSEnctypeIDs": [
      18,
      17,
      23
    ],

```

(17 is aes128-cts-hmac-sha1-96, etc)

Often having only one enctype in user configuration will work but can break at any time if hosts decide to negotiate something different than usual. For simplification you can restrict supported encryption types to 1-2 entries in krb5.conf file. Enctypes listed in diagnostics mode are sorted from most secure to least secure so in most cases first 1-2 positions are good enough to choose from and check if KDC server supports them. When in doubt - contact your ActiveDirectory/Kerberos administrator.


### Testing

There is a docker-compose file in /docs/content/config/kerberos path with predefined Squid configuration files that allows you to run Squid proxy requiring Kerberos authentication (assuming you have Kerberos KDC installed and configured with proper service SPN). It will allow you to test Kerberos authentication for upstream proxy.