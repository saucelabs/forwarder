package middleware

import (
	"net/http"
	"slices"
)

func BehaviorUseNTLMInsteadOfNegotiateModifier(resp *http.Response) error {
	// if response header contain WWW-Authenticate Negotiate and WWW-Authenticate NTLM,
	// remove Negotiate to force NTLM

	wwwAuthenticateValues := resp.Header.Values("WWW-Authenticate")

	if len(wwwAuthenticateValues) < 2 {
		return nil
	}

	if slices.Contains(wwwAuthenticateValues, "Negotiate") && slices.Contains(wwwAuthenticateValues, "NTLM") {
		negotiatePosition := slices.Index(wwwAuthenticateValues, "Negotiate")

		if negotiatePosition == -1 { // should not happen ever at this point
			return nil
		}

		updatedHeaderValues := slices.Delete(wwwAuthenticateValues, negotiatePosition, negotiatePosition+1)
		resp.Header.Del("WWW-Authenticate")

		for _, value := range updatedHeaderValues {
			resp.Header.Add("WWW-Authenticate", value)
		}
	}
	return nil
}
