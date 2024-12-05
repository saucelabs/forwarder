# Gnet

This benchmark runs a gnet based HTTP server and allows to benchmark it using wrk in various configurations.

## Prerequisites

* Install wrk: `brew install wrk`
* Optionally install envoy: `brew install envoy` (for the envoy benchmark)

## Running the benchmark

* Run the test server: `make run`
* Run the benchmark: `make bench`
* Run Forwarder for the benchmark: `make run-forwarder`
* Run proxy benchmark: `make bench-proxy`

Optionally you can run the Envoy benchmark by running `make run-envoy` and `make bench-proxy`. 
