package auth

import (
	"errors"
	"fmt"
)

const (
	errParsingTokenIssuerResponseText = "unable to parse response from %s: %w"
)

var (
	ErrUnexpectedStatusCode  = errors.New("unexpected status code while getting token")
	ErrNoAccessTokenReturned = errors.New("no access_token returned")
	ErrEnvironmentAuth       = errors.New("there was an error authorizing for the requested environment")
	ErrUserAuthPolling       = errors.New("there was an error polling for user auth")
)

func newUnexpectedStatusCodeError(statusCode int) error {
	return fmt.Errorf("%w %d", ErrUnexpectedStatusCode, statusCode)
}

func newNoAccessTokenReturnedError(url string) error {
	return fmt.Errorf("%w from %s", ErrNoAccessTokenReturned, url)
}

func newEnvironmentAuthError(err string, description string) error {
	return fmt.Errorf("%w. Err: %s, Desc: %s", ErrEnvironmentAuth, err, description)
}

func newUserAuthPollingError(err string, description string) error {
	return fmt.Errorf("%w. Err: %s, Desc: %s", ErrUserAuthPolling, err, description)
}

func newErrorParsingTokenIssuerResponse(url string, err error) error {
	return fmt.Errorf(errParsingTokenIssuerResponseText, url, err)
}
