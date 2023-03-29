// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package bind

import (
	"fmt"
	"net/url"
)

func RedactURL(u *url.URL) string {
	return u.Redacted()
}

func RedactUserinfo(ui *url.Userinfo) string {
	if ui == nil {
		return ""
	}
	if _, has := ui.Password(); has {
		return fmt.Sprintf("%s:xxxxx", ui.Username())
	}
	return ui.Username()
}
