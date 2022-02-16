package deploy

import (
	"context"
	"crypto/tls"
	"fmt"
	deploy "github.com/armory-io/deploy-engine/pkg"
	"github.com/armory/armory-cli/cmd/version"
	"net/http"
	"os"
)

type Client struct {
	*deploy.APIClient
	Context context.Context
}

var source = "armory-cli"

func NewDeployClient(basePath, token string, dev bool) (*Client, error) {
	if val, present := os.LookupEnv("ARMORY_DEPLOYORIGIN"); present {
		source = val
	}

	deployClient := &Client{
		Context: context.Background(),
	}
	cfg := deploy.NewConfiguration()
	cfg.Host = basePath
	cfg.Scheme = "https"
	if dev {
		transport := http.DefaultTransport.(*http.Transport)
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient := http.Client{
			Transport: transport,
		}
		cfg.HTTPClient = &httpClient
	}
	cfg.AddDefaultHeader("X-Armory-Client", fmt.Sprintf("%s", source)+"/"+version.Version)
	deployClient.APIClient = deploy.NewAPIClient(cfg)
	deployClient.Context = context.WithValue(deployClient.Context, deploy.ContextAccessToken, token)
	return deployClient, nil
}
