// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cobrautil

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// AppendEnvToUsage appends the environment variable name to the usage string of each Cobra flag.
func AppendEnvToUsage(cmd *cobra.Command, envPrefix string) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		f.Usage += fmt.Sprintf(" env: %s", EnvName(envPrefix, f.Name))
	})
}

// BindFlagsToEnv binds each Cobra flag to its associated Viper configuration (config file and environment variable).
func BindFlagsToEnv(cmd *cobra.Command, envPrefix string) error {
	v := viper.New()

	var bindErr error
	fs := cmd.Flags()
	fs.VisitAll(func(f *pflag.Flag) {
		// Bind environment variable to flag
		if err := v.BindEnv(f.Name, EnvName(envPrefix, f.Name)); err != nil {
			bindErr = err
			return
		}

		// Set default value from environment variable
		if !f.Changed && v.IsSet(f.Name) {
			if err := fs.Set(f.Name, fmt.Sprintf("%v", v.Get(f.Name))); err != nil {
				bindErr = err
				return
			}
		}
	})
	return bindErr
}

func EnvName(envPrefix, flagName string) string {
	name := flagName
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ToUpper(name)
	return fmt.Sprintf("%s_%s", envPrefix, name)
}
