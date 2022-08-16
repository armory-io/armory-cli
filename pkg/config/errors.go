package config

import (
	"errors"
	"fmt"
)

var ErrInvalidArmoryCloudAddr = errors.New("failed to parse supplied Armory Cloud address")

const invalidUrlSchemeErrorText = "provided addr: '%s', expected url to contain scheme http or https"
const missingHostInUrlErrorText = "provided addr: '%s', expected url to contain a host"
const includedPathInUrlErrorText = "provided addr: '%s', expected url to not contain a path"
const armoryCloudAddrParsingErrorText = "failed to parse supplied Armory Cloud address, provided addr: '%s', err: %w"

func newInvalidUrlSchemeError(armoryCloudAddr string) error {
	errorText := fmt.Sprintf(invalidUrlSchemeErrorText, armoryCloudAddr)
	return fmt.Errorf("%w, %s", ErrInvalidArmoryCloudAddr, errorText)
}

func newMissingHostInUrlError(armoryCloudAddr string) error {
	errorText := fmt.Sprintf(missingHostInUrlErrorText, armoryCloudAddr)
	return fmt.Errorf("%w, %s", ErrInvalidArmoryCloudAddr, errorText)
}

func newIncludedPathInUrlError(armoryCloudAddr string) error {
	errorText := fmt.Sprintf(includedPathInUrlErrorText, armoryCloudAddr)
	return fmt.Errorf("%w, %s", ErrInvalidArmoryCloudAddr, errorText)
}

func newArmoryCloudAddrParsingError(armoryCloudAddr string, err error) error {
	return fmt.Errorf(armoryCloudAddrParsingErrorText, armoryCloudAddr, err)
}
