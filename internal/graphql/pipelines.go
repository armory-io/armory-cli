package graphql

import (
	"context"
	"errors"
	"github.com/armory/armory-cli/internal/clierr"
	"github.com/armory/armory-cli/internal/clierr/exitcodes"
	"github.com/machinebox/graphql"
)

type (
	Pipeline struct {
		ID                 string      `json:"id"`
		Status             string      `json:"status"`
		BlockedByPipeline  *Pipeline   `json:"blockedByPipeline"`
		ReplacedByPipeline *Pipeline   `json:"replacedByPipeline"`
		Application        Application `json:"application"`
	}

	Application struct {
		Name string `json:"name"`
	}
)

const getPipelineByIDQuery = `
  query ($pipelineID: uuid!) {
    pipelineById(id: $pipelineID) {
      id
      status
      blockedByPipeline {
        id
      }
      application {
        name
      }
    }
  }
`

func (c *Client) GetPipeline(ctx context.Context, pipelineID string) (*Pipeline, error) {
	request := graphql.NewRequest(getPipelineByIDQuery)
	request.Var("pipelineID", pipelineID)

	requestID := c.newRequestID()

	var pipelineByIDResponse struct {
		PipelineByID *Pipeline `json:"pipelineById"`
	}
	if err := c.doGraphQLRequest(ctx, requestID, request, &pipelineByIDResponse); err != nil {
		return nil, errors.Join(clierr.NewError(
			"Could not fetch deployment",
			requestID,
			err,
			exitcodes.Error,
		), err)
	}

	if pipelineByIDResponse.PipelineByID == nil {
		return nil, clierr.NewError(
			"Could not fetch deployment",
			requestID,
			ErrNotFound,
			exitcodes.Error,
		)
	}

	return pipelineByIDResponse.PipelineByID, nil
}
