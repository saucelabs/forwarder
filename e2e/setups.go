// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"time"

	"github.com/saucelabs/forwarder/e2e/dns"
	"github.com/saucelabs/forwarder/e2e/forwarder"
	sc2450 "github.com/saucelabs/forwarder/e2e/sc-2450"
	"github.com/saucelabs/forwarder/e2e/setup"
	"github.com/saucelabs/forwarder/utils/compose"
)

type setupList struct {
	s []setup.Setup
}

func (l *setupList) Add(s ...setup.Setup) {
	l.s = append(l.s, s...)
}

func (l *setupList) Build() []setup.Setup {
	return l.s
}

func AllSetups() []setup.Setup {
	l := &setupList{}

	SetupDefaults(l)
	SetupAuth(l)
	SetupPac(l)
	SetupGRPC(l)
	SetupFlagProxyLocalhost(l)
	SetupFlagHeader(l)
	SetupFlagResponseHeader(l)
	SetupFlagConnectHeader(l)
	SetupFlagDNSServer(l)
	SetupFlagInsecure(l)
	SetupFlagMITMCACert(l)
	SetupFlagMITMGenCA(l)
	SetupFlagMITMDomains(l)
	SetupFlagDenyDomains(l)
	SetupFlagDirectDomains(l)
	SetupFlagRateLimit(l)
	SetupSC2450(l)

	return l.Build()
}

func SetupDefaults(l *setupList) {
	const run = "^TestProxy"
	for _, httpbinScheme := range forwarder.HttpbinSchemes {
		for _, proxyScheme := range forwarder.ProxySchemes {
			l.Add(setup.Setup{
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
				l.Add(setup.Setup{
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
}

func SetupAuth(l *setupList) {
	const run = "StatusCode|Auth"
	for _, httpbinScheme := range forwarder.HttpbinSchemes {
		l.Add(
			setup.Setup{
				Name: "auth-" + httpbinScheme + "-http",
				Compose: compose.NewBuilder().
					AddService(
						forwarder.HttpbinService().
							WithProtocol(httpbinScheme)).
					AddService(
						forwarder.ProxyService().
							WithBasicAuth("u1:p1!#$%^&*()_{};'?")).
					MustBuild(),
				Run: run,
			},
			setup.Setup{
				Name: "auth-" + httpbinScheme + "-http-http",
				Compose: compose.NewBuilder().
					AddService(
						forwarder.HttpbinService().
							WithProtocol(httpbinScheme)).
					AddService(
						forwarder.ProxyService().
							WithBasicAuth("u1:p1!#$%^&*()_{};'?").
							WithUpstream(forwarder.UpstreamProxyServiceName, "http").
							WithCredentials("u2:p2!#$%^&*()_{};'?", forwarder.UpstreamProxyServiceName+":3128")).
					AddService(
						forwarder.UpstreamProxyService().
							WithBasicAuth("u2:p2!#$%^&*()_{};'?")).
					MustBuild(),
				Run: run,
			},
		)
	}
}

func SetupPac(l *setupList) {
	l.Add(
		setup.Setup{
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
		setup.Setup{
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
		setup.Setup{
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
	)
}

func SetupGRPC(l *setupList) {
	l.Add(
		setup.Setup{
			Name: "grpc",
			Compose: compose.NewBuilder().
				AddService(forwarder.ProxyService()).
				AddService(forwarder.GRPCTestService()).
				MustBuild(),
			Run: "^TestGRPC",
		},
	)
}

func SetupFlagProxyLocalhost(l *setupList) {
	for _, mode := range []string{"deny", "allow"} {
		l.Add(setup.Setup{
			Name: "flag-proxy-localhost-" + mode,
			Compose: compose.NewBuilder().
				AddService(
					forwarder.ProxyService().
						WithLocalhostMode(mode)).
				MustBuild(),
			Run: "^TestFlagProxyLocalhost/" + mode + "$",
		})
	}
}

func SetupFlagHeader(l *setupList) {
	l.Add(setup.Setup{
		Name: "flag-header",
		Compose: compose.NewBuilder().
			AddService(
				forwarder.HttpbinService()).
			AddService(
				forwarder.ProxyService().
					WithHeader("test-add:test-value,-test-rm,-rm-pref*,test-empty;")).
			MustBuild(),
		Run: "^TestFlagHeader$",
	})
}

func SetupFlagResponseHeader(l *setupList) {
	l.Add(setup.Setup{
		Name: "flag-response-header",
		Compose: compose.NewBuilder().
			AddService(
				forwarder.HttpbinService()).
			AddService(
				forwarder.ProxyService().
					WithResponseHeader("test-resp-add:test-resp-value,-test-resp-rm,-resp-rm-pref*,test-resp-empty;")).
			MustBuild(),
		Run: "^TestFlagResponseHeader$",
	})
}

func SetupFlagConnectHeader(l *setupList) {
	l.Add(setup.Setup{
		Name: "flag-connect-header",
		Compose: compose.NewBuilder().
			AddService(
				forwarder.HttpbinService().
					WithProtocol("https")).
			AddService(
				forwarder.ProxyService().
					WithUpstream(forwarder.UpstreamProxyServiceName, "https").
					WithConnectHeader("Proxy-Authorization:Basic dXNlcjpjb25uZWN0LWhlYWRlcg==")).
			AddService(forwarder.UpstreamProxyService().
				WithProtocol("https").
				WithBasicAuth("user:connect-header")).
			MustBuild(),
		Run: "^TestFlagConnectHeader$",
	})
}

func SetupFlagDNSServer(l *setupList) {
	const (
		networkName = "internal"

		dnsIPAddr        = "150.150.100.2"
		invalidDNSIPAddr = "150.150.100.3"
		httpbinIPAddr    = "150.150.100.10"
		proxyIPAddr      = "150.150.100.11"
	)
	for _, s := range []struct {
		name    string
		servers []string
	}{
		{
			name:    "flag-dns-server",
			servers: []string{dnsIPAddr},
		},
		{
			name:    "flag-dns-fallback",
			servers: []string{invalidDNSIPAddr, dnsIPAddr},
		},
	} {
		l.Add(setup.Setup{
			Name: s.name,
			Compose: compose.NewBuilder().
				AddService(
					forwarder.HttpbinService().
						WithIP(networkName, httpbinIPAddr)).
				AddService(
					forwarder.ProxyService().
						WithIP(networkName, proxyIPAddr).
						WithDNSServer(s.servers...).
						WithDNSTimeout(1 * time.Second)).
				AddService(
					dns.Service(networkName, dnsIPAddr)).
				AddNetwork(&compose.Network{
					Name:   networkName,
					Driver: "bridge",
					IPAM: compose.IPAM{
						Config: []compose.IPAMConfig{
							{
								Subnet:  "150.150.100.0/24",
								Gateway: "150.150.100.1",
								IPRange: "150.150.100.10/29",
							},
						},
					},
				}).
				MustBuild(),
			Run: "^TestFlagDNSServer$",
		})
	}
}

func SetupFlagInsecure(l *setupList) {
	l.Add(
		setup.Setup{
			Name: "flag-insecure-true",
			Compose: compose.NewBuilder().
				AddService(
					forwarder.HttpbinService()).
				AddService(
					forwarder.ProxyService().
						WithUpstream(forwarder.UpstreamProxyServiceName, "https").
						Insecure()).
				AddService(
					forwarder.UpstreamProxyService().
						WithSelfSigned("https")).
				MustBuild(),
			Run: "^TestFlagInsecure/true$",
		},
		setup.Setup{
			Name: "flag-insecure-false",
			Compose: compose.NewBuilder().
				AddService(
					forwarder.HttpbinService()).
				AddService(
					forwarder.ProxyService().
						WithUpstream(forwarder.UpstreamProxyServiceName, "https")).
				AddService(
					forwarder.UpstreamProxyService().
						WithSelfSigned("https")).
				MustBuild(),
			Run: "^TestFlagInsecure/false$",
		},
	)
}

func SetupFlagMITMCACert(l *setupList) {
	const run = "^Test(FlagMITM|Proxy.*)$"

	l.Add(setup.Setup{
		Name: "flag-mitm-cacert",
		Compose: compose.NewBuilder().
			AddService(
				forwarder.HttpbinService().WithSelfSigned("https")).
			AddService(
				forwarder.ProxyService().
					WithResponseHeader("test-resp-add:test-resp-value").
					WithMITMCACert().
					Insecure()).
			MustBuild(),
		Run: run,
	})

	for _, upstreamProxyScheme := range forwarder.ProxySchemes {
		l.Add(setup.Setup{
			Name: "flag-mitm-cacert" + "-" + upstreamProxyScheme,
			Compose: compose.NewBuilder().
				AddService(
					forwarder.HttpbinService().WithSelfSigned("https")).
				AddService(
					forwarder.ProxyService().
						WithResponseHeader("test-resp-add:test-resp-value").
						WithMITMCACert().
						Insecure().
						WithUpstream(forwarder.UpstreamProxyServiceName, upstreamProxyScheme)).
				AddService(
					forwarder.UpstreamProxyService().
						WithProtocol(upstreamProxyScheme)).
				MustBuild(),
			Run: run,
		})
	}
}

func SetupFlagMITMGenCA(l *setupList) {
	const run = "^TestFlagMITMGenCA$"

	l.Add(setup.Setup{
		Name: "flag-mitm-genca",
		Compose: compose.NewBuilder().
			AddService(
				forwarder.HttpbinService().WithSelfSigned("https")).
			AddService(
				forwarder.ProxyService().
					WithResponseHeader("test-resp-add:test-resp-value").
					WithMITM().
					Insecure()).
			MustBuild(),
		Run: run,
	})
}

func SetupFlagMITMDomains(l *setupList) {
	const run = "^TestFlagMITMDomains$"

	l.Add(setup.Setup{
		Name: "flag-mitm-domains",
		Compose: compose.NewBuilder().
			AddService(
				forwarder.HttpbinService().WithProtocol("https")).
			AddService(
				forwarder.ProxyService().
					WithResponseHeader("test-resp-add:test-resp-value").
					WithMITMCACert().
					WithMITMDomains("google", "httpbin", "-httpbin").
					Insecure()).
			MustBuild(),
		Run: run,
	})
}

func SetupFlagDenyDomains(l *setupList) {
	const run = "^TestFlagDenyDomains$"

	l.Add(
		setup.Setup{
			Name: "flag-deny-domains",
			Compose: compose.NewBuilder().
				AddService(
					forwarder.HttpbinService()).
				AddService(
					forwarder.ProxyService().
						WithDenyDomains("\\.com")).
				MustBuild(),
			Run: run,
		},
		setup.Setup{
			Name: "flag-deny-domains-exclude",
			Compose: compose.NewBuilder().
				AddService(
					forwarder.HttpbinService()).
				AddService(
					forwarder.ProxyService().
						WithDenyDomains("google", "httpbin", "-httpbin")).
				MustBuild(),
			Run: run,
		},
	)
}

func SetupFlagDirectDomains(l *setupList) {
	for _, scheme := range forwarder.HttpbinSchemes {
		l.Add(
			setup.Setup{
				Name: "flag-direct-domains-" + scheme,
				Compose: compose.NewBuilder().
					AddService(
						forwarder.HttpbinService().
							WithProtocol(scheme)).
					AddService(
						forwarder.ProxyService().
							WithUpstream(forwarder.UpstreamProxyServiceName, "http").
							WithDirectDomains("httpbin")).
					AddService(
						forwarder.UpstreamProxyService()).
					MustBuild(),
				Run: "^TestFlagDirectDomains$",
			},
		)
	}
}

func SetupFlagRateLimit(l *setupList) {
	l.Add(
		setup.Setup{
			Name: "flag-read-write-limit",
			Compose: compose.NewBuilder().
				AddService(
					forwarder.HttpbinService()).
				AddService(
					forwarder.ProxyService().
						WithReadLimit("1M").
						WithWriteLimit("1M")).
				MustBuild(),
			Run: "^TestFlag(Read|Write)Limit$",
		},
	)
}

func SetupSC2450(l *setupList) {
	l.Add(setup.Setup{
		Name: "sc-2450",
		Compose: compose.NewBuilder().
			AddService(
				forwarder.HttpbinService()).
			AddService(
				forwarder.ProxyService()).
			AddService(sc2450.Service()).
			MustBuild(),
		Run: "^TestSC2450$",
	})
}
