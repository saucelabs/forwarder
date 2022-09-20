// Copyright 2021 The forwarder Authors. All rights reserved.
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

// Validator returns new validator.Validate instance with all custom validations registered.
func Validator() *validator.Validate {
	v := validator.New()
	RegisterAll(v)
	return v
}

// RegisterAll adds registers all custom validations with the provider validator.
func RegisterAll(v *validator.Validate) {
	mustRegisterValidation(v, "basicAuth", IsBasicAuth)
	mustRegisterValidation(v, "dnsURI", IsDNSURI)
	mustRegisterValidation(v, "proxyURI", IsProxyURI)
}

func mustRegisterValidation(v *validator.Validate, tag string, fn validator.Func) {
	if err := v.RegisterValidation(tag, fn); err != nil {
		panic(err)
	}
}

// IsBasicAuth checks that the given credentials are valid:
// - Need to be in the format (username:password).
// - Need to have a username, min. length is 3.
// - Need to have password, min. length is 3.
func IsBasicAuth(fl validator.FieldLevel) bool {
	v := fl.Field().String()

	user, passwd, ok := strings.Cut(v, ":")
	if !ok {
		return false
	}
	return len(user) >= 3 && len(passwd) >= 3
}

var validDNSProtocolsRegexp = regexp.MustCompile(`(?mi)udp|tcp`)

// IsDNSURI checks if the given URI is a valid DNS URI:
// - Known protocol: udp, tcp.
// - Some hostname (x.io - min 4 chars), or IP.
// - Port in a valid range: 1 - 65535.
func IsDNSURI(fl validator.FieldLevel) bool {
	v := fl.Field().String()

	// Need to be a valid URI.
	u, err := url.Parse(v)
	if err != nil {
		return false
	}

	return u.Scheme != "" && u.Hostname() != "" && u.Port() != "" &&
		validDNSProtocolsRegexp.MatchString(u.Scheme) &&
		len(u.Hostname()) >= 4 &&
		isPort(u.Port())
}

var validProxySchemesRegexp = regexp.MustCompile(`(?mi)http|https|socks5|socks|quic`)

// IsProxyURI checks if a given URI is a valid proxy URI:
// - Known protocol: http, https, socks5, socks, quic.
// - Some hostname (x.io - min 4 chars), or IP.
// - Port in a valid range: 1 - 65535.
func IsProxyURI(fl validator.FieldLevel) bool {
	v := fl.Field().String()

	// Need to be a valid URI.
	u, err := url.Parse(v)
	if err != nil {
		return false
	}

	return u.Scheme != "" && u.Hostname() != "" && u.Port() != "" &&
		validProxySchemesRegexp.MatchString(u.Scheme) &&
		len(u.Hostname()) >= 4 &&
		isPort(u.Port())
}

// isPort returns true iff port string is a valid port number.
func isPort(port string) bool {
	p, err := strconv.Atoi(port)
	if err != nil {
		return false
	}

	return p >= 1 && p <= 65535
}
