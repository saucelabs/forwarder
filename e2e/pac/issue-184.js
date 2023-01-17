function FindProxyForURL(url, host) {
    if ( dnsDomainIs (host, ".google.com")) {
        return "DIRECT";
    }
    return "PROXY foo";
}
