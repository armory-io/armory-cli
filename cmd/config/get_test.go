package config

import (
	"bytes"
	"encoding/json"
	cliconfig "github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/armory/armory-cli/pkg/model/configClient"
	"github.com/jarcoal/httpmock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestConfigGetTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigGetTestSuite))
}

type ConfigGetTestSuite struct {
	suite.Suite
}

func (suite *ConfigGetTestSuite) SetupSuite() {
	os.Setenv("ARMORY_CLI_TEST", "true")
	httpmock.Activate()
}

func (suite *ConfigGetTestSuite) SetupTest() {
	httpmock.Reset()
}

func (suite *ConfigGetTestSuite) TearDownSuite() {
	os.Unsetenv("ARMORY_CLI_TEST")
	httpmock.DeactivateAndReset()
}

func (suite *ConfigGetTestSuite) TestConfigGetUserRole() {
	var getEnvironmentsExpected []configClient.Environment
	getExpected := []model.RoleConfig{{
		Name:   "test",
		Tenant: "testTenant",
		Grants: []model.GrantConfig{{
			Type:       "api",
			Resource:   "org",
			Permission: "all",
		}},
	}}

	if err := registerResponder(getEnvironmentsExpected, http.StatusOK, "/environments", http.MethodGet); err != nil {
		suite.T().Fatal(err)
	}

	err := registerResponder(getExpected, http.StatusOK, "/roles", http.MethodGet)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}

	outWriter := bytes.NewBufferString("")
	cmd := getConfigGetCmdWithTmpFile(outWriter, "json")
	err = cmd.Execute()
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	jsonContent, err := io.ReadAll(outWriter)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	callCount := httpmock.GetCallCountInfo()
	suite.Equal(1, callCount["GET /environments"])
	suite.Equal(1, callCount["GET /roles"])
	result := model.ConfigurationConfig{}
	json.Unmarshal(jsonContent, &result)
	if len(result.Roles) != 1 {
		suite.T().Fatalf("expected one user role to be retured!")
	}
}

func (suite *ConfigGetTestSuite) TestConfigGetTenants() {
	getEnvironmentsExpected := []configClient.Environment{{
		Name: "testTenant",
	}}
	var getRolesExpected []model.RoleConfig

	if err := registerResponder(getEnvironmentsExpected, http.StatusOK, "/environments", http.MethodGet); err != nil {
		suite.T().Fatal(err)
	}

	err := registerResponder(getRolesExpected, http.StatusOK, "/roles", http.MethodGet)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}

	outWriter := bytes.NewBufferString("")
	cmd := getConfigGetCmdWithTmpFile(outWriter, "json")
	err = cmd.Execute()
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	jsonContent, err := io.ReadAll(outWriter)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	callCount := httpmock.GetCallCountInfo()
	suite.Equal(1, callCount["GET /environments"])
	suite.Equal(1, callCount["GET /roles"])
	result := model.ConfigurationConfig{}
	json.Unmarshal(jsonContent, &result)
	if len(result.Tenants) != 1 {
		suite.T().Fatalf("expected one tenant to be retured!")
	}
}

func (suite *ConfigGetTestSuite) TestConfigGetSystemRole() {
	getEnvironmentsExpected := []configClient.Environment{{
		Name: "testTenant",
	}}
	getExpected := []model.RoleConfig{{
		Name:          "test",
		Tenant:        "testTenant",
		SystemDefined: true,
		Grants: []model.GrantConfig{{
			Type:       "api",
			Resource:   "org",
			Permission: "all",
		}},
	}}

	if err := registerResponder(getEnvironmentsExpected, http.StatusOK, "/environments", http.MethodGet); err != nil {
		suite.T().Fatal(err)
	}
	err := registerResponder(getExpected, http.StatusOK, "/roles", http.MethodGet)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}

	outWriter := bytes.NewBufferString("")
	cmd := getConfigGetCmdWithTmpFile(outWriter, "json")
	err = cmd.Execute()
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	jsonContent, err := io.ReadAll(outWriter)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	callCount := httpmock.GetCallCountInfo()
	suite.Equal(1, callCount["GET /environments"])
	suite.Equal(1, callCount["GET /roles"])
	result := model.ConfigurationConfig{}
	json.Unmarshal(jsonContent, &result)
	if len(result.Roles) != 0 {
		suite.T().Fatalf("expected one user role to be retured!")
	}
}

func getConfigGetCmdWithTmpFile(outWriter io.Writer, output string) *cobra.Command {
	token := "some-token"
	addr := "https://localhost"
	clientId := ""
	clientSecret := ""
	configuration := cliconfig.New(&cliconfig.Input{
		AccessToken:  &token,
		ApiAddr:      &addr,
		ClientId:     &clientId,
		ClientSecret: &clientSecret,
		OutFormat:    &output,
	})
	configApplyCmd := NewConfigGetCmd(configuration)
	configApplyCmd.SetOut(outWriter)
	args := []string{
		"get",
	}
	configApplyCmd.SetArgs(args)
	return configApplyCmd
}

const testGetUserRole = `
roles:
  - name: test
    tenant: testTenantNew
    grants:
      - type: api
        resource: org
        permission: all
`

const testGetSystemRole = `
roles:
  - name: test
    tenant: testTenantNew
    grants:
      - type: api
        resource: org
        permission: all
`
