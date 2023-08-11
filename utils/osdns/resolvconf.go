// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found at https://github.com/golang/go/blob/-/LICENSE.

package osdns

import (
	_ "net" // for go:linkname
	"sync"
	"sync/atomic"
	"time"
	_ "unsafe" // for go:linkname
)

// A resolverConfig is a 1:1 copy of net.resolverConfig struct.
// Source: https://github.com/golang/go/blob/-/src/net/dnsclient_unix.go
// Do not modify.
//
//nolint:unused // used by go:linkname
type resolverConfig struct {
	initOnce sync.Once // guards init of resolverConfig

	// ch is used as a semaphore that only allows one lookup at a
	// time to recheck resolv.conf.
	ch          chan struct{} // guards lastChecked and modTime
	lastChecked time.Time     // last time resolv.conf was checked

	dnsConfig atomic.Pointer[dnsConfig] // parsed resolv.conf structure used in lookups
}

func (conf *resolverConfig) acquireSema() {
	conf.ch <- struct{}{}
}

func (conf *resolverConfig) releaseSema() {
	<-conf.ch
}

//go:linkname resolvConf net.resolvConf
var resolvConf resolverConfig //nolint:gochecknoglobals // used by go:linkname

//go:linkname getSystemDNSConfig net.getSystemDNSConfig
func getSystemDNSConfig() *dnsConfig
