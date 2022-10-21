package configClient

type Environment struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	OrgID string `json:"orgId"`
}

type CreateTenantRequest struct {
	Name string `json:"name"`
}

type CreateTenantResponse struct {
	ID    string `json:"tenantID"`
	Name  string `json:"name"`
	Roles []any  `json:"roles"`
}
