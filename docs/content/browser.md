---
title: Setup Browser
weight: 20
---

# Setup Browser

Configure your browser to use Forwarder as a proxy server.
The below instructions assume that Forwarder is running on `localhost` and listening on port `10000`.

If using `curl` or `wget`, you can simply use `--proxy` flag to specify the proxy server.

```bash
curl --proxy http://localhost:10000 https://example.com/
```

## Firefox

* Go to `about:preferences#general`
* Scroll down to `Network Settings`
* Click `Settings`
* Select `Manual proxy configuration`
* Set `HTTP Proxy` to `localhost` and `Port` to `10000`
* Check `Use this proxy server for all protocols`
* Click `OK`

## Chrome 

* Go to `chrome://settings`
* Scroll down to `System`
* Click `Open proxy settings`
* Click `LAN settings`
* Check `Use a proxy server for your LAN`
* Set `Address` to `localhost` and `Port` to `10000`
* Click `OK`

## Safari
* Go to `Safari > Settings`
* Go to `Advanced` tab
* Scroll to `Proxies`
* Click `Change Settings` 
* Check `Web Proxy (HTTP)`
* Set `Web Proxy Server` to `localhost` and `Port` to `10000`
* Click `OK`

## Edge 

* Go to `edge://settings`
* Scroll down to `System`
* Click `Open your computer's proxy settings`
* Click `LAN settings`
* Check `Use a proxy server for your LAN`
* Set `Address` to `localhost` and `Port` to `10000`
* Click `OK`

## Opera 

* Go to `Settings`
* Scroll down to `Network`
* Click `Change proxy settings`
* Click `LAN settings`
* Check `Use a proxy server for your LAN`
* Set `Address` to `localhost` and `Port` to `10000`
* Click `OK`
