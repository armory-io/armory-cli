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

type sandbox struct {
	ArmoryCloudClient *armoryCloud.Client
}

// newSandbox returns the API for Sandbox operations
func newSandbox(c *ConfigClient) *sandbox {
	return &sandbox{
		ArmoryCloudClient: c.ArmoryCloudClient,
	}
}

func (s *sandbox) Create(ctx context.Context, request *model.CreateSandboxRequest) (*model.CreateSandboxResponse, error) {
	reqBytes, err := json.Marshal(request)
	req, err := s.ArmoryCloudClient.SimpleRequest(ctx, http.MethodPost, fmt.Sprintf("/sandbox/clusters"), bytes.NewReader(reqBytes))
	if err != nil {
		return nil, err
	}

	resp, err := s.ArmoryCloudClient.Http.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, &ConfigError{response: resp}
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var createSandboxResponse model.CreateSandboxResponse
	if err := json.Unmarshal(bodyBytes, &createSandboxResponse); err != nil {
		return nil, err
	}
	return &createSandboxResponse, nil
}

func (s *sandbox) Get(ctx context.Context, clusterId string) (*model.SandboxCluster, error) {
	req, err := s.ArmoryCloudClient.SimpleRequest(ctx, http.MethodGet, fmt.Sprintf("/sandbox/clusters/%s", clusterId), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.ArmoryCloudClient.Http.Do(req)
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

	var sandboxCluster model.SandboxCluster
	if err := json.Unmarshal(bodyBytes, &sandboxCluster); err != nil {
		return nil, err
	}
	return &sandboxCluster, nil
}
