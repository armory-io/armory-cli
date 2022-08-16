package config

import "fmt"

const errorReadingYamlFileText = "error trying to read the YAML file: %w"
const invalidConfigurationObjectErrorText = "error invalid configuration object: %w"
const updateRoleErrorText = "error trying to update role: %w"
const creatingRoleErrorText = "error trying to create role: %w"
const deletingRoleErrorText = "error trying to delete role: %w"
const errorGettingRolesText = "error getting existing roles: %w"
const errorParsingGetConfigResponseText = "error trying to parse response: %w"

func newErrorReadingYamlFile(err error) error {
	return fmt.Errorf(errorReadingYamlFileText, err)
}

func newErrorInvalidConfigurationObject(err error) error {
	return fmt.Errorf(invalidConfigurationObjectErrorText, err)
}

func newErrorUpdateRole(err error) error {
	return fmt.Errorf(updateRoleErrorText, err)
}

func newErrorCreatingRole(err error) error {
	return fmt.Errorf(creatingRoleErrorText, err)
}

func newErrorDeletingRole(err error) error {
	return fmt.Errorf(deletingRoleErrorText, err)
}

func newErrorGettingRoles(err error) error {
	return fmt.Errorf(errorGettingRolesText, err)
}

func newErrorParsingGetConfigResponse(err error) error {
	return fmt.Errorf(errorParsingGetConfigResponseText, err)
}
