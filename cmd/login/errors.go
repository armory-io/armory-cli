package login

import (
	"errors"
)

var (
	ErrGettingDeviceCode      = errors.New("error at getting device code: %w")
	ErrPollingServerResponse  = errors.New("error at polling auth server for response. Err: %w")
	ErrDecodingJwt            = errors.New("error at decoding jwt. Err: %w")
	ErrGettingHomeDirectory   = errors.New("there was an error getting the home directory. Err: %w")
	ErrWritingCredentialsFile = errors.New("there was an error writing the credentials file. Err: %w")
)
