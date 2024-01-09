// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"strings"
	"time"

	"github.com/saucelabs/forwarder/utils/compose"
)

type Service compose.Service

const (
	Image = "saucelabs/forwarder:${FORWARDER_VERSION}"

	ProxyServiceName         = "proxy"
	UpstreamProxyServiceName = "upstream-proxy"
	HttpbinServiceName       = "httpbin"
)

const enabled = "true"

func ProxyService() *Service {
	return &Service{
		Name:  ProxyServiceName,
		Image: Image,
		Environment: map[string]string{
			"FORWARDER_API_ADDRESS": ":10000",
		},
		Ports: []string{
			"3128:3128",
			"10000:10000",
		},
		HealthCheck: healthCheck(),
	}
}

func UpstreamProxyService() *Service {
	return &Service{
		Name:  UpstreamProxyServiceName,
		Image: Image,
		Environment: map[string]string{
			"FORWARDER_API_ADDRESS": ":10000",
			"FORWARDER_NAME":        UpstreamProxyServiceName,
		},
		Ports: []string{
			"10001:10000",
		},
		HealthCheck: healthCheck(),
	}
}

func HttpbinService() *Service {
	return &Service{
		Name:    HttpbinServiceName,
		Image:   Image,
		Command: "httpbin",
		Environment: map[string]string{
			"FORWARDER_API_ADDRESS": ":10000",
		},
		Ports: []string{
			"10002:10000",
		},
		HealthCheck: healthCheck(),
	}
}

func (s *Service) WithProtocol(protocol string) *Service {
	s.Environment["FORWARDER_PROTOCOL"] = protocol

	if protocol == "https" || protocol == "h2" {
		s.Environment["FORWARDER_TLS_CERT_FILE"] = "/etc/forwarder/certs/" + s.Name + ".crt"
		s.Environment["FORWARDER_TLS_KEY_FILE"] = "/etc/forwarder/private/" + s.Name + ".key"
		s.Volumes = append(s.Volumes,
			"./certs/"+s.Name+".crt:/etc/forwarder/certs/"+s.Name+".crt:ro",
			"./certs/"+s.Name+".key:/etc/forwarder/private/"+s.Name+".key:ro",
		)
	}

	return s
}

func (s *Service) WithSelfSigned(protocol string) *Service {
	s.Environment["FORWARDER_PROTOCOL"] = protocol
	return s
}

func (s *Service) Insecure() *Service {
	s.Environment["FORWARDER_INSECURE"] = enabled
	return s
}

func (s *Service) WithMITM() *Service {
	s.Environment["FORWARDER_MITM"] = enabled
	return s
}

func (s *Service) WithMITMCACert() *Service {
	s.Environment["FORWARDER_MITM_CACERT_FILE"] = "/etc/forwarder/certs/mitm-ca.crt"
	s.Environment["FORWARDER_MITM_CAKEY_FILE"] = "/etc/forwarder/private/mitm-ca.key"
	s.Volumes = append(s.Volumes,
		"./certs/ca.crt:/etc/forwarder/certs/mitm-ca.crt:ro",
		"./certs/ca.key:/etc/forwarder/private/mitm-ca.key:ro",
	)

	return s
}

func (s *Service) WithMITMDomains(domains ...string) *Service {
	s.Environment["FORWARDER_MITM_DOMAINS"] = strings.Join(domains, ",")
	return s
}

func (s *Service) WithUpstream(name, protocol string) *Service {
	s.Environment["FORWARDER_PROXY"] = protocol + "://" + name + ":3128"
	if protocol == "https" {
		s.Environment["FORWARDER_CACERT_FILE"] = "/etc/forwarder/certs/ca-certificates.crt"
		s.Volumes = append(s.Volumes, "./certs/ca.crt:/etc/forwarder/certs/ca-certificates.crt:ro")
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

func (s *Service) WithHeader(header string) *Service {
	s.Environment["FORWARDER_HEADER"] = header
	return s
}

func (s *Service) WithResponseHeader(header string) *Service {
	s.Environment["FORWARDER_RESPONSE_HEADER"] = header
	return s
}

func (s *Service) WithGoleak() *Service {
	s.Environment["FORWARDER_GOLEAK"] = enabled
	return s
}

func (s *Service) WithEnv(key, val string) *Service {
	s.Environment[key] = val
	return s
}

func (s *Service) WithIP(network, ipv4 string) *Service {
	s.Network = map[string]compose.ServiceNetwork{
		network: {IPv4: ipv4},
	}
	return s
}

func (s *Service) WithDNSServer(servers ...string) *Service {
	s.Environment["FORWARDER_DNS_SERVER"] = strings.Join(servers, ",")
	return s
}

func (s *Service) WithDNSTimeout(timeout time.Duration) *Service {
	s.Environment["FORWARDER_DNS_TIMEOUT"] = timeout.String()
	return s
}

func (s *Service) WithHTTPDialTimeout(timeout time.Duration) *Service {
	s.Environment["FORWARDER_HTTP_DIAL_TIMEOUT"] = timeout.String()
	return s
}

func (s *Service) WithDenyDomains(domains ...string) *Service {
	s.Environment["FORWARDER_DENY_DOMAINS"] = strings.Join(domains, ",")
	return s
}

func (s *Service) WithDirectDomains(domains ...string) *Service {
	s.Environment["FORWARDER_DIRECT_DOMAINS"] = strings.Join(domains, ",")
	return s
}

func (s *Service) WithReadLimit(limit string) *Service {
	s.Environment["FORWARDER_READ_LIMIT"] = limit
	return s
}

func (s *Service) WithWriteLimit(limit string) *Service {
	s.Environment["FORWARDER_WRITE_LIMIT"] = limit
	return s
}

func (s *Service) Service() *compose.Service {
	return (*compose.Service)(s)
}

func healthCheck() *compose.HealthCheck {
	return &compose.HealthCheck{
		Test:     []string{"CMD", "forwarder", "ready"},
		Interval: time.Second,
		Timeout:  3 * time.Second,
		Retries:  10,
	}
}
