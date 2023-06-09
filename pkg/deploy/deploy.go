package deploy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/armory-io/deploy-engine/pkg/api"
	"github.com/armory/armory-cli/pkg/armoryCloud"
	"github.com/armory/armory-cli/pkg/config"
)

type (
	Client struct {
		ArmoryCloudClient *armoryCloud.Client
	}
)

func NewClient(configuration *config.Configuration) *Client {
	armoryCloudClient := configuration.GetArmoryCloudClient()
	return &Client{armoryCloudClient}
}

func (c *Client) PipelineStatus(ctx context.Context, pipelineID string) (*api.PipelineStatusResponse, *http.Response, error) {
	req, err := c.ArmoryCloudClient.SimpleRequest(ctx, http.MethodGet, fmt.Sprintf("/pipelines/%s", pipelineID), nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.ArmoryCloudClient.Http.Do(req)
	if err != nil {
		return nil, resp, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, resp, &deployError{bodyBytes}
	}

	var pipeline api.PipelineStatusResponse
	if err := json.Unmarshal(bodyBytes, &pipeline); err != nil {
		return nil, resp, err
	}
	return &pipeline, resp, nil
}

func (c *Client) DeploymentStatus(ctx context.Context, deploymentID string) (*api.DeploymentStatusResponse, *http.Response, error) {
	req, err := c.ArmoryCloudClient.SimpleRequest(ctx, http.MethodGet, fmt.Sprintf("/deployments/%s", deploymentID), nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.ArmoryCloudClient.Http.Do(req)
	if err != nil {
		return nil, resp, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, resp, &deployError{bodyBytes}
	}

	var deployment api.DeploymentStatusResponse
	if err := json.Unmarshal(bodyBytes, &deployment); err != nil {
		return nil, resp, err
	}
	return &deployment, resp, nil
}

func (c *Client) StartPipeline(ctx context.Context, options StartPipelineOptions) (*api.StartPipelineResponse, *http.Response, error) {
	structured, err := options.structuredConfig()
	if err != nil {
		return nil, nil, err
	}

	if structured.Kind == "" {
		return nil, nil, ErrNoKind
	}

	request, err := convertPipelineOptionsToAPIRequest(options)
	if err != nil {
		return nil, nil, err
	}

	reqBytes, err := json.Marshal(request)

	if err != nil {
		return nil, nil, err
	}
	requestOptions := []armoryCloud.RequestOption{
		armoryCloud.WithMethod(http.MethodPost),
		armoryCloud.WithPath(fmt.Sprintf("/pipelines/%s", structured.Kind)),
	}
	for key, val := range options.Headers {
		requestOptions = append(requestOptions, armoryCloud.WithHeader(key, val))
	}
	requestOptions = append(requestOptions, armoryCloud.WithBody(bytes.NewReader(reqBytes)))

	req, err := c.ArmoryCloudClient.Request(ctx, requestOptions...)
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.ArmoryCloudClient.Http.Do(req)
	if err != nil {
		if resp != nil {
			return nil, resp, err
		} else {
			return nil, nil, &networkError{}
		}
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode != http.StatusAccepted {
		return nil, resp, &deployError{bodyBytes}
	}

	var startResponse api.StartPipelineResponse
	if err := json.Unmarshal(bodyBytes, &startResponse); err != nil {
		return nil, resp, err
	}
	return &startResponse, resp, nil
}

func (c *Client) GetArmoryCloudClient() *armoryCloud.Client {
	return c.ArmoryCloudClient
}

// deployError keeps a byte slice for the error response. The old implementation stored a pointer to the response itself;
// this would fail if the request resources were greater than what the http pkg was keeping on-hand per request. This
// is because io.ReadAll calls Read on the response body, and the body streams content from the connection. If network
// resources get cleaned up before deployError.Error() was called it would return the "context cancelled" error instead of
// whatever reason the server had returned in the response body.
type deployError struct {
	responseBytes []byte
}

func (d *deployError) Error() string {
	return string(d.responseBytes)
}

type networkError struct {
}

func (d *networkError) Error() string {
	return "Unable to reach Armory Continuous Delivery as A Service. Please check your internet connection and try again. "
}
