// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package compose

import (
	"os"
	"path/filepath"
)

func writeFile(name string, content []byte) error {
	if err := createDir(filepath.Dir(name)); err != nil {
		return err
	}
	return os.WriteFile(name, content, 0o600)
}

func createDir(name string) error {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return os.MkdirAll(name, 0o755)
	}
	return err
}
