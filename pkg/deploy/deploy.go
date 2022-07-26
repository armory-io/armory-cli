package deploy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/armory-io/deploy-engine/api"
	"github.com/armory/armory-cli/cmd/version"
	"io"
	"net/http"
	"net/url"
	"os"
)

type (
	Client struct {
		Context       context.Context
		configuration *Configuration
		http          *http.Client
	}

	Configuration struct {
		Host           string
		Scheme         string
		DefaultHeaders map[string]string
		UserAgent      string
	}
)

var source = "armory-cli"

func NewDeployClient(armoryCloudAddr *url.URL, token string) (*Client, error) {
	if val, present := os.LookupEnv("ARMORY_DEPLOYORIGIN"); present {
		source = val
	}

	productVersion := fmt.Sprintf("%s/%s", source, version.Version)

	return &Client{
		Context: context.Background(),
		http:    http.DefaultClient,
		configuration: &Configuration{
			Host:   armoryCloudAddr.Host,
			Scheme: armoryCloudAddr.Scheme,
			DefaultHeaders: map[string]string{
				"Authorization":   fmt.Sprintf("Bearer %s", token),
				"Content-Type":    "application/json",
				"X-Armory-Client": productVersion,
			},
			UserAgent: productVersion,
		},
	}, nil
}

func (c *Client) request(ctx context.Context, method string, path string, body io.Reader) (*http.Request, error) {
	u := &url.URL{
		Scheme: c.configuration.Scheme,
		Host:   c.configuration.Host,
		Path:   path,
	}

	request, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}

	request.Header.Add("User-Agent", c.configuration.UserAgent)
	for key, value := range c.configuration.DefaultHeaders {
		request.Header.Add(key, value)
	}

	return request, nil
}

func (c *Client) PipelineStatus(ctx context.Context, pipelineID string) (*api.PipelineStatusResponse, *http.Response, error) {
	req, err := c.request(ctx, http.MethodGet, fmt.Sprintf("/pipelines/%s", pipelineID), nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, resp, &deployError{response: resp}
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}

	var pipeline api.PipelineStatusResponse
	if err := json.Unmarshal(bodyBytes, &pipeline); err != nil {
		return nil, resp, err
	}
	return &pipeline, resp, nil
}

func (c *Client) DeploymentStatus(ctx context.Context, deploymentID string) (*api.DeploymentStatusResponse, *http.Response, error) {
	req, err := c.request(ctx, http.MethodGet, fmt.Sprintf("/deployments/%s", deploymentID), nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, resp, &deployError{response: resp}
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}

	var deployment api.DeploymentStatusResponse
	if err := json.Unmarshal(bodyBytes, &deployment); err != nil {
		return nil, resp, err
	}
	return &deployment, resp, nil
}

func (c *Client) StartPipeline(ctx context.Context, request *api.StartPipelineRequest) (*api.StartPipelineResponse, *http.Response, error) {
	reqBytes, err := json.Marshal(request)
	if err != nil {
		return nil, nil, err
	}

	req, err := c.request(ctx, http.MethodPost, "/pipelines/kubernetes", bytes.NewReader(reqBytes))
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, resp, &deployError{response: resp}
	}

	if resp.StatusCode != http.StatusAccepted {
		return nil, resp, &deployError{response: resp}
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}

	var startResponse api.StartPipelineResponse
	if err := json.Unmarshal(bodyBytes, &startResponse); err != nil {
		return nil, resp, err
	}
	return &startResponse, resp, nil
}

type deployError struct {
	response *http.Response
}

func (d *deployError) Error() string {
	responseBytes, err := io.ReadAll(d.response.Body)
	if err != nil {
		return fmt.Sprintf("could not read HTTP response body: %s", err)
	}
	return string(responseBytes)
}
