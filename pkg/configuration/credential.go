package configuration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/armory/armory-cli/pkg/armoryCloud"
	"github.com/armory/armory-cli/pkg/model"
	"io"
	"net/http"
)

// CredentialInterface has methods to work with Credentials resources.
type CredentialInterface interface {
	AddRoles(ctx context.Context, request *model.Credential, roles []string) (*[]model.RoleConfig, error)
	Create(ctx context.Context, credential *model.Credential) (*model.Credential, error)
	Delete(ctx context.Context, credential *model.Credential) error
	List(ctx context.Context) ([]*model.Credential, error)
}

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
	req, err := c.ArmoryCloudClient.SimpleRequest(ctx, http.MethodPut, fmt.Sprintf("/credentials/%s/roles", request.ID), bytes.NewReader(reqBytes))
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
	req, err := c.ArmoryCloudClient.SimpleRequest(ctx, http.MethodPost, "/credentials", bytes.NewReader(reqBytes))
	if err != nil {
		return nil, err
	}

	resp, err := c.ArmoryCloudClient.Http.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, &configError{response: resp}
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
		return &configError{response: resp}
	}

	return nil
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
		return nil, &configError{response: resp}
	}

	var credentials []*model.Credential
	if err := json.NewDecoder(resp.Body).Decode(&credentials); err != nil {
		return nil, err
	}

	return credentials, nil
}
