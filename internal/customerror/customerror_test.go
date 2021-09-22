// Copyright 2021 The customerror Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package customerror

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

const (
	failedCreateSomethingMsg = "Failed to create something"
	code                     = "E1010"
	statusCode               = http.StatusNotFound
)

var ErrFailedToReachServer = errors.New("failed to reach servers")
var ErrFailedToReachServerDeep = fmt.Errorf("%s. %w", ErrFailedToReachServer, errors.New("servers are broken"))
var ErrFailedToReachServerDeepRev = fmt.Errorf("%s. %w", errors.New("servers are broken"), ErrFailedToReachServer)

func TestNew(t *testing.T) {
	type args struct {
		message    string
		code       string
		statusCode int
		err        error
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "should work - with message",
			args: args{message: failedCreateSomethingMsg},
			want: failedCreateSomethingMsg,
		},
		{
			name: "should work - with message, and code",
			args: args{
				message: failedCreateSomethingMsg,
				code:    code,
			},
			want: "E1010: Failed to create something",
		},
		{
			name: "should work - with message, and error",
			args: args{
				message: failedCreateSomethingMsg,
				err:     ErrFailedToReachServer,
			},
			want: "Failed to create something. Original Error: Failed to reach servers",
		},
		{
			name: "should work - with message, and deep error",
			args: args{
				message: failedCreateSomethingMsg,
				err:     ErrFailedToReachServerDeep,
			},
			want: "Failed to create something. Original Error: Failed to reach servers. Servers are broken",
		},
		{
			name: "should work - with message, and status code",
			args: args{
				message:    failedCreateSomethingMsg,
				statusCode: statusCode,
			},
			want: "Failed to create something (404 - Not Found)",
		},
		{
			name: "should work - with message, code, and error",
			args: args{
				message: failedCreateSomethingMsg,
				code:    code,
				err:     ErrFailedToReachServer,
			},
			want: "E1010: Failed to create something. Original Error: Failed to reach servers",
		},
		{
			name: "should work - with message, code, error, and deep error",
			args: args{
				message: failedCreateSomethingMsg,
				code:    code,
				err:     ErrFailedToReachServerDeep,
			},
			want: "E1010: Failed to create something. Original Error: Failed to reach servers. Servers are broken",
		},
		{
			name: "should work - with message, code, error, deep error, and status code",
			args: args{
				message:    failedCreateSomethingMsg,
				code:       code,
				statusCode: statusCode,
				err:        ErrFailedToReachServerDeep,
			},
			want: "E1010: Failed to create something (404 - Not Found). Original Error: Failed to reach servers. Servers are broken",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.message, tt.args.code, tt.args.statusCode, tt.args.err)

			if !strings.EqualFold(got.Error(), tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuiltin(t *testing.T) {
	var ErrFailedToCreateFile = NewFailedToError("create file", "", nil)
	var ErrInvalidPath = NewInvalidError("path", "", nil)
	var ErrMissingPath = NewMissingError("path", "", nil)
	var ErrRequiredPath = NewRequiredError("path is", "", nil)

	testFunc := func(e error) error { return e }

	type args struct {
		err error
	}
	tests := []struct {
		name   string
		args   args
		want   string
		wantAs string
	}{
		{
			name: "Should work - ErrFailedToCreateFile",
			args: args{
				err: ErrFailedToCreateFile,
			},
			want:   "failed to create file (500 - Internal Server Error)",
			wantAs: "failed to create file",
		},
		{
			name: "Should work - ErrInvalidPath",
			args: args{
				err: ErrInvalidPath,
			},
			want:   "invalid path (400 - Bad Request)",
			wantAs: "invalid path",
		},
		{
			name: "Should work - ErrMissingPath",
			args: args{
				err: ErrMissingPath,
			},
			want:   "missing path (400 - Bad Request)",
			wantAs: "missing path",
		},
		{
			name: "Should work - ErrRequiredPath",
			args: args{
				err: ErrRequiredPath,
			},
			want:   "path is required (400 - Bad Request)",
			wantAs: "path is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testFunc(tt.args.err)

			if !errors.Is(err, tt.args.err) {
				t.Errorf("Expected error to be (is) %v, got %v", tt.args.err, err)
			}

			errWrapped := fmt.Errorf("Wrapped %w", err)
			if !errors.Is(errWrapped, tt.args.err) {
				t.Errorf("Expected error to be (is - wrapped) %v, got %v", tt.args.err, errWrapped)
			}

			if !strings.EqualFold(err.Error(), tt.want) {
				t.Errorf(`Expected message to be "%v", got "%v"`, tt.want, err)
			}

			var errAs *CustomError
			if errors.As(err, &errAs) {
				if errAs.Message != tt.wantAs {
					t.Errorf(`Expected message to be (As)"%v", got "%v"`, tt.wantAs, errAs.Message)
				}
			}
		})
	}
}

func TestCustomError_Unwrap(t *testing.T) {
	type fields struct {
		Code       string
		Err        error
		Message    string
		StatusCode int
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		want    string
	}{
		{
			name: "Should work",
			fields: fields{
				Code:       "",
				Err:        errors.New("Wrapped error"),
				Message:    "Main error",
				StatusCode: 0,
			},
			wantErr: true,
			want:    "Wrapped error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cE := &CustomError{
				Code:       tt.fields.Code,
				Err:        tt.fields.Err,
				Message:    tt.fields.Message,
				StatusCode: tt.fields.StatusCode,
			}
			err := cE.Unwrap()

			if (err != nil) != tt.wantErr {
				t.Errorf("CustomError.Unwrap() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err.Error() != tt.want {
				t.Errorf("CustomError.Unwrap() message = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestNew_deepNestedErrors(t *testing.T) {
	expectedErrMsg := "custom message. Original Error: layer 3. layer 2. layer 1"

	layer1 := errors.New("layer 1")

	layer2 := fmt.Errorf("layer 2. %w", layer1)

	layer3 := fmt.Errorf("layer 3. %w", layer2)

	ErrLayered := New("custom message", "", 0, layer3)
	if ErrLayered.Error() != expectedErrMsg {
		t.Errorf("CustomError deep nested errors got %s, want %s", ErrLayered, expectedErrMsg)
	}

	var testFunc = func() error { return ErrLayered }

	errLayered := testFunc()

	if !errors.Is(errLayered, ErrLayered) {
		t.Errorf("Expected %v be ErrLayered", errLayered)
	}

	errSome := errors.New("Some error")

	errWrapped := Wrap(errLayered, errSome)

	if !errors.Is(errWrapped, ErrLayered) {
		t.Errorf("Expected %v be ErrLayered", errWrapped)
	}

	expectedErrWrappedMsg := "custom message. Original Error: layer 3. layer 2. layer 1. Wrapped Error: Some error"

	if errWrapped.Error() != expectedErrWrappedMsg {
		t.Errorf("Expected %v to be %s", errWrapped.Error(), expectedErrWrappedMsg)
	}
}

func TestWrap(t *testing.T) {
	expectedErrMsg := "custom message. Original Error: layer 3. layer 2. layer 1"

	layer1 := errors.New("layer 1")

	layer2 := fmt.Errorf("layer 2. %w", layer1)

	layer3 := fmt.Errorf("layer 3. %w", layer2)

	ErrLayered := New("custom message", "", 0, layer3)
	if ErrLayered.Error() != expectedErrMsg {
		t.Errorf("Wrap got %s, want %s", ErrLayered, expectedErrMsg)
	}

	var testFunc = func() error { return ErrLayered }

	errLayered := testFunc()

	if !errors.Is(errLayered, ErrLayered) {
		t.Errorf("Wrap Is got %s, want %s", errLayered, ErrLayered)
	}

	errSome := errors.New("Some error")

	if !errors.Is(Wrap(errLayered, errSome), ErrLayered) {
		t.Errorf("Wrap Is got %s, want %s", errSome, ErrLayered)
	}
}

func TestCustomError_SetStatusCode(t *testing.T) {
	type fields struct {
		Code       string
		Err        error
		Message    string
		StatusCode int
	}
	type args struct {
		code int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *CustomError
	}{
		{
			name: "Should work",
			fields: fields{
				Code:       "",
				Err:        nil,
				Message:    "test",
				StatusCode: 0,
			},
			args: args{
				code: 400,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cE := New(
				tt.fields.Message,
				tt.fields.Code,
				tt.fields.StatusCode,
				tt.fields.Err,
			).SetStatusCode(tt.args.code)

			if cE.StatusCode != tt.args.code {
				t.Errorf("CustomError.SetStatusCode() = %v, want %v", cE.StatusCode, tt.args.code)
			}
		})
	}
}
