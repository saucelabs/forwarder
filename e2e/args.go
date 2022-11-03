package e2e

import "flag"

var (
	proxy              = flag.String("proxy", "", "URL of the proxy to test against")
	httpbin            = flag.String("httpbin", "", "URL of the httpbin server to test against")
	insecureSkipVerify = flag.Bool("insecure-skip-verify", false, "Skip TLS certificate verification")
)
