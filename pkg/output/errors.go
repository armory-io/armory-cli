package output

import "fmt"

const (
	errJsonMarshalText = "failed to marshal response to json"
	errYamlMarshalText = "failed to marshal response to yaml"
	errHttpRequestText = "request returned an error: status code(%d) %w"
)

func newJsonMarshalError(err error) error {
	return fmt.Errorf("%s: %w", errJsonMarshalText, err)
}

func newYamlMarshalError(err error) error {
	return fmt.Errorf("%s: %w", errYamlMarshalText, err)
}

func newHttpRequestError(statusCode int, err error) error {
	return fmt.Errorf(errHttpRequestText, statusCode, err)
}
