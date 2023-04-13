package armoryCloud

import (
	"context"
	"fmt"
	"github.com/armory/armory-cli/cmd/version"
	"github.com/hashicorp/go-retryablehttp"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type (
	Client struct {
		Context       context.Context
		configuration *Configuration
		Http          *http.Client
		RetryableHttp *http.Client
	}

	Configuration struct {
		Host           string
		Scheme         string
		DefaultHeaders map[string]string
		UserAgent      string
	}
)

var source = "armory-cli"

func NewArmoryCloudClient(armoryCloudAddr *url.URL, token string) (*Client, error) {
	if val, present := os.LookupEnv("ARMORY_DEPLOYORIGIN"); present {
		source = val
	}

	productVersion := fmt.Sprintf("%s/%s", source, version.Version)

	client := &retryablehttp.Client{
		HTTPClient:   http.DefaultClient, // tests fail if I use the cleanhttp.DefaultPooledClient() from retryablehttp
		RetryWaitMin: 500 * time.Millisecond,
		RetryWaitMax: 2 * time.Second,
		RetryMax:     5,
	}
	retryClient := client.StandardClient()
	// if we are in a test don't use the retry client for any network executions; this messes with error handling
	if _, isATest := os.LookupEnv("ARMORY_CLI_TEST"); isATest {
		retryClient = client.HTTPClient
	}

	return &Client{
		Context:       context.Background(),
		RetryableHttp: retryClient,
		Http:          client.HTTPClient,
		configuration: &Configuration{
			Host:   armoryCloudAddr.Host,
			Scheme: armoryCloudAddr.Scheme,
			DefaultHeaders: map[string]string{
				"Authorization":   fmt.Sprintf("Bearer %s", token),
				"Content-Type":    "application/json",
				"X-Armory-Client": productVersion,
				"User-Agent":      productVersion,
			},
			UserAgent: productVersion,
		},
	}, nil
}

func (c *Client) SimpleRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Request, error) {
	return c.Request(
		ctx,
		WithMethod(method),
		WithPath(path),
		WithBody(body),
	)
}

func (c *Client) Request(ctx context.Context, opts ...RequestOption) (*http.Request, error) {
	var builder requestBuilder
	for _, opt := range opts {
		opt(&builder)
	}

	u := &url.URL{
		Scheme: c.configuration.Scheme,
		Host:   c.configuration.Host,
		Path:   builder.path,
	}

	request, err := http.NewRequestWithContext(ctx, builder.method, u.String(), builder.body)
	if err != nil {
		return nil, err
	}

	for key, value := range c.configuration.DefaultHeaders {
		request.Header.Add(key, value)
	}

	for key, value := range builder.headers {
		request.Header.Set(key, value)
	}

	return request, nil
}

type requestBuilder struct {
	method  string
	path    string
	headers map[string]string
	body    io.Reader
}

type RequestOption = func(builder *requestBuilder)

func WithMethod(method string) RequestOption {
	return func(builder *requestBuilder) {
		builder.method = method
	}
}

func WithPath(method string) RequestOption {
	return func(builder *requestBuilder) {
		builder.path = method
	}
}

func WithHeader(key, value string) RequestOption {
	return func(builder *requestBuilder) {
		if builder.headers == nil {
			builder.headers = make(map[string]string)
		}
		builder.headers[key] = value
	}
}

func WithBody(body io.Reader) RequestOption {
	return func(builder *requestBuilder) {
		builder.body = body
	}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.Http.Do(req)
}

func (c *Client) DoWithRetry(req *http.Request) (*http.Response, error) {
	return c.RetryableHttp.Do(req)
}
