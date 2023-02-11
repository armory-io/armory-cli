package deploy

import (
	"context"
	"fmt"
	"github.com/armory-io/deploy-engine/pkg/api"
	"github.com/armory/armory-cli/pkg/armoryCloud"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/deploy"
	"net/http"
)

type (
	MockDeployClient struct {
		ArmoryCloudClient            *armoryCloud.Client
		RecordedStartPipelineOptions deploy.StartPipelineOptions
		startPipelineMock            StartPipelineMock
	}
	StartPipelineMock = func() (*api.StartPipelineResponse, *http.Response, error)
)

func GetMockDeployClient(configuration *config.Configuration) *MockDeployClient {
	armoryCloudClient := configuration.GetArmoryCloudClient()
	return &MockDeployClient{ArmoryCloudClient: armoryCloudClient}
}

func (c *MockDeployClient) PipelineStatus(ctx context.Context, pipelineID string) (*api.PipelineStatusResponse, *http.Response, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (c *MockDeployClient) DeploymentStatus(ctx context.Context, deploymentID string) (*api.DeploymentStatusResponse, *http.Response, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (c *MockDeployClient) MockStartPipelineResponse(exec func() (*api.StartPipelineResponse, *http.Response, error)) {
	c.startPipelineMock = exec
}

func (c *MockDeployClient) StartPipeline(ctx context.Context, options deploy.StartPipelineOptions) (*api.StartPipelineResponse, *http.Response, error) {
	c.RecordedStartPipelineOptions = options
	if c.startPipelineMock != nil {
		return c.startPipelineMock()
	}
	return nil, nil, nil
}

func (c *MockDeployClient) GetArmoryCloudClient() *armoryCloud.Client {
	return c.ArmoryCloudClient
}
