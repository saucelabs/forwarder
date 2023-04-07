// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package testrunner

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Runnable
	ConfigFile any
}

func (c Config) SaveConfigFile(path string) error {
	if c.ConfigFile == nil {
		return nil
	}

	b, err := yaml.Marshal(c.ConfigFile)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", c.Name, err)
	}
	path = filepath.Join(path, c.Name+".yaml")
	if err = writeFile(path, b); err != nil {
		return fmt.Errorf("write to file %s: %w", c.Name, err)
	}
	return nil
}

type E2E struct {
	Name    string
	Configs []Config
}

func (e *E2E) Save(path string) error {
	path = filepath.Join(path, e.Name)
	if err := e.saveConfigFiles(path); err != nil {
		return err
	}
	if err := e.saveRunnerConfig(path); err != nil {
		return err
	}
	return nil
}

func (e *E2E) saveConfigFiles(path string) error {
	for _, c := range e.Configs {
		if err := c.SaveConfigFile(path); err != nil {
			return err
		}
	}
	return nil
}

func (e *E2E) saveRunnerConfig(path string) error {
	runs := make([]Runnable, 0, len(e.Configs))
	for _, c := range e.Configs {
		if c.ConfigFile != nil {
			c.Command = append(c.Command, "--config-file="+filepath.Join(path, c.Name+".yaml"))
		}
		runs = append(runs, c.Runnable)
	}
	b, err := yaml.Marshal(runs)
	if err != nil {
		return err
	}
	return writeFile(filepath.Join(path, "run.yaml"), b)
}

func createDir(name string) error {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return os.MkdirAll(name, 0o755)
	}
	return err
}

func writeFile(name string, content []byte) error {
	if err := createDir(filepath.Dir(name)); err != nil {
		return err
	}
	return os.WriteFile(name, content, 0o600)
}
