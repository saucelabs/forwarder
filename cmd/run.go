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
	localProxyURI    string
	upstreamProxyURI string

	pacProxiesCredentials []string
	pacURI                string
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

- Local proxy: FORWARDER_LOCALPROXY_CREDENTIAL
- Upstream proxy: FORWARDER_UPSTREAMPROXY_CREDENTIAL
- PAC URI: PACMAN_CREDENTIAL
- PAC proxies: PACMAN_PROXIES_CREDENTIAL

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
	`,
	Run: func(cmd *cobra.Command, args []string) {
		p, err := proxy.New(localProxyURI, upstreamProxyURI, pacURI, pacProxiesCredentials, &proxy.LoggingOptions{
			Level:     logLevel,
			FileLevel: fileLevel,
			FilePath:  filePath,
		})
		if err != nil {
			cliLogger.Fatalln(customerror.NewFailedToError("run", "", err))
		}

		p.Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&localProxyURI, "local-proxy-uri", "l", "http://localhost:8080", "Sets local proxy URI")
	runCmd.Flags().StringVarP(&upstreamProxyURI, "upstream-proxy-uri", "u", "", "sets upstream proxy URI")
	runCmd.Flags().StringVarP(&pacURI, "pac-uri", "p", "", "sets URI to PAC content, or directly, the PAC content")
	runCmd.Flags().StringSliceVarP(&pacProxiesCredentials, "pac-proxies-credentials", "d", nil, "sets PAC proxies credentials using standard URI format")
}
