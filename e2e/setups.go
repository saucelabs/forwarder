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
	var ss []setup.Setup
	ss = append(ss, DefaultsSetups()...)
	ss = append(ss, AuthSetups()...)
	ss = append(ss, PacSetups()...)

	ss = append(ss, FlagProxyLocalhost()...)
	ss = append(ss,
		FlagHeaderSetup(),
		FlagResponseHeaderSetup(),

		SC2450Setup(),
	)
	return ss
}

func DefaultsSetups() (ss []setup.Setup) {
	const run = "^TestProxy"
	for _, httpbinScheme := range forwarder.HttpbinSchemes {
		for _, proxyScheme := range forwarder.ProxySchemes {
			ss = append(ss,
				setup.Setup{
					Name: "defaults-" + httpbinScheme + "-" + proxyScheme,
					Compose: compose.NewBuilder().
						AddService(
							forwarder.HttpbinService().
								WithProtocol(httpbinScheme)).
						AddService(
							forwarder.ProxyService().
								WithProtocol(proxyScheme)).
						MustBuild(),
					Run: run,
				})
			for _, upstreamProxyScheme := range forwarder.ProxySchemes {
				ss = append(ss,
					setup.Setup{
						Name: "defaults-" + httpbinScheme + "-" + proxyScheme + "-" + upstreamProxyScheme,
						Compose: compose.NewBuilder().
							AddService(
								forwarder.HttpbinService().
									WithProtocol(httpbinScheme)).
							AddService(
								forwarder.ProxyService().
									WithProtocol(proxyScheme).
									WithUpstream(forwarder.UpstreamProxyServiceName, upstreamProxyScheme)).
							AddService(
								forwarder.UpstreamProxyService().
									WithProtocol(upstreamProxyScheme)).
							MustBuild(),
						Run: run,
					})
			}
		}
	}
	return
}

func AuthSetups() (ss []setup.Setup) {
	const run = "StatusCode|Auth"
	for _, httpbinScheme := range forwarder.HttpbinSchemes {
		ss = append(ss,
			setup.Setup{
				Name: "auth-" + httpbinScheme + "-http",
				Compose: compose.NewBuilder().
					AddService(
						forwarder.HttpbinService().
							WithProtocol(httpbinScheme)).
					AddService(
						forwarder.ProxyService().
							WithBasicAuth("u1:p1")).
					MustBuild(),
				Run: run,
			}, setup.Setup{
				Name: "auth-" + httpbinScheme + "-http-http",
				Compose: compose.NewBuilder().
					AddService(
						forwarder.HttpbinService().
							WithProtocol(httpbinScheme)).
					AddService(
						forwarder.ProxyService().
							WithBasicAuth("u1:p1").
							WithUpstream(forwarder.UpstreamProxyServiceName, "http").
							WithCredentials("u2:p2", forwarder.UpstreamProxyServiceName+":3128")).
					AddService(
						forwarder.UpstreamProxyService().
							WithBasicAuth("u2:p2")).
					MustBuild(),
				Run: run,
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
			Run: "^TestProxy",
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
			Run: "^TestProxy",
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
			Run: "^TestProxyGoogleCom$",
		},
	}
}

func FlagProxyLocalhost() (ss []setup.Setup) {
	for _, mode := range []string{"deny", "allow"} {
		ss = append(ss, setup.Setup{
			Name: "flag-proxy-localhost-" + mode,
			Compose: compose.NewBuilder().
				AddService(
					forwarder.ProxyService().
						WithLocalhostMode(mode)).
				MustBuild(),
			Run: "^TestFlagProxyLocalhost/" + mode + "$",
		})
	}
	return
}

func FlagHeaderSetup() setup.Setup {
	return setup.Setup{
		Name: "flag-header",
		Compose: compose.NewBuilder().
			AddService(
				forwarder.HttpbinService()).
			AddService(
				forwarder.ProxyService().
					WithHeader("test-add:test-value,-test-rm,-rm-pref*,test-empty;")).
			MustBuild(),
		Run: "^TestFlagHeader$",
	}
}

func FlagResponseHeaderSetup() setup.Setup {
	return setup.Setup{
		Name: "flag-response-header",
		Compose: compose.NewBuilder().
			AddService(
				forwarder.HttpbinService()).
			AddService(
				forwarder.ProxyService().
					WithResponseHeader("test-resp-add:test-resp-value,-test-resp-rm,-resp-rm-pref*,test-resp-empty;")).
			MustBuild(),
		Run: "^TestFlagResponseHeader$",
	}
}

func SC2450Setup() setup.Setup {
	return setup.Setup{
		Name: "sc-2450",
		Compose: compose.NewBuilder().
			AddService(
				forwarder.HttpbinService()).
			AddService(
				forwarder.ProxyService()).
			AddService(&compose.Service{
				Name:    "sc-2450",
				Image:   "python:3",
				Command: "python /server.py",
				Volumes: []string{"./sc-2450/server.py:/server.py"},
				HealthCheck: &compose.HealthCheck{
					StartPeriod: 3 * time.Second,
					Interval:    1 * time.Second,
					Retries:     1,
					Test:        []string{"CMD", "true"},
				},
			}).MustBuild(),
		Run: "^TestSC2450$",
	}
}
