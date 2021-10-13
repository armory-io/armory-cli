package deploy

import (
	"context"
	"fmt"
	deploy "github.com/armory-io/deploy-engine/deploy/client"
)

type Client struct {
	*deploy.APIClient
	Context context.Context
}
const UserAgent = "Armory-CLI"

func NewDeployClient(basePath, token string) (*Client, error){
	deployClient:= &Client{
		Context: context.Background(),
	}
	cfg := deploy.NewConfiguration()
	cfg.Host = basePath
	cfg.Scheme = "https"
	cfg.UserAgent = fmt.Sprintf("%s",UserAgent)
	deployClient.APIClient = deploy.NewAPIClient(cfg)
	deployClient.Context = context.WithValue(deployClient.Context, deploy.ContextAccessToken, token)
	return deployClient, nil
}