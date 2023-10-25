package graphql

import (
	"context"
	armoryhttp "github.com/armory-io/go-commons/http/client"
	"github.com/armory-io/go-commons/opentelemetry"
	"github.com/armory/armory-cli/internal/clierr"
	"github.com/armory/armory-cli/internal/clierr/exitcodes"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/google/uuid"
	"github.com/machinebox/graphql"
	"io"
	"net/http"
)

type (
	Client struct {
		client *graphql.Client
	}

	tokenSupplier struct {
		configuration *config.Configuration
	}

	hasuraRoundTripper struct {
		baseRoundTripper http.RoundTripper
	}

	requestIDContextKey struct{}
)

const (
	hasuraRequestIDHeader = "X-Request-ID"
)

func NewClient(configuration *config.Configuration) *Client {
	httpClient := armoryhttp.NewAuthenticatedHTTPClient(&tokenSupplier{configuration: configuration}, opentelemetry.Configuration{})
	httpClient.Transport = &hasuraRoundTripper{baseRoundTripper: httpClient.Transport}

	graphQLEndpoint := configuration.GetArmoryCloudGraphQLAddr()
	graphQLClient := graphql.NewClient(graphQLEndpoint.String(), graphql.WithHTTPClient(httpClient))
	return &Client{client: graphQLClient}
}

func (t *tokenSupplier) GetToken(context.Context) (string, error) {
	return t.configuration.GetAuth().GetToken()
}

func (r *hasuraRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	requestID, ok := request.Context().Value(requestIDContextKey{}).(string)
	if ok {
		request.Header.Set(hasuraRequestIDHeader, requestID)
	}

	response, err := r.baseRoundTripper.RoundTrip(request)
	if err != nil {
		return response, err
	}

	if response.StatusCode != http.StatusOK {
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		return nil, clierr.NewAPIError(
			"API Error",
			response.StatusCode,
			body,
			exitcodes.Error,
		)
	}

	return response, nil
}

func (c *Client) newRequestID() string {
	return uuid.NewString()
}

func (c *Client) doGraphQLRequest(
	ctx context.Context,
	requestID string,
	request *graphql.Request,
	responsePtr any,
) error {
	return c.client.Run(
		context.WithValue(ctx, requestIDContextKey{}, requestID),
		request,
		responsePtr,
	)
}
