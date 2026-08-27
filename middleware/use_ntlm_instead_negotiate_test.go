// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package middleware

import (
	"net/http"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUseNTLMInsteadNegotiate(t *testing.T) {
	resp := http.Response{Header: http.Header{}}
	resp.Header.Add("WWW-Authenticate", "Negotiate")
	resp.Header.Add("WWW-Authenticate", "NTLM")

	err := BehaviorUseNTLMInsteadOfNegotiateModifier(&resp)
	require.NoError(t, err)

	require.Len(t, resp.Header.Values("WWW-Authenticate"), 1)
	require.Equal(t, "NTLM", resp.Header.Get("WWW-Authenticate"))
}

func TestUseNTLMInsteadNegotiateReversedOrder(t *testing.T) {
	resp := http.Response{Header: http.Header{}}
	resp.Header.Add("WWW-Authenticate", "NTLM")
	resp.Header.Add("WWW-Authenticate", "Negotiate")

	err := BehaviorUseNTLMInsteadOfNegotiateModifier(&resp)
	require.NoError(t, err)

	require.Len(t, resp.Header.Values("WWW-Authenticate"), 1)
	require.Equal(t, "NTLM", resp.Header.Get("WWW-Authenticate"))
}

func TestUseNTLMInsteadNegotiateMultipleHeaders(t *testing.T) {
	resp := http.Response{Header: http.Header{}}
	resp.Header.Add("WWW-Authenticate", "Fake1")
	resp.Header.Add("WWW-Authenticate", "NTLM")
	resp.Header.Add("WWW-Authenticate", "Fake2")
	resp.Header.Add("WWW-Authenticate", "Negotiate")
	resp.Header.Add("WWW-Authenticate", "Fake3")

	err := BehaviorUseNTLMInsteadOfNegotiateModifier(&resp)
	require.NoError(t, err)

	require.Len(t, resp.Header.Values("WWW-Authenticate"), 4)
	require.False(t, slices.Contains(resp.Header.Values("WWW-Authenticate"), "Negotiate"))
}
