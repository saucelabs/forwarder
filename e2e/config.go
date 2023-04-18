// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package e2e

import (
	"net"
	"os"
	"strconv"

	"github.com/saucelabs/forwarder/e2e/testrunner"
	"gopkg.in/yaml.v3"
)

type TestConfig struct {
	Proxy          string `yaml:"proxy,omitempty"`
	ProxyAPI       string `yaml:"proxy-api,omitempty"`
	ProxyBasicAuth string `yaml:"proxy-basic-auth,omitempty"`
	HTTPBin        string `yaml:"httpbin,omitempty"`
	HTTPBinAPI     string `yaml:"httpbin-api,omitempty"`
	UpstreamAPI    string `yaml:"upstream-api,omitempty"`
	Insecure       string `yaml:"insecure-skip-verify,omitempty"`
	SC2450         string `yaml:"sc2450,omitempty"`
}

func parseTestConfig(val string) (*TestConfig, error) {
	b, err := os.ReadFile(val)
	if err != nil {
		return nil, err
	}
	var tc TestConfig
	if err := yaml.Unmarshal(b, &tc); err != nil {
		return nil, err
	}
	return &tc, nil
}

type ProxyConfig struct {
	Protocol      string `yaml:"protocol,omitempty"`
	BasicAuth     string `yaml:"basic-auth,omitempty"`
	LocalhostMode string `yaml:"proxy-localhost,omitempty"`
	Goleak        string `yaml:"goleak,omitempty"`
	Pac           string `yaml:"pac,omitempty"`
	Address       string `yaml:"address,omitempty"`
	APIAddress    string `yaml:"api-address,omitempty"`
	Upstream      string `yaml:"proxy,omitempty"`
	Credentials   string `yaml:"credentials,omitempty"`
	Insecure      string `yaml:"insecure,omitempty"`
}

type HTTPBinConfig struct {
	Protocol   string `yaml:"protocol,omitempty"`
	BasicAuth  string `yaml:"basic-auth,omitempty"`
	Address    string `yaml:"address,omitempty"`
	APIAddress string `yaml:"api-address,omitempty"`
}

type set struct {
	name     string
	test     *TestConfig
	proxy    *ProxyConfig
	upstream *ProxyConfig
	httpbin  *HTTPBinConfig
}

const (
	testBin      = "./e2e.test"
	forwarderBin = "../forwarder"
)

func (s *set) toTestRunner() testrunner.E2E {
	configs := []testrunner.Config{
		{
			Runnable: testrunner.Runnable{
				Name:    "test",
				Command: []string{testBin},
			},
			ConfigFile: s.test,
		},
		{
			Runnable: testrunner.Runnable{
				Name:    "proxy",
				Command: []string{forwarderBin, "run"},
			},
			ConfigFile: s.proxy,
		},
		{
			Runnable: testrunner.Runnable{
				Name:    "httpbin",
				Command: []string{forwarderBin, "httpbin"},
			},
			ConfigFile: s.httpbin,
		},
	}
	if s.upstream != nil {
		configs = append(configs, testrunner.Config{
			Runnable: testrunner.Runnable{
				Name:    "upstream",
				Command: []string{forwarderBin, "run"},
			},
			ConfigFile: s.upstream,
		})
	}
	return testrunner.E2E{
		Name:    s.name,
		Configs: configs,
	}
}

type setModifier func(*set)

func withName(name string) setModifier {
	return func(s *set) {
		s.name = name
	}
}

func proxyProtocol(p string) setModifier {
	return func(s *set) {
		s.proxy.Protocol = p
		if p == "h2" {
			p = "https"
		}
		s.test.Proxy = p + "://" + s.proxy.Address
	}
}

func upstreamProtocol(p string) setModifier {
	return func(s *set) {
		s.upstream.Protocol = p
		if p == "h2" {
			p = "https"
		}
		s.proxy.Upstream = p + "://" + s.upstream.Address
	}
}

func httpbinProtocol(p string) setModifier {
	return func(s *set) {
		s.httpbin.Protocol = p
		if p == "h2" {
			p = "https"
		}
		s.test.HTTPBin = p + "://" + s.httpbin.Address
	}
}

func proxyBasicAuth(ba string) setModifier {
	return func(s *set) {
		s.proxy.BasicAuth = ba
		s.test.ProxyBasicAuth = s.proxy.BasicAuth
	}
}

func upstreamBasicAuth(ba string) setModifier {
	return func(s *set) {
		s.upstream.BasicAuth = ba
		s.proxy.Credentials = ba + "@" + s.upstream.Address
	}
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	port := l.Addr().(*net.TCPAddr).Port //nolint:forcetypeassert // we know it's a TCPAddr
	if err := l.Close(); err != nil {
		return 0, err
	}
	return port, nil
}

func defaultSet() (*set, error) {
	port, err := getFreePort()
	if err != nil {
		return nil, err
	}
	proxyAddr := "localhost:" + strconv.Itoa(port)

	port, err = getFreePort()
	if err != nil {
		return nil, err
	}
	proxyAPIAddr := "localhost:" + strconv.Itoa(port)

	port, err = getFreePort()
	if err != nil {
		return nil, err
	}
	httpbinAddr := "localhost:" + strconv.Itoa(port)

	port, err = getFreePort()
	if err != nil {
		return nil, err
	}
	httpbinAPIAddr := "localhost:" + strconv.Itoa(port)

	return &set{
		test: &TestConfig{
			Proxy:      "http://" + proxyAddr,
			ProxyAPI:   proxyAPIAddr,
			HTTPBin:    "http://" + httpbinAddr,
			HTTPBinAPI: httpbinAPIAddr,
			Insecure:   "true",
		},
		proxy: &ProxyConfig{
			Protocol:      "http",
			LocalhostMode: "allow",
			Address:       proxyAddr,
			APIAddress:    proxyAPIAddr,
			Insecure:      "true",
		},
		httpbin: &HTTPBinConfig{
			Protocol:   "http",
			Address:    httpbinAddr,
			APIAddress: httpbinAPIAddr,
		},
	}, nil
}

func defaultSetWithUpstream() (*set, error) {
	s, err := defaultSet()
	if err != nil {
		return nil, err
	}

	port, err := getFreePort()
	if err != nil {
		return nil, err
	}
	upstreamAddr := "localhost:" + strconv.Itoa(port)

	port, err = getFreePort()
	if err != nil {
		return nil, err
	}
	upstreamAPIAddr := "localhost:" + strconv.Itoa(port)

	s.upstream = &ProxyConfig{
		Protocol:      "http",
		LocalhostMode: "allow",
		Address:       upstreamAddr,
		APIAddress:    upstreamAPIAddr,
		Insecure:      "true",
	}

	s.proxy.Upstream = "http://" + upstreamAddr
	return s, nil
}

func newSet(mods ...setModifier) (*set, error) {
	s, err := defaultSet()
	if err != nil {
		return nil, err
	}
	for _, m := range mods {
		m(s)
	}
	return s, nil
}

func newSetWithUpstream(mods ...setModifier) (*set, error) {
	s, err := defaultSetWithUpstream()
	if err != nil {
		return nil, err
	}
	for _, m := range mods {
		m(s)
	}
	return s, nil
}

func AllTests() ([]testrunner.E2E, error) {
	var sets []*set
	s, err := standardSets()
	if err != nil {
		return nil, err
	}

	sets = append(sets, s...)

	res := make([]testrunner.E2E, 0, len(sets))
	for _, s := range sets {
		res = append(res, s.toTestRunner())
	}
	return res, nil
}

func standardSets() ([]*set, error) {
	var res []*set //nolint:prealloc // it is no use
	for _, p := range []string{"http", "https"} {
		for _, h := range []string{"http", "https", "h2"} {
			s, err := newSet(withName("default-"+p+"-"+h), proxyProtocol(p),
				httpbinProtocol(h), proxyBasicAuth("u1:p1"))
			if err != nil {
				return nil, err
			}
			res = append(res, s)
		}
	}
	for _, p := range []string{"http", "https"} {
		for _, u := range []string{"http", "https"} {
			for _, h := range []string{"http", "https", "h2"} {
				s, err := newSetWithUpstream(withName("default-"+p+"-"+u+"-"+h), proxyProtocol(p),
					upstreamProtocol(u), httpbinProtocol(h), proxyBasicAuth("u1:p1"))
				if err != nil {
					return nil, err
				}
				res = append(res, s)
			}
		}
	}

	for _, h := range []string{"http", "https", "h2"} {
		s, err := newSetWithUpstream(withName("upstream-auth-"+h), httpbinProtocol(h), upstreamBasicAuth("u2:p2"))
		if err != nil {
			return nil, err
		}
		res = append(res, s)
	}

	return res, nil
}

// TODO: add pac, sc2450
