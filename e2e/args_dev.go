//go:build dev

package e2e

func init() {
	*proxy = "https://localhost:3128"
	*httpbin = "https://httpbin"
	*insecureSkipVerify = true
}
