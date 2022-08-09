package model

type ConfigurationConfig struct {
	Roles []RoleConfig `yaml:"roles,omitempty"`
}

type RoleConfig struct {
	Name   string        `yaml:"name,omitempty"`
	Tenant string        `yaml:"tenant,omitempty"`
	Grants []GrantConfig `yaml:"grants,omitempty"`
}

type GrantConfig struct {
	Type        string   `yaml:"type,omitempty"`
	Resource    string   `yaml:"resource,omitempty"`
	Permissions []string `yaml:"permissions,omitempty"`
}
