// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/saucelabs/forwarder/internal/customerror"
	"github.com/saucelabs/forwarder/pkg/proxy"
	"github.com/spf13/cobra"
)

var (
	credential            string
	host                  string
	parentProxyCredential string
	parentProxyURL        string
)

// runCmd represents the run command.
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Starts the proxy",
	Long: `Run starts the proxy. Proxy can be protected with basic auth. To enable that,
set the --credential flag. It can also forward connections to a parent proxy.
To do that, just set the --parentProxyURL flag. If the parent proxy requires
auth, just set the --parentProxyCredential flag.

Both local, and parent credentials can be set via environment variables.
For local proxy credential, set PROXY_CREDENTIAL. For remote proxy
credential, set PROXY_PARENT_CREDENTIAL.`,
	Example: `  Start a proxy at localhost:8080
  $ proxy run

  Start a proxy with default values
  $ proxy run

  Start a proxy at localhost:8085
  $ proxy run --host localhost:8085
  
  Start a proxy at localhost:8085 with basic auth
  $ proxy run --host localhost:8085 --credential "user:pwd"
  
  Start a proxy at localhost:8085 with basic auth, and forwarding connection to
  a parent proxy running at http://localhost:8089.
  $ proxy run \
    --host localhost:8085 \
    --credential "user:pwd" \
    --parentProxyURL "http://localhost:8089"

  Start a proxy at localhost:8085 with basic auth, forwarding connection to a
  parent proxy running at http://localhost:8089 which requires basic auth.
  $ proxy run \
    --host localhost:8085 \
    --credential "user:pwd" \
    --parentProxyURL "http://localhost:8089" \
    --parentProxyCredential "user1:pwd1"
	
  Start a proxy at localhost:8085 with basic auth, forwarding connection to a
  parent proxy running at http://localhost:8089 which requires basic auth,
  setting credentials via environment variables.
  $ PROXY_CREDENTIAL="user:pwd" PROXY_PARENT_CREDENTIAL="user1:pwd1" proxy run \
    --host localhost:8085 \
    --parentProxyURL "http://localhost:8089"`,
	Run: func(cmd *cobra.Command, args []string) {
		p, err := proxy.New(host, credential, parentProxyURL, parentProxyCredential, &proxy.LoggingOptions{
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

	runCmd.Flags().StringVarP(&credential, "credential", "c", "", "Sets proxy basic auth credential")
	runCmd.Flags().StringVarP(&host, "host", "o", "localhost:8080", "Sets proxy host") // Can't u `h`, it's reserved by Cobra
	runCmd.Flags().StringVar(&parentProxyCredential, "parentProxyCredential", "", "Sets parent proxy basic auth credential")
	runCmd.Flags().StringVar(&parentProxyURL, "parentProxyURL", "", "Sets parent proxy url")
}
