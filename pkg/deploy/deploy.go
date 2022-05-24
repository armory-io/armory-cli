package deploy

import (
	"context"
	"fmt"
	deploy "github.com/armory-io/deploy-engine/pkg"
	"github.com/armory/armory-cli/cmd/version"
	"net/url"
	"os"
)

type Client struct {
	*deploy.APIClient
	Context context.Context
}

var source = "armory-cli"

func NewDeployClient(armoryCloudAddr *url.URL, token string) (*Client, error) {
	if val, present := os.LookupEnv("ARMORY_DEPLOYORIGIN"); present {
		source = val
	}

	deployClient := &Client{
		Context: context.Background(),
	}
	cfg := deploy.NewConfiguration()
	cfg.Host = armoryCloudAddr.Host
	cfg.Scheme = armoryCloudAddr.Scheme
	cfg.AddDefaultHeader("X-Armory-Client", fmt.Sprintf("%s", source)+"/"+version.Version)
	deployClient.APIClient = deploy.NewAPIClient(cfg)
	deployClient.Context = context.WithValue(deployClient.Context, deploy.ContextAccessToken, token)
	return deployClient, nil
}
