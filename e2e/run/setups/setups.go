// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package setups

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/saucelabs/forwarder/e2e/compose"
)

const (
	ProxyService    = "proxy"
	UpstreamService = "upstream-proxy"
	HttpbinService  = "httpbin"
	forwarderImage  = "saucelabs/forwarder:${FORWARDER_VERSION}"
)

func All() []*compose.Compose {
	var all []*compose.Compose
	all = append(all, standard()...)
	all = append(all, standardUpstream()...)
	all = append(all, upstreamAuth()...)
	all = append(all, pacs()...)
	all = append(all, localhostAllow(), sc2450())
	all = append(all, headerModifiers()...)
	return all
}

func withCommand(command string) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Command = command
	}
}

func withProtocol(protocol string) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Environment["FORWARDER_PROTOCOL"] = protocol
	}
}

func withUpstream(name, protocol string) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Environment["FORWARDER_PROXY"] = protocol + "://" + name + ":3128"
		if protocol == "https" {
			s.Environment["FORWARDER_INSECURE"] = "true"
		}
	}
}

func withBasicAuth(auth string) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Environment["FORWARDER_BASIC_AUTH"] = auth
	}
}

func withCredentials(credentials, address string) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Environment["FORWARDER_CREDENTIALS"] = credentials + "@" + address
	}
}

func withPac(pac string) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Environment["FORWARDER_PAC"] = "/pac.js"
		s.Volumes = append(s.Volumes, pac+":/pac.js")
	}
}

func withLocalhostMode(mode string) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Environment["FORWARDER_PROXY_LOCALHOST"] = mode
	}
}

func withPorts(ports ...string) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Ports = append(s.Ports, ports...)
	}
}

func withVolume(volume string) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Volumes = append(s.Volumes, volume)
	}
}

func withAPIAddress(address string) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Environment["FORWARDER_API_ADDRESS"] = address
	}
}

func withGoleak() compose.ServiceOpt {
	return func(s *compose.Service) {
		s.Environment["FORWARDER_GOLEAK"] = "true"
	}
}

func withWaitFunc(f func(*compose.Service) error) compose.ServiceOpt {
	return func(s *compose.Service) {
		s.WaitFunc = f
	}
}

func withProxyService(opts ...compose.ServiceOpt) compose.Opt {
	defaultOpts := []compose.ServiceOpt{
		withProtocol("http"),
		withAPIAddress(":10000"),
		withPorts("3128:3128", "10000:10000"),
		withWaitFunc(func(s *compose.Service) error {
			return WaitForServerReady("http://localhost:10000")
		}),
	}
	opts = append(defaultOpts, opts...)
	return func(c *compose.Compose) {
		c.AddService(ProxyService, forwarderImage, opts...)
	}
}

func withHttpbinService(opts ...compose.ServiceOpt) compose.Opt {
	defaultOpts := []compose.ServiceOpt{
		withProtocol("http"),
		withCommand("httpbin"),
		withAPIAddress(":10000"),
		withPorts("10010:10000"),
		withWaitFunc(func(s *compose.Service) error {
			return WaitForServerReady("http://localhost:10010")
		}),
	}
	opts = append(defaultOpts, opts...)
	return func(c *compose.Compose) {
		c.AddService(HttpbinService, forwarderImage, opts...)
	}
}

func withUpstreamService(opts ...compose.ServiceOpt) compose.Opt {
	defaultOpts := []compose.ServiceOpt{
		withProtocol("http"),
		withAPIAddress(":10000"),
		withPorts("10020:10000"),
		withWaitFunc(func(s *compose.Service) error {
			return WaitForServerReady("http://localhost:10020")
		}),
	}
	opts = append(defaultOpts, opts...)
	return func(c *compose.Compose) {
		c.AddService(UpstreamService, forwarderImage, opts...)
	}
}

func WithOnStartMakeTest(run string) compose.Opt {
	return func(c *compose.Compose) {
		c.OnStart = func() error {
			cmd := exec.Command("make", "test")
			if run != "" {
				cmd.Env = append(os.Environ(), "RUN="+run)
			}
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err := cmd.Run()
			if err != nil {
				log.Printf("stdout: %s", stdout.String())
				log.Printf("stderr: %s", stderr.String())
			} else {
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
			}
			return err
		}
	}
}

func withComposePath(path string) compose.Opt {
	return func(c *compose.Compose) {
		c.Path = path
	}
}

func withVersion(version string) compose.Opt {
	return func(c *compose.Compose) {
		c.Version = version
	}
}

func newCompose(name string, opts ...compose.Opt) *compose.Compose {
	defaultOpts := []compose.Opt{
		withComposePath("docker-compose.yaml"),
		withVersion("3.8"),
		WithOnStartMakeTest(""),
	}
	opts = append(defaultOpts, opts...)
	return compose.NewCompose(name, opts...)
}

var (
	allHttpbinSchemes = []string{"http", "https", "h2"}
	allProxySchemes   = []string{"http", "https"}
)

func standard() []*compose.Compose {
	var cs []*compose.Compose
	for _, httpbinScheme := range allHttpbinSchemes {
		for _, proxyScheme := range allProxySchemes {
			cs = append(cs, newCompose("default-"+httpbinScheme+"-"+proxyScheme,
				withProxyService(
					withProtocol(proxyScheme),
					withBasicAuth("u1:p1"),
					withGoleak(),
				),
				withHttpbinService(
					withProtocol(httpbinScheme),
				),
			))
		}
	}
	return cs
}

func standardUpstream() []*compose.Compose {
	var cs []*compose.Compose
	for _, httpbinScheme := range allHttpbinSchemes {
		for _, proxyScheme := range allProxySchemes {
			for _, upstreamScheme := range allProxySchemes {
				cs = append(cs, newCompose("default-"+httpbinScheme+"-"+proxyScheme+"-"+upstreamScheme,
					withProxyService(
						withProtocol(proxyScheme),
						withBasicAuth("u1:p1"),
						withUpstream(UpstreamService, upstreamScheme),
						withGoleak(),
					),
					withUpstreamService(
						withProtocol(upstreamScheme),
					),
					withHttpbinService(
						withProtocol(httpbinScheme),
					),
				))
			}
		}
	}
	return cs
}

func upstreamAuth() []*compose.Compose {
	var cs []*compose.Compose
	for _, httpbinScheme := range allHttpbinSchemes {
		cs = append(cs, newCompose("upstream-auth-"+httpbinScheme,
			withProxyService(
				withUpstream(UpstreamService, "http"),
				withCredentials("u2:p2", UpstreamService+":3128")),
			withUpstreamService(
				withBasicAuth("u2:p2"),
			),
			withHttpbinService(
				withProtocol(httpbinScheme),
			),
		))
	}
	return cs
}

func pacs() []*compose.Compose {
	return []*compose.Compose{
		newCompose("pac-direct",
			withProxyService(withPac("./pac/direct.js")),
			withHttpbinService(),
		),
		newCompose("pac-upstream",
			withProxyService(withPac("./pac/upstream.js")),
			withUpstreamService(),
			withHttpbinService(),
		),
		newCompose("pac-issue-184",
			withProxyService(withPac("./pac/issue-184.js")),
			withHttpbinService(),
			WithOnStartMakeTest("GoogleCom"),
		),
	}
}

func localhostAllow() *compose.Compose {
	return newCompose("localhost-allow",
		withProxyService(withLocalhostMode("allow")),
		withHttpbinService(),
		WithOnStartMakeTest("Localhost"),
	)
}

func sc2450() *compose.Compose {
	return newCompose("sc-2450",
		withProxyService(func(s *compose.Service) {
			s.Environment["FORWARDER_SC2450"] = "go"
		}),
		withHttpbinService(),
		func(c *compose.Compose) {
			c.AddService("sc-2450", "python:3",
				withCommand("python /server.py"), withVolume("./sc-2450/server.py:/server.py"),
				withWaitFunc(func(s *compose.Service) error {
					time.Sleep(3 * time.Second)
					return nil
				}),
			)
		},
		WithOnStartMakeTest("SC2450"),
	)
}

func headerModifiers() []*compose.Compose {
	return []*compose.Compose{
		newCompose("header-mods",
			withProxyService(func(s *compose.Service) {
				s.Environment["FORWARDER_TEST_HEADERS"] = "test"
				s.Environment["FORWARDER_HEADER"] = "test-add:test-value,-test-rm,-rm-pref*,test-empty;"
			}),
			withHttpbinService(),
			WithOnStartMakeTest("HeaderMods"),
		),
		newCompose("response-header-mods",
			withProxyService(func(s *compose.Service) {
				s.Environment["FORWARDER_TEST_RESPONSE_HEADERS"] = "test"
				s.Environment["FORWARDER_RESPONSE_HEADER"] = "test-resp-add:test-resp-value,-test-resp-rm,-resp-rm-pref*,test-resp-empty;"
			}),
			withHttpbinService(),
			WithOnStartMakeTest("HeaderRespMods"),
		),
	}
}

// WaitForServerReady checks the API server /readyz endpoint until it returns 200.
func WaitForServerReady(baseURL string) error {
	var client http.Client

	u, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	readyz := fmt.Sprintf("%s/readyz", u)

	req, err := http.NewRequest(http.MethodGet, readyz, http.NoBody)
	if err != nil {
		return err
	}

	const backoff = 200 * time.Millisecond
	const maxWait = 5 * time.Second
	var (
		resp *http.Response
		rerr error
	)
	for i := 0; i < int(maxWait/backoff); i++ {
		resp, rerr = client.Do(req.Clone(context.Background()))

		if resp != nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close() //noline:errcheck // we don't care about the body
			return nil
		}

		time.Sleep(backoff)
	}
	if rerr != nil {
		return fmt.Errorf("%s not ready: %w", u.Hostname(), rerr)
	}

	return fmt.Errorf("%s not ready", u.Hostname())
}
