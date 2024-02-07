// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found at https://github.com/golang/go/blob/-/LICENSE.

package osdns

import (
	"errors"
	"fmt"
)

func Configure(cfg *Config) error {
	// Initialize the resolverConfig.
	getSystemDNSConfig()

	return configure(cfg)
}

func configure(cfg *Config) error {
	resolvConf.acquireSema()
	defer resolvConf.releaseSema()

	procDNSCfg := resolvConf.dnsConfig.Load()
	if procDNSCfg == nil {
		return errors.New("failed to get system DNS config")
	}
	if procDNSCfg.err != nil {
		return fmt.Errorf("failed to get system DNS config: %w", procDNSCfg.err)
	}

	procDNSCfg.servers = make([]string, len(cfg.Servers))
	for i := range cfg.Servers {
		procDNSCfg.servers[i] = cfg.Servers[i].String()
	}
	procDNSCfg.timeout = cfg.Timeout
	procDNSCfg.rotate = cfg.RoundRobin

	// Disable config reload from system dns config file (/etc/resolv.conf).
	procDNSCfg.noReload = true

	resolvConf.dnsConfig.Store(procDNSCfg)

	return nil
}
