package armoryCloud

import (
	"context"
	"fmt"
	"github.com/armory/armory-cli/cmd/version"
	"io"
	"net/http"
	"net/url"
	"os"
)

type (
	Client struct {
		Context       context.Context
		configuration *Configuration
		Http          *http.Client
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

	return &Client{
		Context: context.Background(),
		Http:    http.DefaultClient,
		configuration: &Configuration{
			Host:   armoryCloudAddr.Host,
			Scheme: armoryCloudAddr.Scheme,
			DefaultHeaders: map[string]string{
				"Authorization":   fmt.Sprintf("Bearer %s", token),
				"Content-Type":    "application/json",
				"X-Armory-Client": productVersion,
			},
			UserAgent: productVersion,
		},
	}, nil
}

func (c *Client) Request(ctx context.Context, method string, path string, body io.Reader) (*http.Request, error) {
	u := &url.URL{
		Scheme: c.configuration.Scheme,
		Host:   c.configuration.Host,
		Path:   path,
	}

	request, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}

	request.Header.Add("User-Agent", c.configuration.UserAgent)
	for key, value := range c.configuration.DefaultHeaders {
		request.Header.Add(key, value)
	}

	return request, nil
}
