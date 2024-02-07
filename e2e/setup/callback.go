// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package setup

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func makeTestCallback(run string, debug bool) func() error {
	return func() error {
		cmd := exec.Command("make", "test")
		if run != "" {
			cmd.Env = append(os.Environ(), "RUN="+run)
		}
		if debug {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			stdout.WriteTo(os.Stdout)
			stderr.WriteTo(os.Stderr)
			return err
		}

		if stderr.Len() > 0 {
			fmt.Fprintln(os.Stderr, "stderr:")
			stderr.WriteTo(os.Stderr)
			fmt.Fprintln(os.Stderr)
			return errors.New("unexpected stderr")
		}

		s := strings.Split(stdout.String(), "\n")
		for _, l := range s {
			if strings.HasPrefix(l, "---") || strings.Contains(l, "SKIP:") {
				fmt.Fprintln(os.Stdout, l)
			}
		}

		return nil
	}
}
