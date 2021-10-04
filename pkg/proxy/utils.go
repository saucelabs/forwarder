package proxy

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/saucelabs/customerror"
	"github.com/saucelabs/forwarder/internal/validation"
)

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
