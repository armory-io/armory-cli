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
)

func (c *SandboxCluster) SaveData(fileLocation string) error {
	data, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fileLocation, data, 0644)
	if err != nil {
		return err
	}
	return nil
}
