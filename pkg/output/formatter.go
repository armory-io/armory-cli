package output

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
)

type Formatter func(interface{}) (string, error)


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
		return MarshalToJson
	}
}

func MarshalToJson(input interface{}) (string, error) {
	pretty, err := json.MarshalIndent(input, "", " ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal response to json: %v", err)
	}
	return string(pretty), nil
}

func MarshalToYaml(input interface{}) (string, error) {
	pretty, err := yaml.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response to yaml: %v", err)
	}
	return string(pretty), nil
}