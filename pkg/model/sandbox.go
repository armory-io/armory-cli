package model

import (
	"encoding/json"
	"io/ioutil"
)

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

	SandboxClusterSaveData struct {
		SandboxCluster        SandboxCluster        `json:"cluster"`
		AgentIdentifier       string                `json:"agentIdentifier"`
		CreateSandboxResponse CreateSandboxResponse `json:"response"`
	}
)

// WriteToFile stores the data for debugging info or future use by other commands
func (d *SandboxClusterSaveData) WriteToFile(fileLocation string) error {
	data, err := json.MarshalIndent(d, "", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fileLocation, data, 0644)
}
