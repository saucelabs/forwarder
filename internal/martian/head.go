// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// Copyright 2015 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package martian

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/exp/maps"
)

// writeHeadResponse writes the status line and header of r to w.
func writeHeadResponse(w io.Writer, res *http.Response) error {
	// Status line
	text := res.Status
	if text == "" {
		text = http.StatusText(res.StatusCode)
		if text == "" {
			text = "status code " + strconv.Itoa(res.StatusCode)
		}
	} else {
		// Just to reduce stutter, if user set res.Status to "200 OK" and StatusCode to 200.
		// Not important.
		text = strings.TrimPrefix(text, strconv.Itoa(res.StatusCode)+" ")
	}

	if _, err := fmt.Fprintf(w, "HTTP/%d.%d %03d %s\r\n", res.ProtoMajor, res.ProtoMinor, res.StatusCode, text); err != nil {
		return err
	}

	// Header
	if err := res.Header.Write(w); err != nil {
		return err
	}

	// Add Trailer header if needed
	if len(res.Trailer) > 0 {
		if _, err := io.WriteString(w, "Trailer: "); err != nil {
			return err
		}

		for i, k := range maps.Keys(res.Trailer) {
			if i > 0 {
				if _, err := io.WriteString(w, ", "); err != nil {
					return err
				}
			}
			if _, err := io.WriteString(w, k); err != nil {
				return err
			}
		}
	}

	// End-of-header
	if _, err := io.WriteString(w, "\r\n"); err != nil {
		return err
	}

	return nil
}
