package auth

import (
	"errors"
	"fmt"
)

const errorParsingTokenIssuerResponseText = "unable to parse response from %s: %w"

var ErrUnexpectedStatusCode = errors.New("unexpected status code while getting token")
var ErrNoAccessTokenReturned = errors.New("no access_token returned")
var EnvironmentAuthError = errors.New("there was an error authorizing for the requested environment")
var UserAuthPollingError = errors.New("there was an error polling for user auth")

func newUnexpectedStatusCodeError(statusCode int) error {
	return fmt.Errorf("%w %d", ErrUnexpectedStatusCode, statusCode)
}

func newNoAccessTokenReturnedError(url string) error {
	return fmt.Errorf("%w from %s", ErrNoAccessTokenReturned, url)
}

func newEnvironmentAuthError(err string, description string) error {
	return fmt.Errorf("%w. Err: %s, Desc: %s", EnvironmentAuthError, err, description)
}

func newUserAuthPollingError(err string, description string) error {
	return fmt.Errorf("%w. Err: %s, Desc: %s", UserAuthPollingError, err, description)
}

func newErrorParsingTokenIssuerResponse(url string, err error) error {
	return fmt.Errorf(errorParsingTokenIssuerResponseText, url, err)
}
