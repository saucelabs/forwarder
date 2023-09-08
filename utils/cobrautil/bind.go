// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cobrautil

import (
	"fmt"
	"strings"

	"github.com/spf13/cast"
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
				value := v.Get(f.Name)
				if err := setFlagFromViper(f, value); err != nil {
					var flagName string
					if f.Shorthand != "" && f.ShorthandDeprecated == "" {
						flagName = fmt.Sprintf("-%s, --%s", f.Shorthand, f.Name)
					} else {
						flagName = fmt.Sprintf("--%s", f.Name)
					}
					fmt.Fprintf(cmd.ErrOrStderr(), "invalid argument %q for %q flag: %v", value, flagName, err)
					ok = false
				} else {
					if f.Deprecated != "" {
						fmt.Fprintf(cmd.ErrOrStderr(), "Flag --%s has been deprecated, %s\n", f.Name, f.Deprecated)
					}
					f.Changed = true
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

func setFlagFromViper(f *pflag.Flag, v any) error {
	if vs, ok := v.([]any); ok {
		sr, ok := f.Value.(sliceReplacer)
		if !ok {
			return fmt.Errorf("trying to set list to %s", f.Value.Type())
		}
		ss, err := cast.ToStringSliceE(vs)
		if err != nil {
			return err
		}
		return sr.Replace(ss)
	}

	return f.Value.Set(fmt.Sprintf("%v", v))
}

type sliceReplacer interface {
	Replace(val []string) error
}
