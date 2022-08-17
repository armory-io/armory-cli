package config

import "fmt"

const (
	errReadingYamlFileText            = "error trying to read the YAML file: %w"
	errInvalidConfigurationObjectText = "error invalid configuration object: %w"
	errUpdateRoleText                 = "error trying to update role: %w"
	errCreatingRoleText               = "error trying to create role: %w"
	errDeletingRoleText               = "error trying to delete role: %w"
	errGettingRolesText               = "error getting existing roles: %w"
	errParsingGetConfigResponseText   = "error trying to parse response: %w"
)

func newErrorReadingYamlFile(err error) error {
	return fmt.Errorf(errReadingYamlFileText, err)
}

func newErrorInvalidConfigurationObject(err error) error {
	return fmt.Errorf(errInvalidConfigurationObjectText, err)
}

func newErrorUpdateRole(err error) error {
	return fmt.Errorf(errUpdateRoleText, err)
}

func newErrorCreatingRole(err error) error {
	return fmt.Errorf(errCreatingRoleText, err)
}

func newErrorDeletingRole(err error) error {
	return fmt.Errorf(errDeletingRoleText, err)
}

func newErrorGettingRoles(err error) error {
	return fmt.Errorf(errGettingRolesText, err)
}

func newErrorParsingGetConfigResponse(err error) error {
	return fmt.Errorf(errParsingGetConfigResponseText, err)
}
