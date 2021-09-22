// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package version

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	// Should have real values replaces @ build time.
	buildCommit  = "Commit wasn't set @ build time"
	buildTime    = "Date wasn't set @ build time"
	buildVersion = "Version wasn't set @ build time"

	// Singleton.
	version *Version
)

// Version definition.
type Version struct {
	Commit  string `json:"commit"`
	Time    string `json:"time"`
	Version string `json:"version"`
}

// String prints the version.
func (v *Version) String() string {
	buf := new(strings.Builder)

	// Prints tabular.
	fmt.Fprintln(buf, "Version:\t", buildVersion)
	fmt.Fprintln(buf, "Built time:\t", buildTime)
	fmt.Fprintln(buf, "Git commit:\t", buildCommit)
	fmt.Fprintln(buf, "Go Arch:\t", runtime.GOARCH)
	fmt.Fprintln(buf, "Go OS:\t\t", runtime.GOOS)
	fmt.Fprintln(buf, "Go Version:\t", runtime.Version())

	return buf.String()
}

// Setup version information.
func setupVersion() *Version {
	version = &Version{
		Commit:  buildCommit,
		Time:    buildTime,
		Version: buildVersion,
	}

	return version
}

// Get safely returns the version information.
func Get() *Version {
	// Do nothing, if already setup. Otherwise, can trigger race condition in
	// goroutine cases.
	if version != nil {
		return version
	}

	return setupVersion()
}
