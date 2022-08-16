package config

import (
	"github.com/armory/armory-cli/pkg/armoryCloud"
	"github.com/armory/armory-cli/pkg/auth"
	"github.com/armory/armory-cli/pkg/output"
	log "github.com/sirupsen/logrus"
	"net/url"
	"strings"
)

type CliConfiguration interface {
	GetArmoryCloudEnv() ArmoryCloudEnv
	GetAuthToken() string
	GetCustomerEnvironmentId() string
	GetArmoryCloudAddr() *url.URL
	GetArmoryCloudEnvironmentConfiguration() *ArmoryCloudEnvironmentConfiguration
}

type Configuration struct {
	input *Input
}

func New(input *Input) *Configuration {
	return &Configuration{input: input}
}

// Input
// Everything in this struct should be a pointer and lazily evaluated because cobra will set the value eventually
type Input struct {
	AccessToken  *string
	ApiAddr      *string
	ClientId     *string
	ClientSecret *string
	OutFormat    *string
}

type ArmoryCloudEnv int64

const (
	dev ArmoryCloudEnv = iota
	staging
	prod
)

func (c *Configuration) GetArmoryCloudEnv() ArmoryCloudEnv {
	addr := c.GetArmoryCloudAddr()
	var authTenant ArmoryCloudEnv
	switch addr.Host {
	case "api.cloud.armory.io":
		authTenant = prod
		break
	case "api.staging.cloud.armory.io":
		authTenant = staging
		break
	default:
		authTenant = dev
		break
	}
	return authTenant
}

func (c *Configuration) getAuth() *auth.Auth {
	conf := c.GetArmoryCloudEnvironmentConfiguration()
	return auth.NewAuth(
		*c.input.ClientId,
		*c.input.ClientSecret,
		"client_credentials",
		conf.TokenIssuerUrl,
		conf.Audience,
		*c.input.AccessToken,
	)
}

func (c *Configuration) GetAuthToken() string {
	token, err := c.getAuth().GetToken()
	if err != nil {
		log.Fatalf("failed to fetch access token, err: %s", err.Error())
	}
	return token
}

func (c *Configuration) GetCustomerEnvironmentId() string {
	environment, err := c.getAuth().GetEnvironmentId()
	if err != nil {
		log.Fatalf("failed to fetch environment, err: %s", err.Error())
	}
	return environment
}

func (c *Configuration) GetCustomerOrganizationId() string {
	organization, err := c.getAuth().GetOrganizationId()
	if err != nil {
		log.Fatalf("failed to fetch organization, err: %s", err.Error())
	}
	return organization
}

func (c *Configuration) GetArmoryCloudAddr() *url.URL {
	parsedAddr, err := c.getArmoryCloudAdder()
	if err != nil {
		log.Fatalf(err.Error())
	}
	return parsedAddr
}

func (c *Configuration) getArmoryCloudAdder() (*url.URL, error) {
	armoryCloudAddr := *c.input.ApiAddr
	parsedUrl, err := url.Parse(armoryCloudAddr)
	if err != nil {
		return nil, newArmoryCloudAddrParsingError(armoryCloudAddr, err)
	}

	if parsedUrl.Scheme == "" {
		return nil, newInvalidUrlSchemeError(armoryCloudAddr)
	}

	if parsedUrl.Host == "" {
		return nil, newMissingHostInUrlError(armoryCloudAddr)
	}

	if strings.TrimSuffix(parsedUrl.Path, "/") != "" {
		return nil, newIncludedPathInUrlError(armoryCloudAddr)
	}

	return &url.URL{
		Scheme: parsedUrl.Scheme,
		Host:   parsedUrl.Host,
	}, nil
}

func (c *Configuration) GetArmoryCloudClient() *armoryCloud.Client {
	armoryCloudClient, err := armoryCloud.NewArmoryCloudClient(
		c.GetArmoryCloudAddr(),
		c.GetAuthToken(),
	)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return armoryCloudClient
}

func (c *Configuration) GetOutputType() output.Type {
	var oType output.Type
	switch strings.ToLower(*c.input.OutFormat) {
	case "plain", "", "text":
		oType = output.Text
		break
	case "yaml":
		oType = output.Yaml
		break
	case "json":
		oType = output.Json
		break
	default:
		log.Fatalf("the output type is invalid. Do not specify parameter to get plain text output. Available options: [json, yaml, text]")
	}
	return oType
}

func (c *Configuration) GetOutputFormatter() output.Formatter {
	return output.GetFormatterForOutputType(c.GetOutputType())
}

func (c *Configuration) SetOutputFormatter(formatter string) {
	c.input.OutFormat = &formatter
}

type ArmoryCloudEnvironmentConfiguration struct {
	CloudConsoleBaseUrl string
	CliClientId         string
	TokenIssuerUrl      string
	Audience            string
}

func (c *Configuration) GetArmoryCloudEnvironmentConfiguration() *ArmoryCloudEnvironmentConfiguration {
	var armoryCloudEnvironmentConfiguration *ArmoryCloudEnvironmentConfiguration
	switch c.GetArmoryCloudEnv() {
	case prod:
		armoryCloudEnvironmentConfiguration = &ArmoryCloudEnvironmentConfiguration{
			CloudConsoleBaseUrl: "https://console.cloud.armory.io",
			CliClientId:         "GjHFCN83nbHZaUT4CR4mQ65QYk8uUAKy",
			TokenIssuerUrl:      "https://auth.cloud.armory.io/oauth",
			Audience:            "https://api.cloud.armory.io",
		}
		break
	case staging:
		armoryCloudEnvironmentConfiguration = &ArmoryCloudEnvironmentConfiguration{
			CloudConsoleBaseUrl: "https://console.staging.cloud.armory.io",
			CliClientId:         "sjkd8ufTR3AxHHZz8XZLE0Y8UAIjTM1I",
			TokenIssuerUrl:      "https://auth.staging.cloud.armory.io/oauth",
			Audience:            "https://api.staging.cloud.armory.io",
		}
		break
	case dev:
		armoryCloudEnvironmentConfiguration = &ArmoryCloudEnvironmentConfiguration{
			CloudConsoleBaseUrl: "https://console.dev.cloud.armory.io:3000",
			CliClientId:         "o2QghLMwgT1t1glzGaAOqEiIbbiHqUpc",
			TokenIssuerUrl:      "https://auth.dev.cloud.armory.io/oauth",
			Audience:            "https://api.dev.cloud.armory.io",
		}
		break
	default:
		log.Fatalf("Cannot fetch armory cloud config for unknown armory env")
	}
	return armoryCloudEnvironmentConfiguration
}
