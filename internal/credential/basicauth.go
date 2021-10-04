// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package credential

import (
	"encoding/base64"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/saucelabs/customerror"
)

var (
	ErrMissingCredential        = customerror.NewMissingError("credential", "", nil)
	ErrUsernamePasswordRequired = customerror.NewRequiredError("username, and password are", "", nil)
)

// BasicAuth is the basic authentication credential definition.
//
// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication#basic_authentication_scheme
type BasicAuth struct {
	Username string `json:"username" validate:"required,gte=3"`
	Password string `json:"password" validate:"required,gte=3"`
}

// ToBase64 converts a basic auth credential to the base64 format.
func (bC *BasicAuth) ToBase64() string {
	return base64.StdEncoding.EncodeToString([]byte(bC.Username + ":" + bC.Password))
}

//////
// Factory.
//////

// NewBasicAuthFromText is a BasicAuth factory that automatically parses
// `credential` from text.
func NewBasicAuthFromText(credential string) (*BasicAuth, error) {
	if credential == "" {
		return nil, ErrMissingCredential
	}

	cred := strings.Split(credential, ":")

	if len(cred) != 2 {
		return nil, ErrUsernamePasswordRequired
	}

	return NewBasicAuth(cred[0], cred[1])
}

// NewBasicAuth is the BasicAuth factory.
func NewBasicAuth(username, password string) (*BasicAuth, error) {
	bC := &BasicAuth{
		Username: username,
		Password: password,
	}

	if err := validator.New().Struct(bC); err != nil {
		return nil, customerror.NewInvalidError("credential", "", err)
	}

	return bC, nil
}
