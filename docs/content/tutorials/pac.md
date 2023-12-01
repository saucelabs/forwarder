---
title: PAC files
weight: 50
---

# Proxy Auto-Configuration (PAC)

Forwarder implements [Proxy Auto-Configuration (PAC)](https://developer.mozilla.org/en-US/docs/Web/HTTP/Proxy_servers_and_tunneling/Proxy_Auto-Configuration_PAC_file) file support.
PAC file is a JavaScript file that returns a proxy URL for a given URL.
It can be used to implement complex proxy rules.

Forwarder also implements [Microsoft's PAC extensions for IPv6](https://learn.microsoft.com/en-us/windows/win32/winhttp/ipv6-aware-proxy-helper-api-definitions).
So you can use `FindProxyForURL` and `FindProxyForURLEx` functions to implement IPv6-aware proxy rules.

## PAC file server

Forwarder can be used as a PAC file server.
See [forwarder pac server](cli/forwarder_pac_server.md) command reference for more details. 

## PAC file evaluation

Forwarder can be used to evaluate PAC files.
See [forwarder pac eval](cli/forwarder_pac_eval.md) command reference for more details.

## Functions you can use in PAC files

The following JavaScript functions are implemented in Forwarder:

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
