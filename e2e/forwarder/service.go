// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"github.com/saucelabs/forwarder/utils/compose"
	"github.com/saucelabs/forwarder/utils/wait"
)

type Service compose.Service

const (
	Image = "saucelabs/forwarder:${FORWARDER_VERSION}"

	ProxyServiceName         = "proxy"
	UpstreamProxyServiceName = "upstream-proxy"
	HttpbinServiceName       = "httpbin"
)

func ProxyService() *Service {
	s := &Service{
		Name:        ProxyServiceName,
		Image:       Image,
		Environment: make(map[string]string),
		Ports: []string{
			"3128:3128",
			"10000:10000",
		},
		WaitFunc: func() error {
			return wait.ServerReady("http://localhost:10000")
		},
	}

	return s.WithAPIAddress(":10000")
}

func UpstreamProxyService() *Service {
	s := &Service{
		Name:        UpstreamProxyServiceName,
		Image:       Image,
		Environment: make(map[string]string),
		Ports: []string{
			"10001:10001",
		},
		WaitFunc: func() error {
			return wait.ServerReady("http://localhost:10001")
		},
	}

	return s.WithAPIAddress(":10001")
}

func HttpbinService() *Service {
	s := &Service{
		Name:        HttpbinServiceName,
		Image:       Image,
		Command:     "httpbin",
		Environment: make(map[string]string),
		Ports: []string{
			"10002:10002",
		},
		WaitFunc: func() error {
			return wait.ServerReady("http://localhost:10002")
		},
	}

	return s.WithAPIAddress(":10002")
}

func (s *Service) WithProtocol(protocol string) *Service {
	s.Environment["FORWARDER_PROTOCOL"] = protocol
	return s
}

func (s *Service) WithUpstream(name, protocol string) *Service {
	s.Environment["FORWARDER_PROXY"] = protocol + "://" + name + ":3128"
	if protocol == "https" {
		s.Environment["FORWARDER_INSECURE"] = "true"
	}
	return s
}

func (s *Service) WithBasicAuth(auth string) *Service {
	s.Environment["FORWARDER_BASIC_AUTH"] = auth
	return s
}

func (s *Service) WithCredentials(credentials, address string) *Service {
	s.Environment["FORWARDER_CREDENTIALS"] = credentials + "@" + address
	return s
}

func (s *Service) WithPac(pac string) *Service {
	s.Environment["FORWARDER_PAC"] = "/pac.js"
	s.Volumes = append(s.Volumes, pac+":/pac.js")
	return s
}

func (s *Service) WithLocalhostMode(mode string) *Service {
	s.Environment["FORWARDER_PROXY_LOCALHOST"] = mode
	return s
}

func (s *Service) WithAPIAddress(address string) *Service {
	s.Environment["FORWARDER_API_ADDRESS"] = address
	return s
}

func (s *Service) WithHeader(header string) *Service {
	s.Environment["FORWARDER_HEADER"] = header
	return s
}

func (s *Service) WithResponseHeader(header string) *Service {
	s.Environment["FORWARDER_RESPONSE_HEADER"] = header
	return s
}

func (s *Service) WithGoleak() *Service {
	s.Environment["FORWARDER_GOLEAK"] = "true"
	return s
}

func (s *Service) WithEnv(key, val string) *Service {
	s.Environment[key] = val
	return s
}

func (s *Service) Service() *compose.Service {
	return (*compose.Service)(s)
}
