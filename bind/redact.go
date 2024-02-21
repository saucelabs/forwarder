// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package bind

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/saucelabs/forwarder/header"
)

func RedactURL(u *url.URL) string {
	return u.Redacted()
}

func RedactUserinfo(ui *url.Userinfo) string {
	if ui == nil {
		return ""
	}
	if _, has := ui.Password(); has {
		return ui.Username() + ":xxxxx"
	}
	return ui.Username()
}

func RedactHeader(h header.Header) string {
	return fmt.Sprintf("%q", h.String())
}

func RedactBase64(s string) string {
	if strings.HasPrefix(s, "data:") {
		return "data:xxxxx"
	}

	return s
}
