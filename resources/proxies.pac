function FindProxyForURL(url, host) {
  if (
    dnsDomainIs(host, "intranet.domain.com") ||
    shExpMatch(host, "(*.abcdomain.com|abcdomain.com)")
  )
    return "DIRECT";

  if (isPlainHostName(host)) return "DIRECT";
  else return "PROXY 127.0.0.1:8080; PROXY 127.0.0.1:8081; DIRECT";
}
