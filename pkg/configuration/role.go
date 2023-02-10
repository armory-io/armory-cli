package configuration

import (
	"context"
	"encoding/json"
	"github.com/armory/armory-cli/pkg/armoryCloud"
	"github.com/armory/armory-cli/pkg/model"
	"io"
	"net/http"
)

// RolInterface has methods to work with Rol resources.
type RolInterface interface {
	ListForMachinePrincipals(ctx context.Context, environmentId string) ([]model.RoleConfig, error)
}

// roles implements RolInterface
type roles struct {
	ArmoryCloudClient *armoryCloud.Client
}

// newAgents returns a agents
func newRoles(c *ConfigClient) *roles {
	return &roles{
		ArmoryCloudClient: c.ArmoryCloudClient,
	}
}

func (r *roles) ListForMachinePrincipals(ctx context.Context, environmentId string) ([]model.RoleConfig, error) {
	req, err := r.ArmoryCloudClient.SimpleRequest(ctx, http.MethodGet, "/roles", nil)
	if err != nil {
		return nil, err
	}
	queryParams := req.URL.Query()
	queryParams.Add("envId", environmentId)
	queryParams.Add("principalType", "machine")
	req.URL.RawQuery = queryParams.Encode()

	resp, err := r.ArmoryCloudClient.Http.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &configError{response: resp}
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var roles = make([]model.RoleConfig, 10)
	if err := json.Unmarshal(bodyBytes, &roles); err != nil {
		return nil, err
	}
	return roles, nil
}
