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

var envReplacer = strings.NewReplacer(".", "_", "-", "_") //nolint:gochecknoglobals // false positive

// BindAll updates the given command flags with values from the environment variables and config file.
// The supported formats are: JSON, YAML, TOML, HCL, and Java properties.
// The file format is determined by the file extension, if not specified the default format is YAML.
// The following precedence order of configuration sources is used: command flags, environment variables, config file, default values.
func BindAll(cmd *cobra.Command, envPrefix, configFileFlagName string) error {
	v := viper.New()

	// Flags
	if err := v.BindPFlags(cmd.PersistentFlags()); err != nil {
		return err
	}
	if err := v.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	// Environment variables
	v.SetEnvKeyReplacer(envReplacer)
	envPrefix = strings.ToUpper(envPrefix)
	envPrefix = envReplacer.Replace(envPrefix)
	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()

	// Config file
	if configFileFlagName != "" {
		if f := v.GetString(configFileFlagName); f != "" {
			v.SetConfigType("yaml")
			v.SetConfigFile(f)
			if err := v.ReadInConfig(); err != nil {
				return err
			}
		}
	}

	// Update cobra flags with values from viper
	updateFs := func(fs *pflag.FlagSet) (ok bool) {
		ok = true
		fs.VisitAll(func(f *pflag.Flag) {
			if !f.Changed && v.IsSet(f.Name) {
				s := fmt.Sprintf("%v", v.Get(f.Name))
				s = strings.TrimPrefix(s, "[")
				s = strings.TrimSuffix(s, "]")
				s = strings.NewReplacer(", ", ",", " ", ",").Replace(s)
				if err := fs.Set(f.Name, s); err != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
					ok = false
				}
			}
		})
		return
	}

	if !updateFs(cmd.PersistentFlags()) {
		return fmt.Errorf("failed to update persistent flags")
	}

	if !updateFs(cmd.Flags()) {
		return fmt.Errorf("failed to update flags")
	}

	return nil
}
