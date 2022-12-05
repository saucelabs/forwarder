#!/usr/bin/env bash

on_error() {
  make dump-logs
}

# Profile specifies the number of seconds to run CPU profiling for.
PROFILE=0

# ARGS can be used to pass additional arguments to the e2e test.
ARGS=""

run_test() {
  CI=${CI:-""}
  if [[ -n "${CI}" ]]; then
    trap 'on_error' ERR
  fi

  HTTPBIN_SCHEME=$1
  PROXY_SCHEME=$2
  UPSTREAM_PROXY_SCHEME=$3
  EXTRA_CONF=$4

  CONF="override/httpbin-${HTTPBIN_SCHEME}.yaml override/proxy-${PROXY_SCHEME}.yaml"
  if [[ -n "${UPSTREAM_PROXY_SCHEME}" ]]; then
    CONF="${CONF} override/upstream-proxy-${UPSTREAM_PROXY_SCHEME}.yaml"
  fi
  if [[ -n "${EXTRA_CONF}" ]]; then
    CONF="${CONF} ${EXTRA_CONF}"
  fi
  make up CONF="${CONF}"

  if [[ ${PROFILE} -gt 0 ]]; then
    TMPDIR=$(mktemp -d -t "com.saucelabs.Forwarder.XXXXXX")
    echo ">>> Profiling enabled output in \"${TMPDIR}\""
    curl -sS "http://localhost:10000/debug/pprof/profile?seconds=${PROFILE}" -o "${TMPDIR}/cpu" &
  fi

  if [[ "${HTTPBIN_SCHEME}" == "h2" ]]; then
    HTTPBIN_SCHEME="https"
  fi
  if [[ "${PROXY_SCHEME}" == "h2" ]]; then
    PROXY_SCHEME="https"
  fi
  make test ARGS="${ARGS} -httpbin ${HTTPBIN_SCHEME}://httpbin -proxy ${PROXY_SCHEME}://proxy:3128 -insecure-skip-verify"

  if [[ ${PROFILE} -gt 0 ]]; then
    curl -sS "http://localhost:10000/debug/pprof/allocs" -o "${TMPDIR}/allocs"
    curl -sS "http://localhost:10000/debug/pprof/heap" -o "${TMPDIR}/heap"
    curl -sS "http://localhost:10000/debug/pprof/mutex" -o "${TMPDIR}/mutex"
  fi
}

run_bench() {
  if [[ -z "${RUN}" ]]; then
    echo "RUN is not set, skipping benchmark"
    exit 1
  fi
  if [[ ${PROFILE} -eq 0 ]]; then
    PROFILE=10
  fi
  ARGS="-test.bench ${RUN} -test.benchtime ${PROFILE}s -test.v"
  RUN="^XXX"
  run_test "$@"
}
