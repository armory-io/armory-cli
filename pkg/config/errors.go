package config

import (
	"errors"
)

var (
	ErrInvalidArmoryCloudAddr = errors.New("failed to parse supplied Armory Cloud address")
	ErrInvalidUrlScheme       = errors.New("expected url to contain scheme http or https")
	ErrMissingHostInUrl       = errors.New("expected url to contain a host")
	ErrIncludedPathInUrl      = errors.New("expected url to not contain a path")
)
