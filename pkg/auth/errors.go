package auth

import (
	"errors"
)

var (
	ErrParsingTokenIssuerResponse = errors.New("unable to parse response from")
	ErrUnexpectedStatusCode       = errors.New("unexpected status code while getting token")
	ErrNoAccessTokenReturned      = errors.New("no access_token returned")
	ErrEnvironmentAuth            = errors.New("there was an error authorizing for the requested environment")
	ErrUserAuthPolling            = errors.New("there was an error polling for user auth")
)
