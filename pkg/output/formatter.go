package output

import (
	"encoding/json"
	"fmt"
	de "github.com/armory-io/deploy-engine/pkg"
	"gopkg.in/yaml.v3"
	_nethttp "net/http"
)

type Formattable interface {
	GetFetchError() error
	GetHttpResponse() *_nethttp.Response
	Get() interface{}
}

type Formatter func(Formattable) (string, error)

type Output struct {
	Formatter Formatter
}

func NewOutput(outputFormatter string) *Output {
	return &Output{
		Formatter: parseOutputFormat(outputFormatter),
	}
}

func parseOutputFormat(outputFormat string) Formatter {
	switch {
	case outputFormat == "json":
		return MarshalToJson
	case outputFormat == "yaml":
		return MarshalToYaml
	default:
		return DefaultStructToString
	}
}

func DefaultStructToString(input Formattable) (string, error) {
	err := getRequestError(input)
	if err != nil {
		return "Encountered request error:", err
	}

	return fmt.Sprintf("%v", input), err
}

func MarshalToJson(input Formattable) (string, error) {
	err := getRequestError(input)
	if err != nil {
		return getErrorAsJson(err), nil
	}

	pretty, err := json.MarshalIndent(input.Get(), "", " ")
	if err != nil {
		return getErrorAsJson(err), fmt.Errorf("failed to marshal response to json: %v", err)
	}
	return string(pretty), nil
}

func getErrorAsJson(err error) string {
	return fmt.Sprintf("{ \"error\": \"%s\" }", err)
}

func MarshalToYaml(input Formattable) (string, error) {
	err := getRequestError(input)
	if err != nil {
		return getErrorAsYaml(err), nil
	}

	pretty, err := yaml.Marshal(input.Get())
	if err != nil {
		return "", fmt.Errorf("failed to marshal response to yaml: %v", err)
	}
	return string(pretty), nil
}

func getErrorAsYaml(err error) string {
	return fmt.Sprintf("error: \"%s", err)
}

func getRequestError(input Formattable) error {
	err := input.GetFetchError()
	if err != nil {
		// don't override the received error unless we have an unexpected http response status
		if input.GetHttpResponse() != nil && input.GetHttpResponse().StatusCode >= 300 {
			openAPIErr := err.(de.GenericOpenAPIError)
			err = fmt.Errorf("request returned an error: status code(%d) %s",
				input.GetHttpResponse().StatusCode, string(openAPIErr.Body()))
		}
	}

	return err
}
