// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package setups

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/saucelabs/forwarder/e2e/compose"
	. "github.com/saucelabs/forwarder/e2e/compose/opts"
)

func MakeTest(run string) compose.Opt {
	return func(c *compose.Compose) {
		c.OnStart = func() error {
			cmd := exec.Command("make", "test")
			if run != "" {
				cmd.Env = append(os.Environ(), "RUN="+run)
			}
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				log.Printf("stdout: %s", stdout.String())
				log.Printf("stderr: %s", stderr.String())
				return err
			}

			if stderr.Len() > 0 {
				log.Printf("%s", stderr.String())
			}
			if c.Debug {
				log.Printf("%s", stdout.String())
			} else {
				s := strings.Split(stdout.String(), "\n")
				for _, l := range s {
					if strings.HasPrefix(l, "---") {
						log.Printf("%s", l)
					}
				}
			}

			return nil
		}
	}
}

func NewCompose(name string, opts ...compose.Opt) *compose.Compose {
	defaultOpts := []compose.Opt{
		WithComposePath("docker-compose.yaml"),
		WithVersion("3.8"),
		MakeTest(""),
	}
	opts = append(defaultOpts, opts...)

	return compose.NewCompose(name, opts...)
}

var (
	allHttpbinSchemes = []string{"http", "https", "h2"}
	allProxySchemes   = []string{"http", "https"}
)

func All() []*compose.Compose {
	var all []*compose.Compose
	all = append(all, Standard()...)
	all = append(all, UpstreamAuth()...)
	all = append(all, Pacs()...)
	all = append(all, LocalhostAllow(), SC2450())
	all = append(all, HeaderModifiers()...)
	return all
}

func Standard() []*compose.Compose {
	var cs []*compose.Compose
	for _, httpbinScheme := range allHttpbinSchemes {
		for _, proxyScheme := range allProxySchemes {
			cs = append(cs, NewCompose("default-"+httpbinScheme+"-"+proxyScheme,
				ProxyService(
					WithProtocol(proxyScheme),
					WithBasicAuth("u1:p1"),
					WithGoleak(),
				),
				HttpbinService(
					WithProtocol(httpbinScheme),
				),
			))
			for _, upstreamScheme := range allProxySchemes {
				cs = append(cs, NewCompose("default-"+httpbinScheme+"-"+proxyScheme+"-"+upstreamScheme,
					ProxyService(
						WithProtocol(proxyScheme),
						WithBasicAuth("u1:p1"),
						WithUpstream(UpstreamServiceName, upstreamScheme),
						WithGoleak(),
					),
					UpstreamService(
						WithProtocol(upstreamScheme),
					),
					HttpbinService(
						WithProtocol(httpbinScheme),
					),
				))
			}
		}
	}
	return cs
}

func UpstreamAuth() []*compose.Compose {
	var cs []*compose.Compose
	for _, httpbinScheme := range allHttpbinSchemes {
		cs = append(cs, NewCompose("upstream-auth-"+httpbinScheme,
			ProxyService(
				WithUpstream(UpstreamServiceName, "http"),
				WithCredentials("u2:p2", UpstreamServiceName+":3128")),
			UpstreamService(
				WithBasicAuth("u2:p2"),
			),
			HttpbinService(
				WithProtocol(httpbinScheme),
			),
		))
	}
	return cs
}

func Pacs() []*compose.Compose {
	return []*compose.Compose{
		NewCompose("pac-direct",
			ProxyService(WithPac("./pac/direct.js")),
			HttpbinService(),
		),
		NewCompose("pac-upstream",
			ProxyService(WithPac("./pac/upstream.js")),
			UpstreamService(),
			HttpbinService(),
		),
		NewCompose("pac-issue-184",
			ProxyService(WithPac("./pac/issue-184.js")),
			HttpbinService(),
			MakeTest("GoogleCom"),
		),
	}
}

func LocalhostAllow() *compose.Compose {
	return NewCompose("localhost-allow",
		ProxyService(WithLocalhostMode("allow")),
		HttpbinService(),
		MakeTest("Localhost"),
	)
}

func SC2450() *compose.Compose {
	return NewCompose("sc-2450",
		ProxyService(func(s *compose.Service) {
			s.Environment["FORWARDER_SC2450"] = "go"
		}),
		HttpbinService(),
		func(c *compose.Compose) {
			c.AddService("sc-2450", "python:3",
				WithCommand("python /server.py"), WithVolume("./sc-2450/server.py:/server.py"),
				WithWaitFunc(func(s *compose.Service) error {
					time.Sleep(3 * time.Second)
					return nil
				}),
			)
		},
		MakeTest("SC2450"),
	)
}

func HeaderModifiers() []*compose.Compose {
	return []*compose.Compose{
		NewCompose("header-mods",
			ProxyService(func(s *compose.Service) {
				s.Environment["FORWARDER_TEST_HEADERS"] = "test"
				s.Environment["FORWARDER_HEADER"] = "test-add:test-value,-test-rm,-rm-pref*,test-empty;"
			}),
			HttpbinService(),
			MakeTest("HeaderMods"),
		),
		NewCompose("response-header-mods",
			ProxyService(func(s *compose.Service) {
				s.Environment["FORWARDER_TEST_RESPONSE_HEADERS"] = "test"
				s.Environment["FORWARDER_RESPONSE_HEADER"] = "test-resp-add:test-resp-value,-test-resp-rm,-resp-rm-pref*,test-resp-empty;"
			}),
			HttpbinService(),
			MakeTest("HeaderRespMods"),
		),
	}
}
