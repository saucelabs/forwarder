// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ready

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/spf13/cobra"
)

type command struct {
	apiAddr string
}

func (c *command) RunE(cmd *cobra.Command, args []string) error {
	resp, err := http.Get("http://" + c.apiAddr + "/readyz") //nolint:noctx // net/http.Get must not be called
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return err
		}
		if _, err := cmd.ErrOrStderr().Write(b); err != nil {
			return err
		}

		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func Command() (cmd *cobra.Command) {
	c := command{
		apiAddr: "localhost:10000",
	}

	defer func() {
		fs := cmd.Flags()
		bindAPIAddr(fs, &c.apiAddr)
	}()
	return &cobra.Command{
		Use:   "ready [--api-address <host:port>] [flags]",
		Short: "Readiness probe for the Forwarder",
		Long:  long,
		RunE:  c.RunE,
	}
}

const long = `Readiness probe for the Forwarder.
This is equivalent to calling /readyz endpoint on the Forwarder API server.`
