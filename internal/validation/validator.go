// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package validation

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

var v *validator.Validate

// Validates if a given URL is a valid proxy url. By valid:
// - Known scheme: http, https, socks, socks5, or quic
// - Some hostname: min 4 chars (x.io)
// - Port in a valid range: 80 - 65535.
func proxyURLValidator(fl validator.FieldLevel) bool {
	fieldValue := fl.Field().String()

	parsedURL, err := url.Parse(fieldValue)
	if err != nil {
		return false
	}

	// Check for empty values.
	if parsedURL.Scheme == "" ||
		parsedURL.Hostname() == "" ||
		parsedURL.Port() == "" {
		return false
	}

	scheme := parsedURL.Scheme
	hostname := parsedURL.Hostname()
	portAsString := parsedURL.Port()

	// Validate scheme against common proxy schemes.
	validSchemes := []string{"http", "https", "socks", "socks5", "quic"}

	if !strings.Contains(strings.Join(validSchemes, ","), scheme) {
		return false
	}

	// Validate hostname.
	if len(hostname) < 4 {
		return false
	}

	// Validate port.
	port, err := strconv.Atoi(portAsString)
	if err != nil {
		return false
	}

	if port < 80 || port > 65535 {
		return false
	}

	return true
}

// Get returns validator.
func Get() *validator.Validate {
	if v == nil {
		v = Setup()
	}

	return v
}

// Setup validator.
func Setup() *validator.Validate {
	v = validator.New()

	v.RegisterValidation("proxyURL", proxyURLValidator)

	return v
}
