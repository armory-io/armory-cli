package configClient

type Environment struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	OrgID string `json:"orgId"`
}

type CreateEnvironmentRequest struct {
	Name string `json:"name"`
}

type CreateEnvironmentResponse struct {
	ID    string `json:"tenantId"`
	Name  string `json:"name"`
	Roles []any  `json:"roles"`
}
