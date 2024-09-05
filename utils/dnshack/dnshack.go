// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found at https://github.com/golang/go/blob/-/LICENSE.

//go:build dnshack

package dnshack

import (
	"errors"
	"fmt"
	"net/netip"
	"sync"
	"time"
)

var preConfigure sync.Once

// Configure changes the Go standard library DNS resolver to use the specified
// servers with the specified timeout. If roundRobin is true, the resolver will
// rotate the order of the servers on each request.
//
// Since Go 1.23 it requires the -checklinkname=0 linker flag to work.
func Configure(servers []netip.AddrPort, timeout time.Duration, roundRobin bool) error {
	preConfigure.Do(func() {
		getSystemDNSConfig()
	})

	resolvConf.acquireSema()
	defer resolvConf.releaseSema()

	procDNSCfg := resolvConf.dnsConfig.Load()
	if procDNSCfg == nil {
		return errors.New("failed to get system DNS config")
	}
	if procDNSCfg.err != nil {
		return fmt.Errorf("failed to get system DNS config: %w", procDNSCfg.err)
	}

	procDNSCfg.servers = make([]string, len(servers))
	for i := range servers {
		procDNSCfg.servers[i] = servers[i].String()
	}
	procDNSCfg.timeout = timeout
	procDNSCfg.rotate = roundRobin

	// Disable config reload from system dns config file (/etc/resolv.conf).
	procDNSCfg.noReload = true

	resolvConf.dnsConfig.Store(procDNSCfg)

	return nil
}
