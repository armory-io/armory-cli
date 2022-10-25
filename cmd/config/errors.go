package config

import (
	"errors"
)

var (
	ErrReadingYamlFile            = errors.New("error trying to read the YAML file")
	ErrInvalidConfigurationObject = errors.New("error invalid configuration object")
	ErrUpdateRole                 = errors.New("error trying to update role")
	ErrCreatingRole               = errors.New("error trying to create role")
	ErrDeletingRole               = errors.New("error trying to delete role")
	ErrGettingRoles               = errors.New("error getting existing roles")
	ErrCreatingEnvironment        = errors.New("error trying to create environment")
	ErrGettingEnvironments        = errors.New("error getting environments")
	ErrParsingGetConfigResponse   = errors.New("error trying to parse response")
)
