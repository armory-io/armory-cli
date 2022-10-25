package configuration

import (
	"github.com/armory/armory-cli/pkg/model"
	"github.com/armory/armory-cli/pkg/model/configClient"
)

func UpdateRolesRequest(id, tenant string, grants []model.GrantConfig) (*configClient.UpdateRoleRequest, error) {
	req := configClient.UpdateRoleRequest{
		ID:     id,
		Tenant: tenant,
		Grants: grants,
	}
	return &req, nil
}

func DeleteRolesRequest(roleID string) (*configClient.DeleteRoleRequest, error) {
	req := configClient.DeleteRoleRequest{
		ID: roleID,
	}
	return &req, nil
}

func CreateRoleRequest(config *model.RoleConfig) (*configClient.CreateRoleRequest, error) {
	req := configClient.CreateRoleRequest{
		Name:   config.Name,
		Tenant: config.Tenant,
		Grants: config.Grants,
	}
	return &req, nil
}

func CreateEnvironmentRequest(environment string) (configClient.CreateEnvironmentRequest, error) {
	req := configClient.CreateEnvironmentRequest{
		Name: environment,
	}
	return req, nil
}
