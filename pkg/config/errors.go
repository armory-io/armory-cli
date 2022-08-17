package config

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidArmoryCloudAddr = errors.New("failed to parse supplied Armory Cloud address")
)

const (
	errInvalidUrlSchemeText       = "provided addr: '%s', expected url to contain scheme http or https"
	errMissingHostInURLText       = "provided addr: '%s', expected url to contain a host"
	errIncludedPathInURLText      = "provided addr: '%s', expected url to not contain a path"
	errArmoryCloudAddrParsingText = "failed to parse supplied Armory Cloud address, provided addr: '%s', err: %w"
)

func newInvalidUrlSchemeError(armoryCloudAddr string) error {
	errorText := fmt.Sprintf(errInvalidUrlSchemeText, armoryCloudAddr)
	return fmt.Errorf("%w, %s", ErrInvalidArmoryCloudAddr, errorText)
}

func newMissingHostInUrlError(armoryCloudAddr string) error {
	errorText := fmt.Sprintf(errMissingHostInURLText, armoryCloudAddr)
	return fmt.Errorf("%w, %s", ErrInvalidArmoryCloudAddr, errorText)
}

func newIncludedPathInUrlError(armoryCloudAddr string) error {
	errorText := fmt.Sprintf(errIncludedPathInURLText, armoryCloudAddr)
	return fmt.Errorf("%w, %s", ErrInvalidArmoryCloudAddr, errorText)
}

func newArmoryCloudAddrParsingError(armoryCloudAddr string, err error) error {
	return fmt.Errorf(errArmoryCloudAddrParsingText, armoryCloudAddr, err)
}
