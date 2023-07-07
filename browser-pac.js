// Point browser to this PAC file to attach it to the e2e test proxy.
function FindProxyForURL(url, host) {
    return "PROXY localhost:3128";
}
