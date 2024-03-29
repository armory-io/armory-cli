package configuration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/armory/armory-cli/pkg/armoryCloud"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/hashicorp/go-retryablehttp"
)

type sandbox struct {
	ArmoryCloudClient *armoryCloud.Client
}

// newSandbox returns the API for Sandbox operations
func newSandbox(c *ConfigClient) *sandbox {
	client := &retryablehttp.Client{
		HTTPClient:   c.ArmoryCloudClient.Http,
		RetryWaitMin: 200 * time.Millisecond,
		RetryWaitMax: 2 * time.Second,
		RetryMax:     5,
		CheckRetry:   SandboxRetryPolicy,
		Backoff:      retryablehttp.DefaultBackoff,
	}
	c.ArmoryCloudClient.Http = client.StandardClient()
	return &sandbox{
		ArmoryCloudClient: c.ArmoryCloudClient,
	}
}

func SandboxRetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	// do not retry on context.Canceled or context.DeadlineExceeded
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	//retryablehttp throws error after max retries, so we may find an empty resp
	if err != nil && resp == nil {
		return false, err
	}
	// don't propagate other errors
	if resp.StatusCode >= 404 {
		return true, err
	}

	return false, err
}

func (s *sandbox) Create(ctx context.Context, request *model.CreateSandboxRequest) (*model.CreateSandboxResponse, error) {
	reqBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, err := s.ArmoryCloudClient.SimpleRequest(ctx, http.MethodPost, "/sandbox/clusters", bytes.NewReader(reqBytes))
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
