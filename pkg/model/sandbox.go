package model

type (
	SandboxCluster struct {
		ID                  string  `json:"id"`
		IP                  string  `json:"ip"`
		DNS                 string  `json:"dns"`
		Status              string  `json:"status"`
		CreatedAt           string  `json:"createdAt"`
		ExpiresAt           string  `json:"expiresAt"`
		PercentComplete     float32 `json:"percentComplete"`
		NextPercentComplete float32 `json:"nextPercentComplete"`
	}

	CreateSandboxRequest struct {
		AgentIdentifier string `json:"agentIdentifier"`
		ClientId        string `json:"clientId"`
		ClientSecret    string `json:"clientSecret"`
	}

	CreateSandboxResponse struct {
		ClusterId string `json:"clusterId"`
	}

	SandboxSaveData struct {
		SandboxCluster        SandboxCluster        `json:"cluster"`
		AgentIdentifier       string                `json:"agentIdentifier"`
		CreateSandboxResponse CreateSandboxResponse `json:"response"`
	}
)
