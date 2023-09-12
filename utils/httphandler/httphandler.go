// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package httphandler

import (
	"encoding/json"
	"net/http"
	"runtime"
)

func SendFile(contentType string, content []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Write(content)
	})
}

func SendFileString(contentType, content string) http.Handler {
	return SendFile(contentType, []byte(content))
}

func Version(version, time, commit string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		v := struct {
			Version string `json:"version"`
			Time    string `json:"time"`
			Commit  string `json:"commit"`

			GoArch    string `json:"go_arch"`
			GOOS      string `json:"go_os"`
			GoVersion string `json:"go_version"`
		}{
			Version: version,
			Time:    time,
			Commit:  commit,

			GoArch:    runtime.GOARCH,
			GOOS:      runtime.GOOS,
			GoVersion: runtime.Version(),
		}
		json.NewEncoder(w).Encode(v) //nolint // ignore error
	})
}
