package deploy

import (
	"context"
	"github.com/armory/armory-cli/internal/graphql"
	"github.com/stretchr/testify/assert"
	log "go.uber.org/zap"
	"testing"
	"time"
)

func TestWaiter(t *testing.T) {
	cases := []struct {
		name           string
		pipelineClient func(ctx context.Context, pipelineID string, callCount int) (*graphql.Pipeline, error)
		assertion      func(t *testing.T, err error, callCount int)
	}{
		{
			name: "pipeline is not found, probably because it is a legacy pipeline",
			pipelineClient: func(ctx context.Context, pipelineID string, callCount int) (*graphql.Pipeline, error) {
				return nil, graphql.ErrNotFound
			},
			assertion: func(t *testing.T, err error, callCount int) {
				assert.Equal(t, 1, callCount)
				assert.NoError(t, err)
			},
		},
		{
			name: "pipeline is RUNNING",
			pipelineClient: func(ctx context.Context, pipelineID string, callCount int) (*graphql.Pipeline, error) {
				return &graphql.Pipeline{
					Status: "RUNNING",
				}, nil
			},
			assertion: func(t *testing.T, err error, callCount int) {
				assert.Equal(t, 1, callCount)
				assert.NoError(t, err)
			},
		},
		{
			name: "pipeline is QUEUED, then RUNNING",
			pipelineClient: func(ctx context.Context, pipelineID string, callCount int) (*graphql.Pipeline, error) {
				switch callCount {
				case 1:
					return &graphql.Pipeline{
						Status: "QUEUED",
					}, nil
				default:
					return &graphql.Pipeline{
						Status: "RUNNING",
					}, nil
				}
			},
			assertion: func(t *testing.T, err error, callCount int) {
				assert.Equal(t, 2, callCount)
				assert.NoError(t, err)
			},
		},
		{
			name: "pipeline is QUEUED, then blocked by another pipeline",
			pipelineClient: func(ctx context.Context, pipelineID string, callCount int) (*graphql.Pipeline, error) {
				switch callCount {
				case 1:
					return &graphql.Pipeline{
						Status: "QUEUED",
					}, nil
				default:
					return &graphql.Pipeline{
						Status: "QUEUED",
						BlockedByPipeline: &graphql.Pipeline{
							ID: "blocked-by-pipeline",
						},
					}, nil
				}
			},
			assertion: func(t *testing.T, err error, callCount int) {
				assert.Equal(t, 2, callCount)
				assert.NoError(t, err)
			},
		},
		{
			name: "pipeline is REJECTED",
			pipelineClient: func(ctx context.Context, pipelineID string, callCount int) (*graphql.Pipeline, error) {
				return &graphql.Pipeline{
					Status: "REJECTED",
					BlockedByPipeline: &graphql.Pipeline{
						ID: "blocked-by-pipeline",
					},
				}, nil
			},
			assertion: func(t *testing.T, err error, callCount int) {
				assert.Equal(t, 1, callCount)
				assert.ErrorContains(t, err, "Failed to start deployment")
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			mock := &mockPipelineClient{
				fn: c.pipelineClient,
			}
			waiter := &Waiter{
				client:              mock,
				log:                 log.S(),
				tick:                1 * time.Millisecond,
				cloudConsoleBaseURL: "",
			}

			err := waiter.WaitForPipelineToBeProcessed(ctx, "")
			c.assertion(t, err, mock.callCount)
		})
	}
}

type mockPipelineClient struct {
	callCount int
	fn        func(ctx context.Context, pipelineID string, callCount int) (*graphql.Pipeline, error)
}

func (c *mockPipelineClient) GetPipeline(ctx context.Context, pipelineID string) (*graphql.Pipeline, error) {
	c.callCount++
	return c.fn(ctx, pipelineID, c.callCount)
}
