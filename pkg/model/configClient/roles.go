package configClient

import "github.com/armory/armory-cli/pkg/model"

type CreateRoleRequest struct {
	Name   string              `yaml:"name,omitempty"`
	Tenant string              `yaml:"tenant,omitempty"`
	Grants []model.GrantConfig `yaml:"grants,omitempty"`
}

type CreateRoleResponse struct {
	Name   string              `yaml:"name,omitempty"`
	Tenant string              `yaml:"tenant,omitempty"`
	Grants []model.GrantConfig `yaml:"grants,omitempty"`
}

type UpdateRoleRequest struct {
	Name   string              `yaml:"name,omitempty"`
	Tenant string              `yaml:"tenant,omitempty"`
	Grants []model.GrantConfig `yaml:"grants,omitempty"`
}

type UpdateRoleResponse struct {
	Name   string              `yaml:"name,omitempty"`
	Tenant string              `yaml:"tenant,omitempty"`
	Grants []model.GrantConfig `yaml:"grants,omitempty"`
}

type GetRolesResponse struct {
	Roles []model.RoleConfig `yaml:"roles,omitempty"`
}
