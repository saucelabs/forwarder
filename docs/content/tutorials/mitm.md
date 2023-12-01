---
title: MITM Proxy
---

# Man-In-The-Middle (MITM) Proxy

Forwarder can be used as a MITM proxy.
It can be used to intercept and modify encrypted (HTTPS) traffic.
At the moment HTTP/2 support is limited to frame relay, so it's not possible to modify HTTP/2 traffic.

## Generating CA certificate

Forwarder can generate CA certificate automatically on startup.
To do that, you need to specify `--mitm` flag.
You can get the generated CA certificate from the API endpoint `/cacert`.

```bash
curl -o cacert.pem http://localhost:10000/cacert
```

This is the most secure way to generate the CA.
It's not possible to get certificate private key from Forwarder, and it's not stored anywhere.

Alternatively, you can generate the CA certificate manually and provide it to Forwarder using `--mitm-cacert` and `--mitm-cakey` flags.

## Installing the CA certificate

To use the CA certificate, you need to add it to the list of trusted CA certificates in your browser or operating system.
Note that in `curl` you can use `--cacert` flag to specify the CA certificate file without installing it.

```bash
curl --cacert cacert.pem https://example.com/
```

### Firefox

In Firefox you must install certificate separately from the operating system.

* Go to `about:preferences#privacy`
* Scroll down to `Certificates` section
* Click `View Certificates`
* Go to `Authorities` tab
* Click `Import`
* Select the CA certificate file
* Check `Trust this CA to identify websites`
* Click `OK`

### macOS

```bash
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain cacert.pem
```

### Linux

```bash
sudo cp cacert.pem /usr/local/share/ca-certificates/forwarder-cacert.pem
sudo update-ca-certificates
```

### Windows

```bash
certutil -addstore -f "ROOT" cacert.pem
```

## Enabling MITM only for specific hosts

By default, Forwarder will intercept all HTTPS traffic.
You can specify a list of domain regular expressions for which MITM should be enabled using `--mitm-domains` flag.
The following example will enable MITM for all `example.com` subdomains except `foo.example.com`.

```bash
forwarder run --mitm --mitm-domains '.*\.example\.com$,-foo\.example\.com$'
```

## Using the MITM proxy

Make sure that Forwarder is running with `--mitm` flag and the CA certificate is installed.
Then [configure your browser](/browser.md) to use Forwarder as a proxy server.
