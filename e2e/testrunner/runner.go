// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package testrunner

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

type Runnable struct {
	Name    string   `yaml:"name,omitempty"`
	Command []string `yaml:"command,omitempty"`
}

func (r *Runnable) RunAndSaveOutput(ctx context.Context, path string) error {
	if r.Command == nil || len(r.Command) == 0 {
		return fmt.Errorf("no command to run for %s", r.Name)
	}

	cmd := exec.CommandContext(ctx, r.Command[0], r.Command[1:]...) //nolint:gosec // this is a test runner
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	defer saveOutput(filepath.Join(path, r.Name), stdout, stderr)
	return cmd.Run()
}

func saveOutput(path string, stdout, stderr *bytes.Buffer) {
	if err := createDir(filepath.Dir(path)); err != nil {
		log.Printf("cannot create dir %s: %v", filepath.Dir(path), err)
		return
	}
	if err := writeFile(path+".stdout", stdout.Bytes()); err != nil {
		log.Printf("cannot write to %s.stdout: %v", path, err)
		return
	}
	if err := writeFile(path+".stderr", stderr.Bytes()); err != nil {
		log.Printf("cannot write to %s.stderr: %v", path, err)
		return
	}
}

type Configuration struct {
	Name      string
	Runnables []Runnable
}

type RunnerConfig struct {
	Root             string
	ConcurrencyLimit int
}

type Runner struct {
	RunnerConfig
	g *errgroup.Group
	m *sync.Map
}

func NewRunner(cfg RunnerConfig) *Runner {
	r := &Runner{
		RunnerConfig: cfg,
		g:            &errgroup.Group{},
		m:            &sync.Map{},
	}
	r.g.SetLimit(r.ConcurrencyLimit)
	return r
}

func (r *Runner) run() error {
	return filepath.WalkDir(r.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if r.Root == path || !d.IsDir() {
			return nil
		}
		b, err := os.ReadFile(filepath.Join(path, "run.yaml"))
		if err != nil {
			return err
		}
		var runnables []Runnable
		if err := yaml.Unmarshal(b, &runnables); err != nil {
			return err
		}
		c := Configuration{
			Name:      filepath.Base(path),
			Runnables: runnables,
		}
		r.g.Go(func() error {
			err := r.runConfiguration(&c)
			if err != nil {
				r.m.Store(c.Name, err)
			}
			return err
		})
		return nil
	})
}

func ignoreProcessKilled(err error) error {
	if err.Error() == "signal: killed" || errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func (r *Runner) runConfiguration(c *Configuration) error {
	if len(c.Runnables) == 0 {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)
	path := filepath.Join("test-outputs", c.Name)
	for _, r := range c.Runnables[1:] {
		r := r
		g.Go(func() error {
			return ignoreProcessKilled(r.RunAndSaveOutput(ctx, path))
		})
	}
	g.Go(func() error {
		defer cancel()
		return c.Runnables[0].RunAndSaveOutput(ctx, path)
	})
	return g.Wait()
}

func (r *Runner) Run() error {
	if err := r.run(); err != nil {
		return err
	}
	if testErr := r.g.Wait(); testErr != nil {
		var tests []any
		r.m.Range(func(k, _ interface{}) bool {
			tests = append(tests, k)
			return true
		})
		return fmt.Errorf("failed: %v", tests)
	}
	return nil
}
