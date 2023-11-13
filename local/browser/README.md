# Browser test

This directory contains a test that loads real websites over the proxy.
It assumes the proxy is available on `localhost:3128`.

The test is implemented using Microsoft's [Playwright](https://playwright.dev/) and uses Firefox.

## Running the test

To run the test, first start the proxy, and then run the test with `make`.
You can also run the test directly with `make test` provided you have the dependencies installed.

### Running whit MITM

When running with MITM, use `/e2e/certs/ca.crt` as the MITM CA certificate and add it to the browser's trusted certificates,
see [Setting Up Certificate Authorities (CAs) in Firefox](https://support.mozilla.org/en-US/kb/setting-certificate-authorities-firefox).
Do not add the certificate to the system's trusted certificates.

## Extending the test

Note that it's possible to record test cases using [test generator](https://playwright.dev/docs/codegen).
