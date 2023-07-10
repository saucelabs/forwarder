// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package compose

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
		c: newCompose(),
	}
}

func (b *Builder) AddService(s ServiceBuilder) *Builder {
	if b.error == nil {
		b.error = b.c.addService(s.Service())
	}
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
