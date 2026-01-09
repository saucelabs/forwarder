// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build e2e

package tests

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/saucelabs/forwarder/e2e/forwarder"
	tspb "github.com/saucelabs/forwarder/internal/martian/h2/testservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

func TestGRPC(t *testing.T) {
	tlsCfg := &tls.Config{
		InsecureSkipVerify: true,
	}

	t.Setenv("HTTPS_PROXY", proxy) // set proxy for grpc.Dial
	conn, err := grpc.Dial(forwarder.GRPCTestServiceName+":1443", grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	fixture := tspb.NewTestServiceClient(conn)

	t.Run("Echo", func(t *testing.T) {
		ctx := context.Background()
		req := &tspb.EchoRequest{
			Payload: "Hello",
		}
		resp, err := fixture.Echo(ctx, req)
		if err != nil {
			t.Fatalf("fixture.Echo(...) = _, %v, want _, nil", err)
		}
		if got, want := resp.GetPayload(), req.GetPayload(); got != want {
			t.Errorf("resp.GetPayload() = %s, want = %s", got, want)
		}
	})

	t.Run("LargeEcho", func(t *testing.T) {
		// Sends a >128KB payload through the proxy. Since the standard gRPC frame size is only 16KB,
		// this exercises frame merging, splitting and flow control code.
		payload := make([]byte, 128*1024)
		rand.Read(payload)
		req := &tspb.EchoRequest{
			Payload: base64.StdEncoding.EncodeToString(payload),
		}

		// This test also covers using gzip compression. Ideally, we would test more compression types
		// but the golang gRPC implementation only provides a gzip compressor.
		tests := []struct {
			name           string
			useCompression bool
		}{
			{"RawData", false},
			{"Gzip", true},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				ctx := context.Background()
				var resp *tspb.EchoResponse
				if tc.useCompression {
					resp, err = fixture.Echo(ctx, req, grpc.UseCompressor(gzip.Name))
				} else {
					resp, err = fixture.Echo(ctx, req)
				}
				if err != nil {
					t.Fatalf("fixture.Echo(...) = _, %v, want _, nil", err)
				}
				if got, want := resp.GetPayload(), req.GetPayload(); got != want {
					t.Errorf("resp.GetPayload() = %s, want = %s", got, want)
				}
			})
		}
	})

	t.Run("Stream", func(t *testing.T) {
		ctx := context.Background()
		stream, err := fixture.DoubleEcho(ctx)
		if err != nil {
			t.Fatalf("fixture.DoubleEcho(ctx) = _, %v, want _, nil", err)
		}

		var received []*tspb.EchoResponse

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				resp, err := stream.Recv()
				if errors.Is(err, io.EOF) {
					return
				}
				if err != nil {
					t.Errorf("stream.Recv() = %v, want nil", err)
					return
				}
				received = append(received, resp)
			}
		}()

		var sent []*tspb.EchoRequest
		for range 5 {
			payload := make([]byte, 20*1024)
			rand.Read(payload)
			req := &tspb.EchoRequest{
				Payload: base64.StdEncoding.EncodeToString(payload),
			}
			if err := stream.Send(req); err != nil {
				t.Fatalf("stream.Send(req) = %v, want nil", err)
			}
			sent = append(sent, req)
		}
		if err := stream.CloseSend(); err != nil {
			t.Fatalf("stream.CloseSend() = %v, want nil", err)
		}
		wg.Wait()

		for i, req := range sent {
			want := req.GetPayload()
			if got := received[2*i].GetPayload(); got != want {
				t.Errorf("received[2*i].GetPayload() = %s, want %s", got, want)
			}
			if got := received[2*i+1].GetPayload(); got != want {
				t.Errorf("received[2*i+1].GetPayload() = %s, want %s", got, want)
			}
		}
	})
}
