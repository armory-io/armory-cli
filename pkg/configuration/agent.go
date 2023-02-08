package configuration

import (
	"context"
	"encoding/json"
	"github.com/armory/armory-cli/pkg/armoryCloud"
	"github.com/armory/armory-cli/pkg/model"
	"net/http"
)

// AgentInterface has methods to work with Agent resources.
type AgentInterface interface {
	List(ctx context.Context) ([]model.Agent, error)
}

// agents implements AgentInterface
type agents struct {
	ArmoryCloudClient *armoryCloud.Client
}

// newAgents returns a agents
func newAgents(c *ConfigClient) *agents {
	return &agents{
		ArmoryCloudClient: c.ArmoryCloudClient,
	}
}

func (a *agents) List(ctx context.Context) ([]model.Agent, error) {
	req, err := a.ArmoryCloudClient.SimpleRequest(ctx, http.MethodGet, "/identity/connected-agents", nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.ArmoryCloudClient.Http.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &configError{response: resp}
	}

	var agents []model.Agent
	if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
		return nil, err
	}

	return agents, nil
}
