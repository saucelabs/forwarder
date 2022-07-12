// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/saucelabs/customerror"
	"github.com/saucelabs/forwarder/pkg/proxy"
	"github.com/spf13/cobra"
)

var (
	dnsURIs []string

	localProxyURI    string
	upstreamProxyURI string

	siteCredentials []string

	pacProxiesCredentials []string
	pacURI                string

	automaticallyRetryPort bool
	proxyLocalhost         bool
)

// runCmd represents the run command.
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Starts the proxy",
	Long: `Run starts the proxy. Proxy can be protected with basic auth.
It can forward connections to an upstream proxy (protected, or not).
The upstream proxy can be automatically setup via PAC (protected, or not).
Also, credentials for proxies specified in PAC can be set.

All credentials can be set via env vars:

- Local proxy: FORWARDER_LOCALPROXY_AUTH
- Upstream proxy: FORWARDER_UPSTREAMPROXY_AUTH
- PAC URI: PACMAN_AUTH
- PAC proxies: PACMAN_PROXIES_AUTH
- Target URLs: FORWARDER_SITE_CREDENTIALS

Note: Can't setup upstream, and PAC at the same time.
`,
	Example: `  Start a proxy listening to http://localhost:8080:
  $ forwarder run

  Start a proxy listening to http://0.0.0.0:8085:
  $ forwarder run -l "http://0.0.0.0:8085"

  Start a protected proxy:
  $ forwarder run -l "http://user:pwd@localhost:8085"

  Start a protected proxy, forwarding connection to an upstream proxy running at
  http://localhost:8089:
  $ forwarder run \
    -l "http://user:pwd@localhost:8085" \
    -u "http://localhost:8089"

  Start a protected proxy, forwarding connection to a protected upstream proxy
  running at http://user1:pwd1@localhost:8089:
  $ forwarder run \
    -l "http://user:pwd@localhost:8085" \
    -u "http://user1:pwd1@localhost:8089"

  Start a protected proxy, forwarding connection to an upstream proxy, setup via
  PAC - server running at http://localhost:8090:
  $ forwarder run \
    -l "http://user:pwd@localhost:8085" \
    -p "http://localhost:8090"

  Start a protected proxy, forwarding connection to an upstream proxy, setup via
  PAC - protected server running at http://user2:pwd2@localhost:8090:
  $ forwarder run \
    -l "http://user:pwd@localhost:8085" \
    -p "http://user2:pwd2@localhost:8090"

  Start a protected proxy, forwarding connection to an upstream proxy, setup via
  PAC - protected server running at http://user2:pwd2@localhost:8090, specifying
  credential for protected proxies specified in PAC:
  $ forwarder run \
    -l "http://user:pwd@localhost:8085" \
    -p "http://user2:pwd2@localhost:8090" \
	  -d "http://user3:pwd4@localhost:8091,http://user4:pwd5@localhost:8092"

  Start a protected proxy, forwarding connection to an upstream proxy, setup via
  PAC - protected server running at http://user2:pwd2@localhost:8090, specifying
  credential for protected proxies specified in PAC, also forwarding "localhost"
  requests thru the upstream proxy:
  $ forwarder run \
    -t \
    -l "http://user:pwd@localhost:8085" \
    -p "http://user2:pwd2@localhost:8090" \
	  -d "http://user3:pwd4@localhost:8091,http://user4:pwd5@localhost:8092"

  Start a protected proxy that adds basic auth header to requests to foo.bar:8090
  and qux.baz:80.
  $ forwarder run \
    -t \
    -l "http://user:pwd@localhost:8085" \
	--site-credentials "user1:pwd1@foo.bar:8090,user2:pwd2@qux:baz:80"
	`,
	Run: func(cmd *cobra.Command, args []string) {
		p, err := proxy.New(localProxyURI, upstreamProxyURI, pacURI, pacProxiesCredentials, &proxy.Options{
			LoggingOptions: &proxy.LoggingOptions{
				Level:     logLevel,
				FileLevel: fileLevel,
				FilePath:  filePath,
			},

			AutomaticallyRetryPort: automaticallyRetryPort,
			DNSURIs:                dnsURIs,
			ProxyLocalhost:         proxyLocalhost,
			SiteCredentials:        siteCredentials,
		})
		if err != nil {
			cliLogger.Fatalln(customerror.NewFailedToError("run", customerror.WithError(err)))
		}

		p.Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&localProxyURI, "local-proxy-uri", "l", "http://localhost:8080", "sets local proxy URI")
	runCmd.Flags().StringVarP(&upstreamProxyURI, "upstream-proxy-uri", "u", "", "sets upstream proxy URI")
	runCmd.Flags().StringSliceVarP(&dnsURIs, "dns-uri", "n", nil, "sets dns URI")
	runCmd.Flags().StringVarP(&pacURI, "pac-uri", "p", "", "sets URI to PAC content, or directly, the PAC content")
	runCmd.Flags().StringSliceVarP(&pacProxiesCredentials, "pac-proxies-credentials", "d", nil, "sets PAC proxies credentials using standard URI format")
	runCmd.Flags().StringSliceVar(&siteCredentials, "site-credentials", nil, "sets site based credentials")
	runCmd.Flags().BoolVarP(&proxyLocalhost, "proxy-localhost", "t", false, "if set, will proxy localhost requests to an upstream proxy - if any")
	runCmd.Flags().BoolVarP(&automaticallyRetryPort, "find-port", "r", true, "if set, and the specified local proxy port is in-use, it will find, and use an available one")
}
