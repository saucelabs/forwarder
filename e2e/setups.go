// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"time"

	"github.com/saucelabs/forwarder/e2e/forwarder"
	"github.com/saucelabs/forwarder/e2e/setup"
	"github.com/saucelabs/forwarder/utils/compose"
)

func AllSetups() []setup.Setup {
	var all []setup.Setup
	all = append(all, DefaultSetups()...)
	all = append(all, UpstreamAuthSetups()...)
	all = append(all, PacSetups()...)
	all = append(all, LocalhostAllowSetup(), SC2450Setup())
	all = append(all, HeaderModifiersSetups()...)
	return all
}

func DefaultSetups() (ss []setup.Setup) {
	for _, httpbinScheme := range forwarder.HttpbinSchemes {
		for _, proxyScheme := range forwarder.ProxySchemes {
			ss = append(ss, setup.Setup{
				Name: "default-" + httpbinScheme + "-" + proxyScheme,
				Compose: compose.NewBuilder().
					AddService(
						forwarder.ProxyService().
							WithProtocol(proxyScheme).
							WithBasicAuth("u1:p1").
							WithGoleak()).
					AddService(
						forwarder.HttpbinService().
							WithProtocol(httpbinScheme)).
					MustBuild(),
			})
		}
	}
	return
}

func UpstreamAuthSetups() (ss []setup.Setup) {
	for _, httpbinScheme := range forwarder.HttpbinSchemes {
		ss = append(ss, setup.Setup{
			Name: "upstream-auth-" + httpbinScheme,
			Compose: compose.NewBuilder().
				AddService(
					forwarder.ProxyService().
						WithUpstream(forwarder.UpstreamProxyServiceName, "http").
						WithCredentials("u2:p2", forwarder.UpstreamProxyServiceName+":3128")).
				AddService(
					forwarder.UpstreamProxyService().
						WithBasicAuth("u2:p2")).
				AddService(
					forwarder.HttpbinService().
						WithProtocol(httpbinScheme)).
				MustBuild(),
		})
	}
	return
}

func PacSetups() []setup.Setup {
	return []setup.Setup{
		{
			Name: "pac-direct",
			Compose: compose.NewBuilder().
				AddService(
					forwarder.ProxyService().
						WithPac("./pac/direct.js")).
				AddService(
					forwarder.HttpbinService()).
				MustBuild(),
		},
		{
			Name: "pac-upstream",
			Compose: compose.NewBuilder().
				AddService(
					forwarder.ProxyService().
						WithPac("./pac/upstream.js")).
				AddService(
					forwarder.UpstreamProxyService()).
				AddService(
					forwarder.HttpbinService()).
				MustBuild(),
		},
		{
			Name: "pac-issue-184",
			Compose: compose.NewBuilder().
				AddService(
					forwarder.ProxyService().
						WithPac("./pac/issue-184.js")).
				AddService(
					forwarder.HttpbinService()).
				MustBuild(),
			Run: "GoogleCom",
		},
	}
}

func LocalhostAllowSetup() setup.Setup {
	return setup.Setup{
		Name: "localhost-allow",
		Compose: compose.NewBuilder().
			AddService(
				forwarder.ProxyService().
					WithLocalhostMode("allow")).
			MustBuild(),
		Run: "Localhost",
	}
}

func SC2450Setup() setup.Setup {
	return setup.Setup{
		Name: "sc-2450",
		Compose: compose.NewBuilder().
			AddService(
				forwarder.ProxyService().
					WithEnv("FORWARDER_SC2450", "go")).
			AddService(
				forwarder.HttpbinService()).
			AddService(&compose.Service{
				Name:    "sc-2450",
				Image:   "python:3",
				Command: "python /server.py",
				Volumes: []string{"./sc-2450/server.py:/server.py"},
				WaitFunc: func() error {
					time.Sleep(3 * time.Second)
					return nil
				},
			}).MustBuild(),
		Run: "SC2450",
	}
}

func HeaderModifiersSetups() []setup.Setup {
	return []setup.Setup{
		{
			Name: "header-mods",
			Compose: compose.NewBuilder().
				AddService(
					forwarder.ProxyService().
						WithEnv("FORWARDER_TEST_HEADERS", "test").
						WithEnv("FORWARDER_HEADER", "test-add:test-value,-test-rm,-rm-pref*,test-empty;")).
				AddService(
					forwarder.HttpbinService()).
				MustBuild(),
			Run: "HeaderMods",
		},
		{
			Name: "response-header-mods",
			Compose: compose.NewBuilder().
				AddService(
					forwarder.ProxyService().
						WithEnv("FORWARDER_TEST_RESPONSE_HEADERS", "test").
						WithEnv("FORWARDER_RESPONSE_HEADER", "test-resp-add:test-resp-value,-test-resp-rm,-resp-rm-pref*,test-resp-empty;")).
				AddService(
					forwarder.HttpbinService()).
				MustBuild(),
			Run: "ResponseHeaderMods",
		},
	}
}
