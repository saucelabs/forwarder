// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package compose

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"slices"
	"strings"
	"time"
)

type Command struct {
	rt     string
	sep    string
	dir    string
	stdout io.Writer
	stderr io.Writer
}

func NewCommand(c *Compose, dir string, stdout, stderr io.Writer) (*Command, error) {
	rt := os.Getenv("CONTAINER_RUNTIME")
	if rt == "" {
		rt = "docker"
	}

	sep := "_"
	if rt == "docker" || hasDockerCompose() {
		sep = "-"
	}

	if dir == "" {
		d, err := os.MkdirTemp("", "compose-*")
		if err != nil {
			return nil, fmt.Errorf("create temp file: %w", err)
		}
		dir = d
	}

	f, err := os.Create(path.Join(dir, "compose.yaml"))
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer f.Close()

	if _, err := c.WriteTo(f); err != nil {
		return nil, fmt.Errorf("write compose to file: %w", err)
	}
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("close temp file: %w", err)
	}

	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}

	return &Command{
		rt:     rt,
		sep:    sep,
		dir:    dir,
		stdout: stdout,
		stderr: stderr,
	}, nil
}

func (c *Command) Runtime() string {
	return c.rt
}

func (c *Command) Project() string {
	if p := os.Getenv("COMPOSE_PROJECT_NAME"); p != "" {
		return p
	}

	return path.Base(c.dir)
}

func (c *Command) File() string {
	return path.Join(c.dir, "compose.yaml")
}

func (c *Command) Close() error {
	return os.RemoveAll(c.dir)
}

func (c *Command) Up(args ...string) error {
	if slices.ContainsFunc(args, func(s string) bool { return s == "-d" || s == "--detach" }) {
		return c.quietRun(c.cmd("up", args))
	}

	return c.run(c.cmd("up", args))
}

func (c *Command) Down(args ...string) error {
	return c.quietRun(c.cmd("down", args))
}

func (c *Command) Ps(args ...string) error {
	return c.run(c.cmd("ps", args))
}

func (c *Command) Logs(args ...string) error {
	return c.run(c.cmd("logs", args))
}

const healthy = "healthy"

type serviceHealth struct {
	Service string
	Health  string
}

func (sh serviceHealth) String() string {
	return fmt.Sprintf("%s: %s", sh.Service, sh.Health)
}

func (c *Command) Wait(interval, timeout time.Duration, services []string) error {
	to := time.NewTimer(timeout)
	defer to.Stop()
	t := time.NewTicker(interval)
	defer t.Stop()

	lastStatus := make([]serviceHealth, len(services))

	for {
		select {
		case <-to.C:
			return fmt.Errorf("timeout waiting for services to be healthy: %v", lastStatus)
		case <-t.C:
			n := 0
			for i, s := range services {
				h := c.serviceHealth(s)
				if h == "" || h == healthy {
					n++
				}
				lastStatus[i] = serviceHealth{Service: s, Health: h}
			}
			if n == len(services) {
				return nil
			}
		}
	}
}

func (c *Command) serviceHealth(s string) string {
	args := []string{
		"inspect",
		"--format",
		"{{.State.Health.Status}}",
		c.serviceContainerName(s),
	}

	cmd := exec.Command(c.rt, args...) //nolint:gosec // this is a command runner
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if bytes.Contains(stderr.Bytes(), []byte("no such object")) {
			return "no such object"
		}
		if bytes.Contains(stderr.Bytes(), []byte("nil pointer evaluating *define.HealthCheckResults.Status")) {
			return ""
		}

		stdout.WriteTo(c.stdout)
		stderr.WriteTo(c.stderr)
	}

	return strings.TrimSpace(stdout.String())
}

func (c *Command) serviceContainerName(s string) string {
	return fmt.Sprintf("%s%s%s%s1", c.Project(), c.sep, s, c.sep)
}

func (c *Command) cmd(subcmd string, args []string) *exec.Cmd {
	allArgs := []string{
		"compose",
		subcmd,
	}
	allArgs = append(allArgs, args...)

	cmd := exec.Command(c.rt, allArgs...) //nolint:gosec // this is a command runner
	cmd.Dir = c.dir

	return cmd
}

func (c *Command) quietRun(cmd *exec.Cmd) error {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		stdout.WriteTo(c.stdout)
		stderr.WriteTo(c.stderr)
	}
	return err
}

func (c *Command) run(cmd *exec.Cmd) error {
	cmd.Stdout = c.stdout
	cmd.Stderr = c.stderr
	return cmd.Run()
}

func hasDockerCompose() bool {
	_, err := exec.LookPath("docker-compose")
	return err == nil
}
