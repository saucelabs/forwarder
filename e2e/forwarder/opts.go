// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"github.com/saucelabs/forwarder/e2e/compose"
)

func WithCommand(command string) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Command = command
	}
}

func WithProtocol(protocol string) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Environment["FORWARDER_PROTOCOL"] = protocol
	}
}

func WithUpstream(name, protocol string) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Environment["FORWARDER_PROXY"] = protocol + "://" + name + ":3128"
		if protocol == "https" {
			s.Environment["FORWARDER_INSECURE"] = "true"
		}
	}
}

func WithBasicAuth(auth string) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Environment["FORWARDER_BASIC_AUTH"] = auth
	}
}

func WithCredentials(credentials, address string) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Environment["FORWARDER_CREDENTIALS"] = credentials + "@" + address
	}
}

func WithPac(pac string) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Environment["FORWARDER_PAC"] = "/pac.js"
		s.Volumes = append(s.Volumes, pac+":/pac.js")
	}
}

func WithLocalhostMode(mode string) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Environment["FORWARDER_PROXY_LOCALHOST"] = mode
	}
}

func WithPorts(ports ...string) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Ports = append(s.Ports, ports...)
	}
}

func WithVolume(volume string) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Volumes = append(s.Volumes, volume)
	}
}

func WithAPIAddress(address string) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Environment["FORWARDER_API_ADDRESS"] = address
	}
}

func WithGoleak() compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Environment["FORWARDER_GOLEAK"] = "true"
	}
}

func WithWaitFunc(f func(*compose.Service) error) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.WaitFunc = f
	}
}

func WithComposePath(path string) compose.Opt {
	return func(c *compose.Compose) {
		c.Path = path
	}
}

func WithVersion(version string) compose.Opt {
	return func(c *compose.Compose) {
		c.Version = version
	}
}
