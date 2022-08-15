package model

type ConfigurationConfig struct {
	AllowAutoDelete bool         `yaml:"allowAutoDelete"`
	Roles           []RoleConfig `yaml:"roles,omitempty"`
}

type ConfiguationOutput struct {
	Roles []RoleConfig `yaml:"roles,omitempty"`
}

type RoleConfig struct {
	Name   string        `yaml:"name,omitempty"`
	Tenant string        `yaml:"tenant,omitempty"`
	Grants []GrantConfig `yaml:"grants,omitempty"`
}

type GrantConfig struct {
	Type       string `yaml:"type,omitempty"`
	Resource   string `yaml:"resource,omitempty"`
	Permission string `yaml:"permission,omitempty"`
}
