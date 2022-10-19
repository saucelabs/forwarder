function FindProxyForURL(url, host) {
    ip = dnsResolve(host);
    return "PROXY " + ip + ":8080";
}
