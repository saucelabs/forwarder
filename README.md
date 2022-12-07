# Forwarder

Forwarder is HTTP proxy.
It can be used to proxy HTTP/HTTPS/HTTP2 requests, Server Sent Events, WebSockets, TCP traffic and more.
It supports downstream HTTP(S) and SOCKS5 proxies and basic authentication.

## Proxy Auto-Configuration (PAC)

Forwarder implements [Proxy Auto-Configuration (PAC)](https://developer.mozilla.org/en-US/docs/Web/HTTP/Proxy_servers_and_tunneling/Proxy_Auto-Configuration_PAC_file) file support.
PAC file is a JavaScript file that returns a proxy URL for a given URL.
It can be used to implement complex proxy rules.

Forwarder also implements [Microsoft's PAC extensions for IPv6](https://learn.microsoft.com/en-us/windows/win32/winhttp/ipv6-aware-proxy-helper-api-definitions).
So you can use `FindProxyForURL` and `FindProxyForURLEx` functions to implement IPv6-aware proxy rules.

Forwarder can be used as a PAC file server (see `pac-server` command), or as a PAC file test util (see `pac-eval` command).
