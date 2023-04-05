package configuration

import (
	"context"
	"encoding/json"
	"github.com/armory/armory-cli/pkg/armoryCloud"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/samber/lo"
	"net/http"
)

// agents implements AgentClient
type agents struct {
	ArmoryCloudClient *armoryCloud.Client
}

// newAgents returns a agents
func newAgents(c *ConfigClient) *agents {
	return &agents{
		ArmoryCloudClient: c.ArmoryCloudClient,
	}
}

func (a *agents) Get(ctx context.Context, agentIdentifier string) (*model.Agent, error) {
	req, err := a.ArmoryCloudClient.SimpleRequest(ctx, http.MethodGet, "/identity/connected-agents", nil)
	if err != nil {
		return nil, err
	}

	queryParams := req.URL.Query()
	queryParams.Add("agent-identifier", agentIdentifier)
	req.URL.RawQuery = queryParams.Encode()

	resp, err := a.ArmoryCloudClient.Http.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &ConfigError{response: resp}
	}

	var agents []*model.Agent
	if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
		return nil, err
	}

	agent, exists := lo.Find(agents, func(a *model.Agent) bool {
		return agentIdentifier == a.AgentIdentifier
	})

	if exists {
		return agent, nil
	}

	return nil, nil
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
		return nil, &ConfigError{response: resp}
	}

	var agents []model.Agent
	if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
		return nil, err
	}

	return agents, nil
}
