# PAC and Proxy

Source: <https://github.com/elazarl/goproxy/issues/429>

```go
gServer := goproxy.NewProxyHttpServer()

gServer.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
    proxy, _ := url.Parse("http://user:pass@proxy.host:port")
    gServer.Tr = &http.Transport{
        Proxy: http.ProxyURL(proxy),
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
})
```

Source: <https://github.com/elazarl/goproxy/issues/439>

```go
u, err := url.Parse(externalProxy)

proxy.ConnectDial = proxy.NewConnectDialToProxyWithHandler(externalProxy, func(req *http.Request) {
    if u.User != nil {
        credentials := base64.StdEncoding.EncodeToString([]byte(u.User.String()))
        req.Header.Add("Proxy-Authorization", "Basic "+credentials)
    }
})
```
