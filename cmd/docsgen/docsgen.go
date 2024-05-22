// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path"

	"github.com/saucelabs/forwarder/command/forwarder"
	"github.com/saucelabs/forwarder/command/run"
	"github.com/saucelabs/forwarder/utils/docsgen"
	"github.com/spf13/cobra"
)

var (
	docsDir = flag.String("docs-dir", "", "path to the docs directory")

	cliDir, cfgDir, dataDir string
)

func main() {
	flag.Parse()

	cliDir = path.Join(*docsDir, "content", "cli")
	cfgDir = path.Join(*docsDir, "content", "config")
	dataDir = path.Join(*docsDir, "data")

	for _, dir := range []string{cliDir, cfgDir, dataDir} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			log.Fatal(err)
		}
	}

	contentDir := path.Join(*docsDir, "content")

	cg := forwarder.CommandGroups()
	cg.Add(&cobra.Command{
		Use: "forwarder",
	})
	if err := docsgen.WriteCommandIndex(cg, cliDir, "Forwarder"); err != nil {
		log.Fatal(err)
	}

	if err := docsgen.WriteCommandDoc(forwarder.Command(), cliDir); err != nil {
		log.Fatal(err)
	}

	if err := docsgen.WriteDefaultConfig(forwarder.Command(), cfgDir); err != nil {
		log.Fatal(err)
	}

	p, err := run.Metrics()
	if err != nil {
		log.Fatal(err)
	}

	if err := docsgen.WriteCommandProm("forwarder run", p, contentDir); err != nil {
		log.Fatal(err)
	}

	gh := newGitHubClient()

	r, _, err := gh.Repositories.GetLatestRelease(context.Background(), owner, repo)
	if err != nil {
		log.Fatal(err)
	}
	if len(r.Assets) == 0 {
		log.Fatalf("no assets found for release %s", r.GetTagName())
	}

	if err := writeDataAssets(r); err != nil {
		log.Fatal(err)
	}

	if err := writeDataLatest(r); err != nil {
		log.Fatal(err)
	}
}
