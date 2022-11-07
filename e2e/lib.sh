#!/usr/bin/env bash

on_error() {
  make dump-logs
}

run_test() {
  trap 'on_error' ERR

  HTTPBIN_SCHEME=$1
  PROXY_SCHEME=$2
  UPSTREAM_PROXY_SCHEME=$3
  EXTRA_CONF=$4

  CONF="httpbin-${HTTPBIN_SCHEME}.yaml override/proxy-${PROXY_SCHEME}.yaml"
  if [[ -n "${UPSTREAM_PROXY_SCHEME}" ]]; then
    CONF="${CONF} override/upstream-proxy-${UPSTREAM_PROXY_SCHEME}.yaml"
  fi
  if [[ -n "${EXTRA_CONF}" ]]; then
    CONF="${CONF} ${EXTRA_CONF}"
  fi
  make up CONF="${CONF}"

  if [[ "${HTTPBIN_SCHEME}" == "h2" ]]; then
    HTTPBIN_SCHEME="https"
  fi
  if [[ "${PROXY_SCHEME}" == "h2" ]]; then
    PROXY_SCHEME="https"
  fi

  make test ARGS="-httpbin ${HTTPBIN_SCHEME}://httpbin -proxy ${PROXY_SCHEME}://proxy:3128 -insecure-skip-verify"
}
