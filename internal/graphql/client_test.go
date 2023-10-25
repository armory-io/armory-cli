package graphql

import (
	"context"
	"encoding/json"
	"github.com/armory-io/go-commons/server/serr"
	"github.com/armory/armory-cli/internal/clierr"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/machinebox/graphql"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientHeaderPropagation(t *testing.T) {
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// Client propagates authorization token and Hasura request ID to server.
		assert.Equal(t, "Bearer access-token", request.Header.Get("Authorization"))
		assert.Equal(t, "request-id", request.Header.Get("X-Request-ID"))

		assert.NoError(t, json.NewEncoder(writer).Encode(map[string]any{
			"data": map[string]any{},
		}))

		writer.WriteHeader(200)
	}))

	client := NewClient(config.New(&config.Input{
		AccessToken:  lo.ToPtr("access-token"),
		ApiAddr:      lo.ToPtr(server.URL),
		ClientId:     lo.ToPtr(""),
		ClientSecret: lo.ToPtr(""),
	}))

	err := client.doGraphQLRequest(ctx, "request-id", &graphql.Request{}, nil)
	assert.NoError(t, err)
}

func TestAPIErrors(t *testing.T) {
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusUnauthorized)

		assert.NoError(t, json.NewEncoder(writer).Encode(serr.ResponseContract{
			ErrorId: "error-id",
			Errors: []serr.ResponseContractErrorDTO{{
				Message: "not authorized",
			}},
		}))
	}))

	client := NewClient(config.New(&config.Input{
		AccessToken:  lo.ToPtr("access-token"),
		ApiAddr:      lo.ToPtr(server.URL),
		ClientId:     lo.ToPtr(""),
		ClientSecret: lo.ToPtr(""),
	}))

	err := client.doGraphQLRequest(ctx, "request-id", &graphql.Request{}, nil)
	var apiError *clierr.APIError
	assert.ErrorAs(t, err, &apiError)
	assert.ErrorContains(t, apiError, "401")
}

func TestGraphQLErrors(t *testing.T) {
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.NoError(t, json.NewEncoder(writer).Encode(map[string][]map[string]string{
			"errors": {
				{
					"message": "you can't GraphQL here!",
				},
			},
		}))
	}))

	client := NewClient(config.New(&config.Input{
		AccessToken:  lo.ToPtr("access-token"),
		ApiAddr:      lo.ToPtr(server.URL),
		ClientId:     lo.ToPtr(""),
		ClientSecret: lo.ToPtr(""),
	}))

	err := client.doGraphQLRequest(ctx, "request-id", &graphql.Request{}, nil)
	assert.ErrorContains(t, err, "you can't GraphQL here!")
}
