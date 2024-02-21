// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/google/go-github/v56/github"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

type asset struct {
	OS   string `yaml:"os"`
	Arch string `yaml:"arch"`
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

func assetOS(name string) string {
	switch {
	case strings.HasSuffix(name, ".deb"):
		return "Debian/Ubuntu"
	case strings.HasSuffix(name, ".rpm"):
		return "RedHat/CentOS/Fedora"

	case strings.Contains(name, "linux"):
		return "Linux"
	case strings.Contains(name, "windows"):
		return "Windows"
	case strings.Contains(name, "darwin"):
		return "macOS"
	default:
		return ""
	}
}

func assetArch(name string) string {
	switch regexp.MustCompile("all|amd64|arm64|aarch64|x86_64").FindString(name) {
	case "all":
		return "all"
	case "amd64", "x86_64":
		return "x86-64"
	case "arm64", "aarch64":
		return "ARM64"
	}
	return ""
}

func writeDataAssets(r *github.RepositoryRelease) error {
	f, err := os.Create(path.Join(dataDir, "assets.yml"))
	if err != nil {
		return err
	}
	bw := bufio.NewWriter(f)

	assets := make([]asset, 0, len(r.Assets))
	for _, a := range r.Assets {
		v := asset{
			OS:   assetOS(a.GetName()),
			Arch: assetArch(a.GetName()),
			Name: a.GetName(),
			URL:  a.GetBrowserDownloadURL(),
		}
		if v.OS == "" || v.Arch == "" {
			continue
		}

		if v.OS == "macOS" && !strings.Contains(v.Name, "signed") {
			continue
		}

		assets = append(assets, v)
	}

	osOrder := map[string]int{
		"macOS":                0,
		"Debian/Ubuntu":        1,
		"RedHat/CentOS/Fedora": 2,
		"Linux":                3,
		"Windows":              4,
	}
	slices.SortFunc(assets, func(a, b asset) int {
		if c := osOrder[a.OS] - osOrder[b.OS]; c != 0 {
			return c
		}
		return strings.Compare(a.Arch, b.Arch)
	})

	if err := yaml.NewEncoder(bw).Encode(assets); err != nil {
		return err
	}
	if err := bw.Flush(); err != nil {
		return err
	}

	return f.Close()
}

func writeDataLatest(r *github.RepositoryRelease) error {
	version := strings.TrimPrefix(r.GetTagName(), "v")

	f, err := os.Create(path.Join(dataDir, "latest.yml"))
	if err != nil {
		return err
	}

	for _, a := range r.Assets {
		name := a.GetName()

		if name == "checksums" {
			fmt.Fprintf(f, "checksums: %s\n", a.GetBrowserDownloadURL())
		} else if strings.HasPrefix(name, repo) {
			fmt.Fprintf(f, "%s: %s\n", name[len(repo)+1+len(version)+1:], a.GetBrowserDownloadURL())
		}
	}

	return f.Close()
}
