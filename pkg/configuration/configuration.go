package configuration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/armory/armory-cli/pkg/armoryCloud"
	cliconfig "github.com/armory/armory-cli/pkg/config"
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

func NewClient(configuration *cliconfig.Configuration) *ConfigClient {
	armoryCloudClient := configuration.GetArmoryCloudClient()
	return &ConfigClient{
		ArmoryCloudClient: armoryCloudClient,
	}
}

func (c *ConfigClient) CreateRole(ctx context.Context, request *configClient.CreateRoleRequest) (*configClient.CreateRoleResponse, *http.Response, error) {
	reqBytes, err := json.Marshal(request)
	req, err := c.ArmoryCloudClient.SimpleRequest(ctx, http.MethodPost, fmt.Sprintf("/roles"), bytes.NewReader(reqBytes))
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
	req, err := c.ArmoryCloudClient.SimpleRequest(ctx, http.MethodPut, fmt.Sprintf("/roles/%s", request.ID), bytes.NewReader(reqBytes))
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
	req, err := c.ArmoryCloudClient.SimpleRequest(ctx, http.MethodGet, fmt.Sprintf("/roles"), nil)
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

func (c *ConfigClient) DeleteRole(ctx context.Context, request *configClient.DeleteRoleRequest) (*http.Response, error) {
	reqBytes, err := json.Marshal(request)
	req, err := c.ArmoryCloudClient.SimpleRequest(ctx, http.MethodDelete, fmt.Sprintf("/roles/%s", request.ID), bytes.NewReader(reqBytes))
	if err != nil {
		return nil, err
	}

	resp, err := c.ArmoryCloudClient.Http.Do(req)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode != http.StatusNoContent {
		return resp, &configError{response: resp}
	}

	return resp, nil
}

func (c *ConfigClient) GetEnvironments(ctx context.Context) ([]configClient.Environment, error) {
	req, err := c.ArmoryCloudClient.SimpleRequest(ctx, http.MethodGet, "/environments", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.ArmoryCloudClient.Http.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &configError{response: resp}
	}

	var environments []configClient.Environment
	if err := json.NewDecoder(resp.Body).Decode(&environments); err != nil {
		return nil, err
	}

	return environments, nil
}

func (c *ConfigClient) CreateEnvironment(ctx context.Context, request configClient.CreateEnvironmentRequest) (*configClient.CreateEnvironmentResponse, *http.Response, error) {
	reqBytes, err := json.Marshal(request)
	req, err := c.ArmoryCloudClient.SimpleRequest(ctx, http.MethodPost, "/environments", bytes.NewReader(reqBytes))
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

	var environment configClient.CreateEnvironmentResponse
	if err := json.Unmarshal(bodyBytes, &environment); err != nil {
		return nil, resp, err
	}
	return &environment, resp, nil
}
