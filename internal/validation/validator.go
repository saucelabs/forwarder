// Copyright 2021 The pacman Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package validation

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	v *validator.Validate

	validProxySchemesRegex = regexp.MustCompile(`(?mi)http|https|socks5|socks|quic`)
	validTextOrURIRegex    = regexp.MustCompile(`(?mi)http|https|file|.pac|FindProxyForURL`)
)

// Checks if the given credential is valid. By valid:
// - Need to be in the format (username:password)
// - Need to have a username, min. length is 3
// - Need to have password, min. length is 3
// - Total min. length is 7 chars.
//
// TODO: Better name it.
func basicAuthCredentialValidator(fl validator.FieldLevel) bool {
	fieldValue := strings.ToLower(fl.Field().String())

	// Can't be empty.
	if fieldValue == "" {
		return false
	}

	// Min. length is 3 chars.
	if len(fieldValue) < 7 {
		return false
	}

	// Need to have `:`.
	if !strings.Contains(fieldValue, ":") {
		return false
	}

	// Need to have 2 components: username, and password.
	credential := strings.Split(fieldValue, ":")

	if len(credential) != 2 {
		return false
	}

	// Min. username length is 3 char.
	if len(credential[0]) < 3 {
		return false
	}

	// Min. password length is 3 char.
	if len(credential[1]) < 3 {
		return false
	}

	return true
}

// Checks if the given text, or URI is valid for the PAC loader. By valid:
// - Remote loading: `http`, or `https`
// - Local loading: `file`
// - Direct loading: `function` keyword
//
// TODO: Better name it.
func pacTextOrURIValidator(fl validator.FieldLevel) bool {
	fieldValue := strings.ToLower(fl.Field().String())

	// Can't be empty.
	if fieldValue == "" {
		return false
	}

	// Min. length is 3 chars.
	if len(fieldValue) < 4 {
		return false
	}

	// Validate scheme against common proxy schemes.
	if !validTextOrURIRegex.MatchString(fieldValue) {
		return false
	}

	return true
}

// Checks if a given URI is a valid proxy url. By valid:
// - Known scheme: http, https, socks, socks5, or quic
// - Some hostname: min 4 chars (x.io)
// - Port in a valid range: 80 - 65535.
//
// TODO: This should be proxyURI
func proxyURIValidator(fl validator.FieldLevel) bool {
	fieldValue := strings.ToLower(fl.Field().String())

	// Can't be empty.
	if fieldValue == "" {
		return false
	}

	// Need to be a valid URI.
	parsedURL, err := url.Parse(fieldValue)
	if err != nil {
		return false
	}

	scheme := parsedURL.Scheme
	hostname := parsedURL.Hostname()
	portAsString := parsedURL.Port()

	// URI components can't be empty.
	if scheme == "" || hostname == "" || portAsString == "" {
		return false
	}

	// Need to be a valid proxy schemes.
	if !validProxySchemesRegex.MatchString(scheme) {
		return false
	}

	// Need to have a valid hostname.
	if len(hostname) < 4 {
		return false
	}

	// Need to be in a valid port range.
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

	v.RegisterValidation("proxyURI", proxyURIValidator)
	v.RegisterValidation("pacTextOrURI", pacTextOrURIValidator)
	v.RegisterValidation("basicAuth", basicAuthCredentialValidator)

	return v
}
