package configCmd

import (
	"github.com/armory/armory-cli/pkg/model"
	"github.com/armory/armory-cli/pkg/model/configClient"
)

func UpdateRolesRequest(config *model.RoleConfig) (*configClient.UpdateRoleRequest, error) {
	req := configClient.UpdateRoleRequest{
		Name:   config.Name,
		Tenant: config.Tenant,
		Grants: config.Grants,
	}
	return &req, nil
}

func DeleteRolesRequest(roleName string) (*configClient.DeleteRoleRequest, error) {
	req := configClient.DeleteRoleRequest{
		Name: roleName,
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
