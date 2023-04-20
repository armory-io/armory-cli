package configuration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/armory/armory-cli/pkg/armoryCloud"
	"github.com/armory/armory-cli/pkg/model"
)

// credentials implements CredentialInterface
type credentials struct {
	ArmoryCloudClient *armoryCloud.Client
}

// newCredentials returns a credentials
func newCredentials(c *ConfigClient) *credentials {
	return &credentials{
		ArmoryCloudClient: c.ArmoryCloudClient,
	}
}

func (c *credentials) AddRoles(ctx context.Context, request *model.Credential, roles []string) (*[]model.RoleConfig, error) {
	reqBytes, err := json.Marshal(roles)
	if err != nil {
		return nil, err
	}
	req, err := c.ArmoryCloudClient.SimpleRequest(ctx, http.MethodPut, fmt.Sprintf("/credentials/%s/roles", request.ID), bytes.NewReader(reqBytes))
	if err != nil {
		return nil, err
	}

	resp, err := c.ArmoryCloudClient.Http.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &ConfigError{response: resp}
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var role []model.RoleConfig
	if err := json.Unmarshal(bodyBytes, &role); err != nil {
		return nil, err
	}
	return &role, nil
}

func (c *credentials) Create(ctx context.Context, credential *model.Credential) (*model.Credential, error) {
	reqBytes, err := json.Marshal(credential)
	if err != nil {
		return nil, err
	}
	req, err := c.ArmoryCloudClient.SimpleRequest(ctx, http.MethodPost, "/credentials", bytes.NewReader(reqBytes))
	if err != nil {
		return nil, err
	}

	resp, err := c.ArmoryCloudClient.Http.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, &ConfigError{response: resp}
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bodyBytes, &credential); err != nil {
		return credential, err
	}
	return credential, nil
}

func (c *credentials) Delete(ctx context.Context, credential *model.Credential) error {
	req, err := c.ArmoryCloudClient.SimpleRequest(ctx, http.MethodDelete, fmt.Sprintf("/credentials/%s", credential.ID), nil)
	if err != nil {
		return err
	}

	resp, err := c.ArmoryCloudClient.Http.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return &ConfigError{response: resp}
	}

	return nil
}

func (c *credentials) GetRoles(ctx context.Context, credential *model.Credential) (*[]model.RoleConfig, error) {
	req, err := c.ArmoryCloudClient.SimpleRequest(ctx, http.MethodGet, fmt.Sprintf("/credentials/%s/roles", credential.ID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.ArmoryCloudClient.Http.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &ConfigError{response: resp}
	}

	var roles *[]model.RoleConfig
	if err := json.NewDecoder(resp.Body).Decode(&roles); err != nil {
		return nil, err
	}

	return roles, nil
}

func (c *credentials) List(ctx context.Context) ([]*model.Credential, error) {
	req, err := c.ArmoryCloudClient.SimpleRequest(ctx, http.MethodGet, "/credentials", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.ArmoryCloudClient.Http.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &ConfigError{response: resp}
	}

	var credentials []*model.Credential
	if err := json.NewDecoder(resp.Body).Decode(&credentials); err != nil {
		return nil, err
	}

	return credentials, nil
}
