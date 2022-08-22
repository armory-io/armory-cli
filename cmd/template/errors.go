package template

import (
	"errors"
	"fmt"
)

const (
	errCanaryTemplateBuildText    = "error trying to build canary template: %w"
	errCanaryTemplateParseText    = "error trying to parse canary template: %w"
	errBlueGreenTemplateParseText = "error trying to parse bluegreen template: %w"
)

var (
	ErrUnknownFeature      = errors.New("unknown feature specified for template")
	ErrCanaryTemplateParse = errors.New("error trying to parse canary template")
)

func newUnknownFeatureError(feature string) error {
	return fmt.Errorf("%w: %s", ErrUnknownFeature, feature)
}

func newCanaryBuildTemplateError(err error) error {
	return fmt.Errorf(errCanaryTemplateBuildText, err)
}

func newCanaryParseTemplateError(err error) error {
	return fmt.Errorf(errCanaryTemplateParseText, err)
}

func newBlueGreenParseTemplateError(err error) error {
	return fmt.Errorf(errBlueGreenTemplateParseText, err)
}
