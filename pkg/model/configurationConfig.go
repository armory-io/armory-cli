package model

type ConfigurationConfig struct {
	AllowAutoDelete bool         `yaml:"allowAutoDelete"`
	Environments    []string     `yaml:"tenants,omitempty"`
	Roles           []RoleConfig `yaml:"roles,omitempty"`
}

type ConfigurationOutput struct {
	Environments []string     `yaml:"tenants,omitempty"`
	Roles        []RoleConfig `yaml:"roles,omitempty"`
}

type RoleConfig struct {
	ID            string        `json:"id" yaml:"-"`
	EnvID         string        `json:"envId" yaml:"-"`
	Name          string        `yaml:"name,omitempty"`
	Tenant        string        `yaml:"tenant,omitempty"`
	SystemDefined bool          `yaml:"systemDefined,omitempty" json:"systemDefined,omitempty"`
	Grants        []GrantConfig `yaml:"grants,omitempty"`
}

type GrantConfig struct {
	Type       string `yaml:"type,omitempty"`
	Resource   string `yaml:"resource,omitempty"`
	Permission string `yaml:"permission,omitempty"`
}
