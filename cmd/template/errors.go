package template

import (
	"errors"
)

var (
	ErrCanaryTemplateBuild    = errors.New("error trying to build canary template")
	ErrBlueGreenTemplateParse = errors.New("error trying to parse bluegreen template")
	ErrUnknownFeature         = errors.New("unknown feature specified for template")
	ErrCanaryTemplateParse    = errors.New("error trying to parse canary template")
)
