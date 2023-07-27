package output

import (
	"encoding/json"
	"fmt"
	errorUtils "github.com/armory/armory-cli/pkg/errors"
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

func GetFormatterForOutputType(outputFormat Type) Formatter {
	switch {
	case outputFormat == Json:
		return MarshalToJson
	case outputFormat == Yaml:
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

	pretty, err := json.MarshalIndent(input.Get(), "", "  ")
	if err != nil {
		return getErrorAsJson(err), errorUtils.NewWrappedError(ErrJsonMarshal, err)
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
		return "", errorUtils.NewWrappedError(ErrYamlMarshal, err)
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
			errContext := fmt.Sprintf(": status code(%d)", input.GetHttpResponse().StatusCode)
			err = errorUtils.NewWrappedErrorWithDynamicContext(ErrHttpRequest, err, errContext)
		}
	}

	return err
}
