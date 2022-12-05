#!/usr/bin/env bash

set -eu -o pipefail

source ./lib.sh

# Runs tests for all combinations of "http" and "https" with and without upstream proxy.
for HTTPBIN_SCHEME in "http" "https" "h2"; do
  for PROXY_SCHEME in "http" "https"; do
    for UPSTREAM_PROXY_SCHEME in "" "http" "https"; do
      run_test "${HTTPBIN_SCHEME}" "${PROXY_SCHEME}" "${UPSTREAM_PROXY_SCHEME}" "override/proxy-basic-auth.yaml override/proxy-goleak.yaml"
    done
  done
done

# Runs tests with upstream proxy basic auth enabled.
for HTTPBIN_SCHEME in "http" "https" "h2"; do
  run_test "${HTTPBIN_SCHEME}" "http" "http" "override/upstream-basic-auth.yaml"
done

# Runs tests with PAC
run_test "http" "http" "" "override/proxy-pac-direct.yaml"
run_test "http" "http" "http" "override/proxy-pac-upstream.yaml"

RUN="Localhost" run_test "http" "http" "" "override/proxy-localhost.yaml"

# Runs sc issue repro tests.
RUN="SC[0-9]+" run_test "http" "http" "" "sc-2450/service.yaml"
