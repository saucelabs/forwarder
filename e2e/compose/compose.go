// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package compose

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Service struct {
	Name        string               `yaml:"-"`
	Image       string               `yaml:"image,omitempty"`
	Command     string               `yaml:"command,omitempty"`
	Environment map[string]string    `yaml:"environment,omitempty"`
	Ports       []string             `yaml:"ports,omitempty"`
	Volumes     []string             `yaml:"volumes,omitempty"`
	WaitFunc    func(*Service) error `yaml:"-"`
}

type ServiceOpt func(*Service)

func (svc *Service) Wait() error {
	if svc.WaitFunc != nil {
		return svc.WaitFunc(svc)
	}
	return nil
}

type Compose struct {
	Name     string              `yaml:"-"`
	Version  string              `yaml:"version"`
	Services map[string]*Service `yaml:"services,omitempty"`
	Path     string              `yaml:"-"`
	OnStart  func() error        `yaml:"-"`
	Debug    bool                `yaml:"-"`
}

type Opt func(*Compose)

func NewCompose(name string, opts ...Opt) *Compose {
	c := &Compose{Name: name}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Compose) AddService(name, image string, opts ...ServiceOpt) {
	svc := &Service{Name: name, Image: image, Environment: map[string]string{}}
	for _, opt := range opts {
		opt(svc)
	}
	if c.Services == nil {
		c.Services = map[string]*Service{}
	}
	c.Services[name] = svc
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

func (c *Compose) save(path string) error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return writeFile(path, b)
}

func (c *Compose) up() error {
	if c.Path == "" {
		c.Path = filepath.Join(os.TempDir(), "testrunner", "docker-compose.yaml")
	}
	if err := c.save(c.Path); err != nil {
		return fmt.Errorf("save compose %s at %s: %w", c.Name, c.Path, err)
	}
	cmd := exec.Command("docker", "compose", "-f", c.Path, "up", "-d", "--force-recreate", "--remove-orphans") //nolint:gosec // local usage only
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Printf("compose stdout: %s", stdout.String())
		log.Printf("compose stderr: %s", stderr.String())
		return fmt.Errorf("%s up: %w", c.Name, err)
	}
	for _, svc := range c.Services {
		if err := svc.Wait(); err != nil {
			return fmt.Errorf("%s wait for %s: %w", c.Name, svc.Name, err)
		}
	}
	return nil
}

func (c *Compose) down() error {
	cmd := exec.Command("docker", "compose", "-f", c.Path, "down", "-v", "--remove-orphans") //nolint:gosec // local usage only
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Printf("compose stdout: %s", stdout.String())
		log.Printf("compose stderr: %s", stderr.String())
		return fmt.Errorf("%s down: %w", c.Name, err)
	}
	return nil
}

func (c *Compose) Run(preserve bool) error {
	log.Printf("running %s", c.Name)
	if err := c.up(); err != nil {
		return err
	}
	if c.OnStart == nil {
		return fmt.Errorf("no OnStart function defined for %s", c.Name)
	}
	if err := c.OnStart(); err != nil {
		return err
	}
	if !preserve {
		if err := c.down(); err != nil {
			return err
		}
	}
	return nil
}
