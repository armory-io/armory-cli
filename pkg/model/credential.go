package model

type Credential struct {
	ID             string   `json:"id,omitempty" yaml:"id,omitempty"`
	Name           string   `json:"name,omitempty" yaml:"name,omitempty"`
	ClientId       string   `json:"clientId,omitempty" yaml:"clientId,omitempty"`
	ClientSecret   string   `json:"clientSecret,omitempty" yaml:"clientSecret,omitempty"`
	CreatedBy      string   `json:"createdBy,omitempty" yaml:"createdBy,omitempty"`
	CreatedFor     string   `json:"createdFor,omitempty" yaml:"createdFor,omitempty"`
	CreatedIso8601 string   `json:"createdIso8601,omitempty" yaml:"createdIso8601,omitempty"`
	Scope          []string `json:"scope,omitempty" yaml:"scope,omitempty"`
	ScopeGroups    []string `json:"scopeGroups,omitempty" yaml:"scopeGroups,omitempty"`
}
