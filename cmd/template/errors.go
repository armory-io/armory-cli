package template

import (
	"errors"
	"fmt"
)

const canaryTemplateBuildErrorText = "error trying to build canary template: %w"
const canaryTemplateParseErrorText = "error trying to parse canary template: %w"

const blueGreenTemplateParseErrorText = "error trying to parse bluegreen template: %w"

var ErrUnknownFeature = errors.New("unknown feature specified for template")

func newUnknownFeatureError(feature string) error {
	return fmt.Errorf("%w: %s", ErrUnknownFeature, feature)
}

func newCanaryBuildTemplateError(err error) error {
	return fmt.Errorf(canaryTemplateBuildErrorText, err)
}

func newCanaryParseTemplateError(err error) error {
	return fmt.Errorf(canaryTemplateParseErrorText, err)
}

func newBlueGreenParseTemplateError(err error) error {
	return fmt.Errorf(blueGreenTemplateParseErrorText, err)
}
