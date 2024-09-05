// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hostsfile

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReadLocalhostAliases(t *testing.T) {
	data := `
127.0.0.1	localhost
255.255.255.255	broadcasthost
::1             localhost
127.0.0.1 kubernetes.docker.internal
# End of section

127.0.0.1	SL-666

192.168.0.60	fedora
`
	l, err := readLocalhostAliases(strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	golden := []string{
		"SL-666",
		"kubernetes.docker.internal",
		"localhost",
	}

	if diff := cmp.Diff(l, golden); diff != "" {
		t.Errorf("unexpected result (-want +got):\n%s", diff)
	}
}
