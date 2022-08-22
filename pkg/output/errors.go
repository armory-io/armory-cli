package output

import (
	"errors"
)

var (
	ErrJsonMarshal = errors.New("failed to marshal response to json")
	ErrYamlMarshal = errors.New("failed to marshal response to yaml")
	ErrHttpRequest = errors.New("request returned an error: status code(%d) %w")
)
