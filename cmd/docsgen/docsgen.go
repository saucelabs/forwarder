// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"flag"
	"log"
	"os"
	"path"

	"github.com/saucelabs/forwarder/command/forwarder"
	"github.com/spf13/cobra"
)

var (
	docsDir = flag.String("docs-dir", "", "path to the docs directory")

	cliDir, cfgDir string
)

func main() {
	flag.Parse()

	cliDir = path.Join(*docsDir, "content", "cli")
	cfgDir = path.Join(*docsDir, "content", "config")

	for _, dir := range []string{cliDir, cfgDir} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			log.Fatal(err)
		}
	}

	cg := forwarder.CommandGroups()
	cg.Add(&cobra.Command{
		Use: "forwarder",
	})
	if err := writeCommandIndex(cg); err != nil {
		log.Fatal(err)
	}

	if err := writeCommandDoc(forwarder.Command()); err != nil {
		log.Fatal(err)
	}

	if err := writeDefaultConfig(forwarder.Command()); err != nil {
		log.Fatal(err)
	}
}
