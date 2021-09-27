package output

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
)

type Formatter func(interface{}, error) (string, error)


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

func DefaultStructToString(input interface{}, err error) (string, error) {
	return fmt.Sprintf("%v",input), nil
}

func MarshalToJson(input interface{}, err error) (string, error) {
	if err != nil {
		return getErrorAsJson(err), err
	}

	pretty, err := json.MarshalIndent(input, "", " ")
	if err != nil {
		return getErrorAsJson(err), fmt.Errorf("failed to marshal response to json: %v", err)
	}
	return string(pretty), nil
}

func getErrorAsJson(err error) string {
	return fmt.Sprintf("{ \"error\": \"%s\" }", err)
}

func MarshalToYaml(input interface{}, err error) (string, error) {
	if err != nil {
		return "", err
	}
	pretty, err := yaml.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response to yaml: %v", err)
	}
	return string(pretty), nil
}