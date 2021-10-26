package proxy

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/saucelabs/customerror"
	"github.com/saucelabs/forwarder/internal/validation"
)

var ErrFailedToCopyOptions = customerror.NewFailedToError("deepCopy options", "", nil)

// Loads, validate credential from env var, and set URI's user.
func loadCredentialFromEnvVar(envVar string, uri *url.URL) error {
	credentialFromEnvVar := os.Getenv(envVar)

	if credentialFromEnvVar != "" {
		if err := validation.Get().Var(credentialFromEnvVar, "basicAuth"); err != nil {
			errMsg := fmt.Sprintf("env var (%s)", envVar)

			return customerror.NewInvalidError(errMsg, "", err)
		}

		cred := strings.Split(credentialFromEnvVar, ":")

		uri.User = url.UserPassword(cred[0], cred[1])
	}

	return nil
}

// Copy from `source` to `target`.
//
// Basic deep copy implementation.
func deepCopy(source, target interface{}) error {
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(source); err != nil {
		return customerror.Wrap(ErrFailedToCopyOptions, err)
	}

	if err := gob.NewDecoder(buf).Decode(target); err != nil {
		return customerror.Wrap(ErrFailedToCopyOptions, err)
	}

	return nil
}
