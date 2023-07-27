package deploy

import (
	"bytes"
	api "github.com/armory-io/deploy-engine/pkg/api"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/jarcoal/httpmock"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	testPipeline = api.PipelineStatusResponse{
		ID:               "12345",
		Application:      "app",
		StartedAtIso8601: time.Time{}.Format(time.RFC3339),
		Status:           api.WorkflowStatusRunning,
		Steps: []*api.PipelineStep{
			{
				Type:   "deployment",
				Status: api.WorkflowStatusRunning,
				Deployment: &api.PipelineDeploymentStepResponse{
					ID: "5678",
				},
			},
		},
	}
)

func TestDeployStatus(t *testing.T) {
	cases := []struct {
		name              string
		responder         func() (httpmock.Responder, error)
		format            string
		args              []string
		assertion         func(t *testing.T, reader io.Reader)
		expectErrContains string
	}{
		{
			name: "json format success",
			responder: func() (httpmock.Responder, error) {
				return httpmock.NewJsonResponder(http.StatusOK, testPipeline)
			},
			format: "json",
			args: []string{
				"status",
				"--test=true",
				"--deploymentId=12345",
			},
			assertion: func(t *testing.T, reader io.Reader) {
				output, err := io.ReadAll(reader)
				assert.NoError(t, err)

				expected, err := os.ReadFile("./resources/status/status.json")
				assert.NoError(t, err)

				assert.Equal(t, string(expected), string(output))
			},
		},
		{
			name: "yaml format success",
			responder: func() (httpmock.Responder, error) {
				return httpmock.NewJsonResponder(http.StatusOK, testPipeline)
			},
			format: "yaml",
			args: []string{
				"status",
				"--test=true",
				"--deploymentId=12345",
			},
			assertion: func(t *testing.T, reader io.Reader) {
				output, err := io.ReadAll(reader)
				assert.NoError(t, err)

				expected, err := os.ReadFile("./resources/status/status.yaml")
				assert.NoError(t, err)

				assert.Equal(t, string(expected), string(output))
			},
		},
		{
			name: "default format success",
			responder: func() (httpmock.Responder, error) {
				return httpmock.NewJsonResponder(http.StatusOK, testPipeline)
			},
			args: []string{
				"status",
				"--test=true",
				"--deploymentId=12345",
			},
			assertion: func(t *testing.T, reader io.Reader) {
				output, err := io.ReadAll(reader)
				assert.NoError(t, err)

				expected, err := os.ReadFile("./resources/status/status.txt")
				assert.NoError(t, err)

				assert.Equal(t, string(expected), string(output))
			},
		},
		{
			name: "default format, pipeline awaiting approval",
			responder: func() (httpmock.Responder, error) {
				return httpmock.NewJsonResponder(http.StatusOK, api.PipelineStatusResponse{
					Application:      "app",
					Status:           api.WorkflowStatusAwaitingApproval,
					StartedAtIso8601: time.Time{}.Format(time.RFC3339),
				})
			},
			args: []string{
				"status",
				"--test=true",
				"--deploymentId=12345",
			},
			assertion: func(t *testing.T, reader io.Reader) {
				output, err := io.ReadAll(reader)
				assert.NoError(t, err)

				expected, err := os.ReadFile("./resources/status/status_awaiting_approval.txt")
				assert.NoError(t, err)

				assert.Equal(t, string(expected), string(output))
			},
		},
		{
			name: "default format, pipeline paused",
			responder: func() (httpmock.Responder, error) {
				return httpmock.NewJsonResponder(http.StatusOK, api.PipelineStatusResponse{
					Application:      "app",
					Status:           api.WorkflowStatusPaused,
					StartedAtIso8601: time.Time{}.Format(time.RFC3339),
					Steps: []*api.PipelineStep{{
						Type:   "pause",
						Status: api.WorkflowStatusPaused,
						Pause: &api.PauseStepResponse{
							Duration: 5,
							Unit:     api.TimeUnitMinutes,
						},
					}},
				})
			},
			args: []string{
				"status",
				"--test=true",
				"--deploymentId=12345",
			},
			assertion: func(t *testing.T, reader io.Reader) {
				output, err := io.ReadAll(reader)
				assert.NoError(t, err)

				expected, err := os.ReadFile("./resources/status/status_paused.txt")
				assert.NoError(t, err)

				assert.Equal(t, string(expected), string(output))
			},
		},
		{
			name: "failed to fetch",
			responder: func() (httpmock.Responder, error) {
				return httpmock.NewJsonResponder(http.StatusInternalServerError, `{"code":2, "message":"invalid operation", "details":[]}`)
			},
			args: []string{
				"status",
				"--test=true",
				"--deploymentId=12345",
			},
			expectErrContains: "invalid operation",
		},
		{
			name: "invalid arguments",
			args: []string{
				"status",
			},
			expectErrContains: "required flag(s) \"deploymentId\" not set",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			if c.responder != nil {
				responder, err := c.responder()
				assert.NoError(t, err)
				httpmock.RegisterResponder("GET", "https://localhost/pipelines/12345", responder)
			}

			cmd := NewDeployCmd(config.New(&config.Input{
				AccessToken:  lo.ToPtr("some-token"),
				ApiAddr:      lo.ToPtr("https://localhost"),
				ClientId:     lo.ToPtr(""),
				ClientSecret: lo.ToPtr(""),
				OutFormat:    lo.ToPtr(c.format),
				IsTest:       lo.ToPtr(true),
			}))

			writer := bytes.NewBufferString("")
			cmd.SetOut(writer)
			cmd.SetArgs(c.args)

			if c.expectErrContains != "" {
				assert.ErrorContains(t, cmd.Execute(), c.expectErrContains)
			} else {
				assert.NoError(t, cmd.Execute())
				c.assertion(t, writer)
			}
		})
	}
}
