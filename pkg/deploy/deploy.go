package deploy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/armory-io/deploy-engine/api"
	"github.com/armory/armory-cli/pkg/armoryCloud"
	"github.com/armory/armory-cli/pkg/config"
	"io"
	"net/http"
)

type (
	DeployClient struct {
		ArmoryCloudClient *armoryCloud.Client
	}
)

func GetDeployClient(configuration *config.Configuration) *DeployClient {
	armoryCloudClient := configuration.GetArmoryCloudClient()
	return &DeployClient{
		ArmoryCloudClient: armoryCloudClient,
	}
}

var source = "armory-cli"

func (c *DeployClient) PipelineStatus(ctx context.Context, pipelineID string) (*api.PipelineStatusResponse, *http.Response, error) {
	req, err := c.ArmoryCloudClient.Request(ctx, http.MethodGet, fmt.Sprintf("/pipelines/%s", pipelineID), nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.ArmoryCloudClient.Http.Do(req)
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

func (c *DeployClient) DeploymentStatus(ctx context.Context, deploymentID string) (*api.DeploymentStatusResponse, *http.Response, error) {
	req, err := c.ArmoryCloudClient.Request(ctx, http.MethodGet, fmt.Sprintf("/deployments/%s", deploymentID), nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.ArmoryCloudClient.Http.Do(req)
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

func (c *DeployClient) StartPipeline(ctx context.Context, request *api.StartPipelineRequest) (*api.StartPipelineResponse, *http.Response, error) {
	reqBytes, err := json.Marshal(request)
	if err != nil {
		return nil, nil, err
	}

	req, err := c.ArmoryCloudClient.Request(ctx, http.MethodPost, "/pipelines/kubernetes", bytes.NewReader(reqBytes))
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.ArmoryCloudClient.Http.Do(req)
	if err != nil {
		if resp != nil {
			return nil, resp, &deployError{response: resp}
		} else {
			return nil, nil, &networkError{}
		}

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

type networkError struct {
}

func (d *networkError) Error() string {
	return "Unable to reach Armory Continuous Delivery as A Service. Please check your internet connection and try again. "
}