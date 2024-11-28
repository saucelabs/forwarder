# Browser test

This directory contains a test that loads real websites over the proxy.
It assumes the proxy is available on `localhost:3128`.

The test is implemented using Microsoft's [Playwright](https://playwright.dev/) and uses Chromium.

## Running the test

To run the test, first start the proxy, and then run the test with `make`.
You can also run the test directly with `make test` provided you have the dependencies installed.
