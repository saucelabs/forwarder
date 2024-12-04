// Copyright 2015 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package header

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/saucelabs/forwarder/internal/martian"
)

const (
	h20Prefix = "2.0"
	h11Prefix = "1.1"
	h10Prefix = "1.0"
)

// ViaModifier is a header modifier that checks for proxy redirect loops.
type ViaModifier struct {
	tag string
}

// NewViaModifier returns a new Via modifier.
func NewViaModifier(requestedBy string) *ViaModifier {
	return NewViaModifierWithBoundary(requestedBy, randomBoundary())
}

// randomBoundary generates a 10 character string to ensure that Martians that
// are chained together with the same requestedBy value do not collide.  This func
// panics if io.Readfull fails.
func randomBoundary() string {
	var buf [10]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf[:])
}

func NewViaModifierWithBoundary(requestedBy, boundary string) *ViaModifier {
	return &ViaModifier{
		tag: requestedBy + "-" + boundary,
	}
}

// ModifyRequest sets the Via header and provides loop-detection. If Via is
// already present, it will be appended to the existing value. If a loop is
// detected an error is added to the context and the request round trip is
// skipped.
//
// http://tools.ietf.org/html/draft-ietf-httpbis-p1-messaging-14#section-9.9
func (m *ViaModifier) ModifyRequest(req *http.Request) error {
	via := req.Header.Get("Via")

	var sb strings.Builder
	sb.Grow(m.nextLen(via))

	if via != "" {
		if strings.Contains(via, m.tag) {
			req.Close = true
			return martian.ErrorStatus{
				Err:    fmt.Errorf("via: detected request loop, header contains %s", via),
				Status: 400,
			}
		}

		sb.WriteString(via)
		sb.WriteString(", ")
	}

	switch req.ProtoMajor*10 + req.ProtoMinor {
	case 20:
		sb.WriteString(h20Prefix)
	case 11:
		sb.WriteString(h11Prefix)
	case 10:
		sb.WriteString(h10Prefix)
	default:
		fmt.Fprintf(&sb, "%d.%d", req.ProtoMajor, req.ProtoMinor)
	}

	sb.WriteByte(' ')
	sb.WriteString(m.tag)

	req.Header.Set("Via", sb.String())

	return nil
}

func (m *ViaModifier) nextLen(via string) int {
	l := 0

	if via != "" {
		l += len(via) + 2
	}

	l += len(h11Prefix) + 1 + len(m.tag)

	return l
}
