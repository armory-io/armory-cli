package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDeployError_Error was written and added to prove that our implementation of deployError had
// a bug. For context, read the godoc for deployError. This test _may_ prevent regressions but it is
// probably not useful in the longterm.
func TestDeployError_Error(t *testing.T) {
	expected := strings.Repeat("x", 10_000)
	// mock our http server

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprint(w, expected)
		assert.NoError(t, err)
	}))
	defer svr.Close()

	// resources for making our little http client
	ctx, cancel := context.WithCancel(context.Background())

	req, err := http.NewRequestWithContext(ctx, "GET", svr.URL, nil)
	assert.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	capturedErr := &deployError{bodyBytes}

	// simulate a cleanup of resources, the client is now done with the request so cancel the ctx
	cancel()

	assert.Error(t, capturedErr)
	assert.ErrorContains(t, capturedErr, expected)
}

func TestStartPipeline(t *testing.T) {
	cases := []struct {
		name            string
		yaml            string
		inputHeaders    map[string]string
		expectedPath    string
		expectedHeaders map[string]string
	}{
		{
			name: "kubernetes deployment",
			yaml: `
kind: kubernetes
application: kubernetes-application
`,
			expectedPath: "/pipelines/kubernetes",
			expectedHeaders: map[string]string{
				"Content-Type": "application/vnd.start.kubernetes.pipeline.v2+json",
				"Accept":       "application/vnd.start.kubernetes.pipeline.v2+json",
			},
		},
		{
			name: "no kind specified -> goes to kubernetes endpoint for now",
			yaml: `
application: kubernetes-application
`,
			expectedPath: "/pipelines/kubernetes",
			expectedHeaders: map[string]string{
				"Content-Type": "application/vnd.start.kubernetes.pipeline.v2+json",
				"Accept":       "application/vnd.start.kubernetes.pipeline.v2+json",
			},
		},
		{
			name: "provided headers override default headers",
			yaml: `
application: kubernetes-application
`,
			inputHeaders: map[string]string{
				"Content-Type": "application/vnd.start.kubernetes.pipeline.v2.link+json",
			},
			expectedPath: "/pipelines/kubernetes",
			expectedHeaders: map[string]string{
				"Content-Type": "application/vnd.start.kubernetes.pipeline.v2.link+json",
				"Accept":       "application/vnd.start.kubernetes.pipeline.v2+json",
			},
		},
		{
			name: "lambda deployment",
			yaml: `
kind: lambda
application: lambda-application
`,
			expectedPath: "/pipelines",
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			name: "banana cloud deployment",
			yaml: `
kind: banana cloud
application: lambda-application
`,
			expectedPath: "/pipelines",
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			name: "k8s with enqueueOne strategy",
			yaml: `
kind: kubernetes
application: classic-k8s-app
deploymentConfig:
  ifDeploymentInProgress:
    strategy: enqueueOne
`,
			expectedPath: "/pipelines",
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := context.Background()
			s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				for key, value := range c.expectedHeaders {
					assert.Equal(t, value, request.Header.Get(key))
				}

				assert.Equal(t, c.expectedPath, request.URL.Path)
				writer.WriteHeader(http.StatusAccepted)
				assert.NoError(t, json.NewEncoder(writer).Encode(map[string]string{
					"pipelineId": "1-800-pipelines",
				}))
			}))
			defer s.Close()

			client := NewClient(config.New(&config.Input{
				AccessToken:  lo.ToPtr("my-token"),
				ApiAddr:      lo.ToPtr(s.URL),
				ClientId:     lo.ToPtr(""),
				ClientSecret: lo.ToPtr(""),
			}))

			var unstructured map[string]any
			assert.NoError(t, yaml.Unmarshal([]byte(c.yaml), &unstructured))

			resp, _, err := client.StartPipeline(ctx, StartPipelineOptions{
				UnstructuredDeployment: unstructured,
				Headers:                c.inputHeaders,
			})

			assert.NoError(t, err)
			assert.Equal(t, "1-800-pipelines", resp.PipelineID)
		})
	}

}
