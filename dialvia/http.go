// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dialvia

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type HTTPProxyDialer struct {
	dial      ContextDialerFunc
	proxyURL  *url.URL
	tlsConfig *tls.Config

	ConnectRequestModifier func(req *http.Request) error
}

func HTTPProxy(dial ContextDialerFunc, proxyURL *url.URL) *HTTPProxyDialer {
	if dial == nil {
		panic("dial is required")
	}
	if proxyURL == nil {
		panic("proxy URL is required")
	}
	if proxyURL.Scheme != "http" {
		panic("proxy URL scheme must be http")
	}

	return &HTTPProxyDialer{
		dial:     dial,
		proxyURL: proxyURL,
	}
}

func HTTPSProxy(dial ContextDialerFunc, proxyURL *url.URL, tlsConfig *tls.Config) *HTTPProxyDialer {
	if dial == nil {
		panic("dial is required")
	}
	if proxyURL == nil {
		panic("proxy URL is required")
	}
	if proxyURL.Scheme != "https" {
		panic("proxy URL scheme must be https")
	}
	if tlsConfig == nil {
		panic("TLS config is required")
	}

	tlsConfig.ServerName = proxyURL.Hostname()
	tlsConfig.NextProtos = []string{"http/1.1"}

	return &HTTPProxyDialer{
		dial:      dial,
		proxyURL:  proxyURL,
		tlsConfig: tlsConfig,
	}
}

func (d *HTTPProxyDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	res, conn, err := d.DialContextR(ctx, network, addr)
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode/100 != 2 {
		b, err := httputil.DumpResponse(res, true)
		if err != nil {
			b = []byte(fmt.Sprintf("error dumping response: %s", err))
		}

		conn.Close()
		return nil, fmt.Errorf("proxy connection failed status=%d\n\n%s", res.StatusCode, string(b))
	}

	return conn, nil
}

// DialContextR is like DialContext but returns the HTTP response as well.
// The caller is responsible for closing the response body.
func (d *HTTPProxyDialer) DialContextR(ctx context.Context, network, addr string) (*http.Response, net.Conn, error) {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, nil, fmt.Errorf("unsupported network: %s", network)
	}

	conn, err := d.dial(ctx, "tcp", d.proxyURL.Host)
	if err != nil {
		return nil, nil, err
	}
	if d.proxyURL.Scheme == "https" {
		conn = tls.Client(conn, d.tlsConfig)
	}

	pbw := bufio.NewWriterSize(conn, 1024)
	pbr := bufio.NewReaderSize(conn, 1024)

	req := http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Host: addr},
		Host:   addr,
		Header: http.Header{},
	}

	// Don't send the default Go HTTP client User-Agent.
	req.Header.Add("User-Agent", "")
	if u := d.proxyURL.User; u != nil {
		pass, _ := u.Password()
		auth := u.Username() + ":" + pass
		req.Header.Add("Proxy-Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	}

	if cm := d.ConnectRequestModifier; cm != nil {
		if err := cm(&req); err != nil {
			conn.Close()
			return nil, nil, err
		}
	}

	if err := req.Write(pbw); err != nil {
		conn.Close()
		return nil, nil, err
	}
	if err := pbw.Flush(); err != nil {
		conn.Close()
		return nil, nil, err
	}

	resCh := make(chan *http.Response, 1)
	errCh := make(chan error, 1)

	go func() {
		res, err := http.ReadResponse(pbr, &req) //nolint:bodyclose // caller is responsible for closing the response body
		if err != nil {
			errCh <- err
		} else {
			resCh <- res
		}
	}()

	select {
	case <-ctx.Done():
		conn.Close()
		return nil, nil, ctx.Err()
	case err := <-errCh:
		conn.Close()
		return nil, nil, err
	case res := <-resCh:
		return res, conn, nil
	}
}
