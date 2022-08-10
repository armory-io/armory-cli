package configCmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/armory/armory-cli/pkg/armoryCloud"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/armory/armory-cli/pkg/model/configClient"
	"io"
	"net/http"
)

type (
	ConfigClient struct {
		ArmoryCloudClient *armoryCloud.Client
	}
)

func GetConfigClient(configuration *config.Configuration) *ConfigClient {
	armoryCloudClient := configuration.GetArmoryCloudClient()
	return &ConfigClient{
		ArmoryCloudClient: armoryCloudClient,
	}
}

func (c *ConfigClient) CreateRole(ctx context.Context, request *configClient.CreateRoleRequest) (*configClient.CreateRoleResponse, *http.Response, error) {
	reqBytes, err := json.Marshal(request)
	req, err := c.ArmoryCloudClient.Request(ctx, http.MethodPost, fmt.Sprintf("/roles"), bytes.NewReader(reqBytes))
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.ArmoryCloudClient.Http.Do(req)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, resp, &configError{response: resp}
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}

	var role configClient.CreateRoleResponse
	if err := json.Unmarshal(bodyBytes, &role); err != nil {
		return nil, resp, err
	}
	return &role, resp, nil
}

func (c *ConfigClient) UpdateRole(ctx context.Context, request *configClient.UpdateRoleRequest) (*configClient.UpdateRoleResponse, *http.Response, error) {
	reqBytes, err := json.Marshal(request)
	req, err := c.ArmoryCloudClient.Request(ctx, http.MethodPut, fmt.Sprintf("/roles/%s", request.Name), bytes.NewReader(reqBytes))
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.ArmoryCloudClient.Http.Do(req)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, resp, &configError{response: resp}
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}

	var role configClient.UpdateRoleResponse
	if err := json.Unmarshal(bodyBytes, &role); err != nil {
		return nil, resp, err
	}
	return &role, resp, nil
}

func (c *ConfigClient) GetRoles(ctx context.Context) ([]model.RoleConfig, *http.Response, error) {
	req, err := c.ArmoryCloudClient.Request(ctx, http.MethodGet, fmt.Sprintf("/roles"), nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.ArmoryCloudClient.Http.Do(req)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, resp, &configError{response: resp}
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}

	var roles = make([]model.RoleConfig, 10)
	if err := json.Unmarshal(bodyBytes, &roles); err != nil {
		return nil, resp, err
	}
	return roles, resp, nil
}

type configError struct {
	response *http.Response
}

func (d *configError) Error() string {
	responseBytes, err := io.ReadAll(d.response.Body)
	if err != nil {
		return fmt.Sprintf("could not read HTTP response body: %s", err)
	}
	return string(responseBytes)
}
