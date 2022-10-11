package forwarder

import (
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
)

const (
	AuthorizationHeader      = "Authorization"
	ProxyAuthorizationHeader = "Proxy-Authorization"
)

// BasicAuthUtil exposes common Basic Authentication functionalities from the standard library,
// and allows to customize the Authentication header.
// This is useful when you want to use Basic Authentication for a proxy.
//
// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Proxy-Authorization
// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Authorization
type BasicAuthUtil struct {
	Header string
}

// AuthenticatedRequest parses the provided HTTP request for Basic Authentication credentials
// and returns true if the provided credentials match the expected username and password.
// Returns false if the request is unauthenticated.
// Uses constant-time comparison in order to mitigate timing attacks.
func (ba *BasicAuthUtil) AuthenticatedRequest(r *http.Request, expectedUser, expectedPass string) bool {
	user, pass, ok := ba.BasicAuth(r)
	if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(expectedUser)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(expectedPass)) != 1 {
		return false
	}

	return true
}

// BasicAuth returns the username and password provided in the request's authorization header,
// if the request uses HTTP Basic Authentication.
// See RFC 2617, Section 2.
func (ba *BasicAuthUtil) BasicAuth(r *http.Request) (username, password string, ok bool) {
	auth := r.Header.Get(ba.Header)
	if auth == "" {
		return "", "", false
	}
	return parseBasicAuth(auth)
}

// parseBasicAuth parses an HTTP Basic Authentication string.
// "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==" returns ("Aladdin", "open sesame", true).
func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	// Case insensitive prefix match. See Issue 22736.
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return "", "", false
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return "", "", false
	}
	cs := string(c)
	username, password, ok = strings.Cut(cs, ":")
	if !ok {
		return "", "", false
	}
	return username, password, true
}

// SetBasicAuth sets the request's authorization header to use HTTP
// Basic Authentication with the provided username and password.
//
// With HTTP Basic Authentication the provided username and password
// are not encrypted. It should generally only be used in an HTTPS
// request.
//
// The username may not contain a colon. Some protocols may impose
// additional requirements on pre-escaping the username and
// password. For instance, when used with OAuth2, both arguments must
// be URL encoded first with url.QueryEscape.
func (ba *BasicAuthUtil) SetBasicAuth(r *http.Request, username, password string) {
	r.Header.Set(ba.Header, "Basic "+basicAuth(username, password))
}

// See 2 (end of page 4) https://www.ietf.org/rfc/rfc2617.txt
// "To receive authorization, the client sends the userid and password, separated by a single colon (":") character,
// within a base64 encoded string in the credentials."
// It is not meant to be urlencoded.
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// Wrap wraps the provided http.Handler with basic authentication.
// If header is Proxy-Authorization and the request is not authenticated, the handler is not called and a 407 Proxy Authentication Required is returned.
// Otherwise, if the request is not authenticated, the handler is not called and a 401 Unauthorized is returned.
// The provided username and password are used to authenticate the request.
func (ba *BasicAuthUtil) Wrap(h http.Handler, expectedUser, expectedPass string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !ba.AuthenticatedRequest(r, expectedUser, expectedPass) {
			if ba.Header == ProxyAuthorizationHeader {
				w.Header().Set("Proxy-Authenticate", "Basic realm=\"SauceLabs Forwarder\"")
				w.Header().Set("Proxy-Connection", "close")
				w.WriteHeader(http.StatusProxyAuthRequired)
			} else {
				w.Header().Set("WWW-Authenticate", "Basic realm=\"SauceLabs Forwarder\"")
				w.Header().Set("Connection", "close")
				w.WriteHeader(http.StatusUnauthorized)
			}
			return
		}

		h.ServeHTTP(w, r)
	})
}
