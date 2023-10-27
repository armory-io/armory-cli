package deploy

import (
	"context"
	"errors"
	"fmt"
	"github.com/armory-io/go-commons/awaitility"
	"github.com/armory/armory-cli/internal/clierr"
	"github.com/armory/armory-cli/internal/clierr/exitcodes"
	"github.com/armory/armory-cli/internal/graphql"
	"github.com/armory/armory-cli/pkg/console"
	"go.uber.org/zap"
	"time"
)

const (
	defaultWaiterTickDuration = 5 * time.Second
)

type (
	Waiter struct {
		client              pipelineClient
		tick                time.Duration
		log                 *zap.SugaredLogger
		cloudConsoleBaseURL string
	}

	pipelineClient interface {
		GetPipeline(ctx context.Context, pipelineID string) (*graphql.Pipeline, error)
	}
)

func NewWaiter(
	gqlclient *graphql.Client,
	cloudConsoleBaseURL string,
) *Waiter {
	return &Waiter{
		client:              gqlclient,
		tick:                defaultWaiterTickDuration,
		cloudConsoleBaseURL: cloudConsoleBaseURL,
	}
}

func (w *Waiter) WaitForPipelineToBeProcessed(ctx context.Context, pipelineID string) error {
	var pipelineErr error

	deadline, _ := ctx.Deadline() // If the deadline isn't provided, then awaitility will error.
	if err := awaitility.Await(w.tick, time.Until(deadline), func() bool {
		pipeline, err := w.client.GetPipeline(ctx, pipelineID)
		if err != nil {
			if errors.Is(err, graphql.ErrNotFound) {
				console.Debugf("Could not find deployment %q\n", pipelineID)
				return true
			}
			console.Stderrf("Could not fetch deployment, will retry in %s; err: %s", w.tick, err)
			return false
		}

		done := w.isPipelineProcessed(pipeline)
		if done {
			pipelineErr = w.checkError(pipeline)
		}
		return done
	}); err != nil {
		return err
	}
	return pipelineErr
}

func (w *Waiter) checkError(pipeline *graphql.Pipeline) error {
	if pipeline.Status == "REJECTED" {
		return clierr.NewError(
			"Failed to start deployment",
			"",
			fmt.Errorf("cannot start deployment for application %q because there is an in-progress deployment: %s/deployments/pipeline/%s", pipeline.Application.Name, w.cloudConsoleBaseURL, pipeline.BlockedByPipeline.ID),
			exitcodes.Error,
		)
	}
	return nil
}

func (w *Waiter) isPipelineProcessed(pipeline *graphql.Pipeline) bool {
	if pipeline.Status == "REJECTED" && pipeline.BlockedByPipeline == nil {
		console.Debugln("The pipeline has been rejected but the blockedBy pipeline field has not yet been set.")
		return false
	}

	return pipeline.Status != "QUEUED" || pipeline.BlockedByPipeline != nil
}
