package config

import (
	"net/url"
	"strings"
	"time"

	"github.com/armory/armory-cli/pkg/armoryCloud"
	"github.com/armory/armory-cli/pkg/auth"
	"github.com/armory/armory-cli/pkg/errors"
	"github.com/armory/armory-cli/pkg/output"
	log "github.com/sirupsen/logrus"
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

	// clock is a function that returns the current timestamp. It should only be used in tests to make
	// log timestamps deterministic.
	clock func() time.Time
}

func New(input *Input) *Configuration {
	c := &Configuration{
		input: input,
	}
	if input.IsTest != nil && *input.IsTest {
		c.clock = func() time.Time { return time.Time{} }
	}
	return c
}

// Input
// Everything in this struct should be a pointer and lazily evaluated because cobra will set the value eventually
type Input struct {
	AccessToken  *string
	ApiAddr      *string
	ClientId     *string
	ClientSecret *string
	OutFormat    *string
	IsTest       *bool
}

type ArmoryCloudEnv int64
type ArmoryApplicationEnvironment string

const (
	dev ArmoryCloudEnv = iota
	staging
	prod
	envDev     ArmoryApplicationEnvironment = "dev"
	envStaging ArmoryApplicationEnvironment = "staging"
	envProd    ArmoryApplicationEnvironment = "prod"
)

func (c *Configuration) GetArmoryCloudEnv() ArmoryCloudEnv {
	addr := c.GetArmoryCloudAddr()
	var authTenant ArmoryCloudEnv
	switch addr.Host {
	case "api.cloud.armory.io":
		authTenant = prod
	case "api.staging.cloud.armory.io":
		authTenant = staging
	default:
		authTenant = dev
	}
	return authTenant
}

func (c *Configuration) GetAuth() *auth.Auth {
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
	token, err := c.GetAuth().GetToken()
	if err != nil {
		log.Fatalf("failed to fetch access token, err: %s", err.Error())
	}
	return token
}

func (c *Configuration) GetCustomerEnvironmentId() string {
	environment, err := c.GetAuth().GetEnvironmentId()
	if err != nil {
		log.Fatalf("failed to fetch environment, err: %s", err.Error())
	}
	return environment
}

func (c *Configuration) GetCustomerOrganizationId() string {
	organization, err := c.GetAuth().GetOrganizationId()
	if err != nil {
		log.Fatalf("failed to fetch organization, err: %s", err.Error())
	}
	return organization
}

func (c *Configuration) GetArmoryCloudAddr() *url.URL {
	parsedAddr, err := c.getArmoryCloudAddr()
	if err != nil {
		log.Fatalf(err.Error())
	}
	return parsedAddr
}

func (c *Configuration) GetArmoryCloudGraphQLAddr() *url.URL {
	addr, err := c.getArmoryCloudAddr()
	if err != nil {
		log.Fatal(err)
	}
	if strings.HasPrefix(addr.Host, "localhost:") {
		addr.Scheme = "http"
		addr.Host = "localhost:8081"
	}
	addr.Path = "/v1/graphql"
	return addr
}

func (c *Configuration) Now() time.Time {
	if c.clock == nil {
		return time.Now()
	}
	return c.clock()
}

func (c *Configuration) getArmoryCloudAddr() (*url.URL, error) {
	armoryCloudAddr := *c.input.ApiAddr
	parsedUrl, err := url.Parse(armoryCloudAddr)
	if err != nil {
		return nil, errors.NewWrappedErrorWithDynamicContext(ErrInvalidArmoryCloudAddr, err, ", provided addr: "+armoryCloudAddr)
	}

	if parsedUrl.Scheme == "" {
		return nil, errors.NewErrorWithDynamicContext(ErrInvalidUrlScheme, ", provided addr: "+armoryCloudAddr)
	}

	if parsedUrl.Host == "" {
		return nil, errors.NewErrorWithDynamicContext(ErrMissingHostInUrl, ", provided addr: "+armoryCloudAddr)
	}

	if strings.TrimSuffix(parsedUrl.Path, "/") != "" {
		return nil, errors.NewErrorWithDynamicContext(ErrIncludedPathInUrl, ", provided addr: "+armoryCloudAddr)
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
	case "yaml":
		oType = output.Yaml
	case "json":
		oType = output.Json
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
	CloudConsoleBaseUrl    string
	CliClientId            string
	TokenIssuerUrl         string
	Audience               string
	AWSAccountID           string
	ApplicationEnvironment ArmoryApplicationEnvironment
}

func (c *Configuration) GetArmoryCloudEnvironmentConfiguration() *ArmoryCloudEnvironmentConfiguration {
	var armoryCloudEnvironmentConfiguration *ArmoryCloudEnvironmentConfiguration
	switch c.GetArmoryCloudEnv() {
	case prod:
		armoryCloudEnvironmentConfiguration = &ArmoryCloudEnvironmentConfiguration{
			CloudConsoleBaseUrl:    "https://console.cloud.armory.io",
			CliClientId:            "GjHFCN83nbHZaUT4CR4mQ65QYk8uUAKy",
			TokenIssuerUrl:         "https://auth.cloud.armory.io/oauth",
			Audience:               "https://api.cloud.armory.io",
			ApplicationEnvironment: envProd,
			AWSAccountID:           "961214755549",
		}
	case staging:
		armoryCloudEnvironmentConfiguration = &ArmoryCloudEnvironmentConfiguration{
			CloudConsoleBaseUrl:    "https://console.staging.cloud.armory.io",
			CliClientId:            "sjkd8ufTR3AxHHZz8XZLE0Y8UAIjTM1I",
			TokenIssuerUrl:         "https://auth.staging.cloud.armory.io/oauth",
			Audience:               "https://api.staging.cloud.armory.io",
			ApplicationEnvironment: envStaging,
			AWSAccountID:           "200597635891",
		}
	case dev:
		armoryCloudEnvironmentConfiguration = &ArmoryCloudEnvironmentConfiguration{
			CloudConsoleBaseUrl:    "https://console.dev.cloud.armory.io:3000",
			CliClientId:            "o2QghLMwgT1t1glzGaAOqEiIbbiHqUpc",
			TokenIssuerUrl:         "https://auth.dev.cloud.armory.io/oauth",
			Audience:               "https://api.dev.cloud.armory.io",
			ApplicationEnvironment: envDev,
		}
	default:
		log.Fatalf("Cannot fetch armory cloud config for unknown armory env")
	}
	return armoryCloudEnvironmentConfiguration
}

func (c *Configuration) GetIsTest() *bool {
	return c.input.IsTest
}
