// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package compose

import (
	"os"
	"path"
	"strings"
)

type ServiceBuilder interface {
	Service() *Service
}

func (s *Service) Service() *Service {
	return s
}

type Builder struct {
	c     *Compose
	error error
}

func NewBuilder() *Builder {
	return &Builder{
		c: New(),
	}
}

func (b *Builder) AddService(sb ServiceBuilder) *Builder {
	if b.error != nil {
		return b
	}

	s := sb.Service()

	for i, v := range s.Volumes {
		s.Volumes[i] = absVolume(v)
	}

	b.error = b.c.AddService(s)

	return b
}

func absVolume(v string) string {
	a := strings.Split(v, ":")
	if path.IsAbs(a[0]) {
		return v
	}

	a[0] = path.Join(curDir(), a[0])
	return strings.Join(a, ":")
}

func curDir() string {
	d, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return d
}

func (b *Builder) AddNetwork(n *Network) *Builder {
	if b.error != nil {
		return b
	}

	b.error = b.c.AddNetwork(n)

	return b
}

func (b *Builder) Build() (*Compose, error) {
	if b.error != nil {
		return nil, b.error
	}
	return b.c, nil
}

func (b *Builder) MustBuild() *Compose {
	c, err := b.Build()
	if err != nil {
		panic(err)
	}
	return c
}
