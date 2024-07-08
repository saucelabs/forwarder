---
title: SSLKEYLOGFILE
---

# SSLKEYLOGFILE

The SSLKEYLOGFILE environment variable enables logging of keys for decrypting SSL/TLS traffic.
Wireshark can then utilize these keys to decrypt SSL/TLS traffic.
By using SSLKEYLOGFILE, you can enhance the Forwarder's MITM capabilities with Wireshark-based tools.

## Prerequisites

[Wireshark](https://www.wireshark.org) must be installed on your machine.

## Run Forwarder with SSLKEYLOGFILE

To enable SSLKEYLOGFILE in Forwarder, set the environment variable `SSLKEYLOGFILE` to the path where the SSL keys should be logged.
Alternatively, you can use the `--http-tls-keylog-file` flag.

To use SSLKEYLOGFILE in Forwarder, you also need to enable MITM, see [MITM Proxy tutorial](/tutorials/mitm) for more details on this topic.
For the purpose of this tutorial, we will use the `--mitm` flag to enable MITM with self-signed certificates.

### Example

```bash
SSLKEYLOGFILE=/path/to/sslkeylog.log forwarder run --mitm 
```

## Run Wireshark

Open Wireshark, set Display filter to `http` and Capture filter to `tcp port 443`, and start capturing traffic on your network interface.

![Wireshark](/img/wireshark-0.png)

Go to `Edit` > `Preferences` > `Protocols` > `TLS`. Set the `(Pre)-Master-Secret log filename` to the same value as the `SSLKEYLOGFILE` environment variable.

Now if you run some traffic through the Forwarder, you should see decrypted traffic in Wireshark.

```bash
curl -k -x localhost:3128 https://example.com
```  

![Wireshark HTTP](/img/wireshark-1.png)
