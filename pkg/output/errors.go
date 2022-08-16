package output

import "fmt"

const jsonMarshalErrorText = "failed to marshal response to json"
const yamlMarshalErrorText = "failed to marshal response to yaml"
const httpRequestErrorText = "request returned an error: status code(%d) %w"

func newJsonMarshalError(err error) error {
	return fmt.Errorf("%s: %w", jsonMarshalErrorText, err)
}

func newYamlMarshalError(err error) error {
	return fmt.Errorf("%s: %w", yamlMarshalErrorText, err)
}

func newHttpRequestError(statusCode int, err error) error {
	return fmt.Errorf(httpRequestErrorText, statusCode, err)
}
