// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ready

import (
	"github.com/spf13/pflag"
)

func bindAPIAddr(fs *pflag.FlagSet, addr *string) {
	fs.StringVarP(addr,
		"api-address", "", *addr, "<host:port>"+
			"The API server address. ")
}
