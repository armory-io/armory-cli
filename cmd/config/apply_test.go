package config

import (
	"bytes"
	cliconfig "github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/armory/armory-cli/pkg/model/configClient"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/jarcoal/httpmock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func TestConfigApplyTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigApplyTestSuite))
}

type ConfigApplyTestSuite struct {
	suite.Suite
}

func (suite *ConfigApplyTestSuite) SetupSuite() {
	os.Setenv("ARMORY_CLI_TEST", "true")
	httpmock.Activate()
}

func (suite *ConfigApplyTestSuite) SetupTest() {
	httpmock.Reset()
}

func (suite *ConfigApplyTestSuite) TearDownSuite() {
	os.Unsetenv("ARMORY_CLI_TEST")
	httpmock.DeactivateAndReset()
}

func (suite *ConfigApplyTestSuite) TestConfigApplyCreateRole() {
	getExpected := []model.RoleConfig{}
	getEnvironmentsExpected := []configClient.Environment{{
		Name: "testTenant",
	}}
	postExpected := model.RoleConfig{
		Name:   "test",
		Tenant: "testTenant",
		Grants: []model.GrantConfig{{
			Type:       "api",
			Resource:   "org",
			Permission: "all",
		},
		}}
	if err := registerResponder(getExpected, http.StatusOK, "/roles", http.MethodGet); err != nil {
		suite.T().Fatal(err)
	}
	if err := registerResponder(getEnvironmentsExpected, http.StatusOK, "/environments", http.MethodGet); err != nil {
		suite.T().Fatal(err)
	}
	if err := registerResponder(postExpected, http.StatusCreated, "/roles", http.MethodPost); err != nil {
		suite.T().Fatal(err)
	}

	tempFile := util.TempAppFile("", "app", testConfigYamlStrForCreate)
	if tempFile == nil {
		suite.T().Fatal("TestDeployStartJsonSuccess failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	cmd := getConfigApplyCmdWithTmpFile(outWriter, tempFile, "json")
	if err := cmd.Execute(); err != nil {
		suite.T().Fatal(err)
	}
	if _, err := ioutil.ReadAll(outWriter); err != nil {
		suite.T().Fatal(err)
	}
	callCount := httpmock.GetCallCountInfo()
	suite.Equal(1, callCount["POST /roles"])
	suite.Equal(1, callCount["GET /roles"])
	suite.Equal(1, callCount["GET /environments"])
}

func (suite *ConfigApplyTestSuite) TestConfigApplyUpdateRole() {
	getExpected := []model.RoleConfig{{
		ID:     "test-role-id",
		Name:   "test",
		Tenant: "testTenant",
		EnvID:  "env-id",
		Grants: []model.GrantConfig{{
			Type:       "api",
			Resource:   "organization",
			Permission: "full",
		}},
	}}
	getEnvironmentsExpected := []configClient.Environment{{
		Name: "testTenant",
		ID:   "env-id",
	}}
	putExpected := model.RoleConfig{
		Name:   "test",
		Tenant: "testTenant",
		Grants: []model.GrantConfig{{
			Type:       "api",
			Resource:   "tenant",
			Permission: "full",
		},
		}}
	if err := registerResponder(getExpected, http.StatusOK, "/roles", http.MethodGet); err != nil {
		suite.T().Fatal(err)
	}
	if err := registerResponder(getEnvironmentsExpected, http.StatusOK, "/environments", http.MethodGet); err != nil {
		suite.T().Fatal(err)
	}
	if err := registerResponder(putExpected, http.StatusOK, "/roles/test-role-id", http.MethodPut); err != nil {
		suite.T().Fatal(err)
	}

	tempFile := util.TempAppFile("", "app", testConfigYamlStrForUpdate)
	if tempFile == nil {
		suite.T().Fatal("TestDeployStartJsonSuccess failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	cmd := getConfigApplyCmdWithTmpFile(outWriter, tempFile, "json")
	if err := cmd.Execute(); err != nil {
		suite.T().Fatal(err)
	}
	if _, err := ioutil.ReadAll(outWriter); err != nil {
		suite.T().Fatal(err)
	}
	callCount := httpmock.GetCallCountInfo()
	suite.Equal(1, callCount["PUT /roles/test-role-id"])
	suite.Equal(1, callCount["GET /roles"])
	suite.Equal(1, callCount["GET /environments"])
}

func (suite *ConfigApplyTestSuite) TestConfigApplyUpdateOfSystemRoleIsBlocked() {
	getExpected := []model.RoleConfig{{
		ID:            "system-role-id",
		Name:          "test",
		EnvID:         "env-id",
		Tenant:        "testTenant",
		SystemDefined: true,
		Grants: []model.GrantConfig{{
			Type:       "api",
			Resource:   "org",
			Permission: "all",
		}},
	}}
	getEnvironmentsExpected := []configClient.Environment{{
		Name: "testTenant",
		ID:   "env-id",
	}}

	if err := registerResponder(getExpected, http.StatusOK, "/roles", http.MethodGet); err != nil {
		suite.T().Fatal(err)
	}
	if err := registerResponder(getEnvironmentsExpected, http.StatusOK, "/environments", http.MethodGet); err != nil {
		suite.T().Fatal(err)
	}

	tempFile := util.TempAppFile("", "app", testConfigYamlStrForUpdate)
	if tempFile == nil {
		suite.T().Fatal("TestDeployStartJsonSuccess failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	cmd := getConfigApplyCmdWithTmpFile(outWriter, tempFile, "json")
	if err := cmd.Execute(); err != nil {
		suite.T().Fatal(err)
	}
	if _, err := ioutil.ReadAll(outWriter); err != nil {
		suite.T().Fatal(err)
	}
	callCount := httpmock.GetCallCountInfo()
	suite.Equal(0, callCount["PUT /roles/system-role-id"])
	suite.Equal(1, callCount["GET /roles"])
	suite.Equal(1, callCount["GET /environments"])
}

func (suite *ConfigApplyTestSuite) TestConfigApplyDeleteOfSystemRoleIsBlocked() {
	getExpected := []model.RoleConfig{{
		ID:            "system-role-id",
		Name:          "test",
		EnvID:         "env-id",
		Tenant:        "testTenant",
		SystemDefined: true,
		Grants: []model.GrantConfig{{
			Type:       "api",
			Resource:   "org",
			Permission: "all",
		}},
	}}
	getEnvironmentsExpected := []configClient.Environment{{
		Name: "testTenant",
		ID:   "env-id",
	}}

	if err := registerResponder(getExpected, http.StatusOK, "/roles", http.MethodGet); err != nil {
		suite.T().Fatal(err)
	}
	if err := registerResponder(getEnvironmentsExpected, http.StatusOK, "/environments", http.MethodGet); err != nil {
		suite.T().Fatal(err)
	}

	tempFile := util.TempAppFile("", "app", testConfigYamlStrForDeleteSystemRoles)
	if tempFile == nil {
		suite.T().Fatal("TestDeployStartJsonSuccess failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	cmd := getConfigApplyCmdWithTmpFile(outWriter, tempFile, "json")
	if err := cmd.Execute(); err != nil {
		suite.T().Fatal(err)
	}
	if _, err := ioutil.ReadAll(outWriter); err != nil {
		suite.T().Fatal(err)
	}
	callCount := httpmock.GetCallCountInfo()
	suite.Equal(0, callCount["DELETE /roles/system-role-id"])
	suite.Equal(1, callCount["GET /roles"])
	suite.Equal(1, callCount["GET /environments"])
}

func (suite *ConfigApplyTestSuite) TestConfigApplyDeleteRoleAllowAutoDelete() {
	getExpected := []model.RoleConfig{{
		ID:     "role-id-1",
		Name:   "test",
		EnvID:  "env-id",
		Tenant: "testTenant",
		Grants: []model.GrantConfig{{
			Type:       "api",
			Resource:   "org",
			Permission: "all",
		}},
	},
		{
			ID:     "role-id-2",
			Name:   "test2",
			EnvID:  "env-id",
			Tenant: "testTenant",
			Grants: []model.GrantConfig{{
				Type:       "api",
				Resource:   "org",
				Permission: "all",
			}},
		},
	}
	getEnvironmentsExpected := []configClient.Environment{{
		Name: "testTenant",
		ID:   "env-id",
	}}
	deleteExpected := model.RoleConfig{
		Name:   "test2",
		Tenant: "testTenant",
		Grants: []model.GrantConfig{{
			Type:       "api",
			Resource:   "org",
			Permission: "all",
		},
		}}
	putExpected := model.RoleConfig{
		Name:   "test",
		Tenant: "testTenant",
		Grants: []model.GrantConfig{{
			Type:       "api",
			Resource:   "tenant",
			Permission: "all",
		},
		}}
	if err := registerResponder(getExpected, http.StatusOK, "/roles", http.MethodGet); err != nil {
		suite.T().Fatal(err)
	}
	if err := registerResponder(getEnvironmentsExpected, http.StatusOK, "/environments", http.MethodGet); err != nil {
		suite.T().Fatal(err)
	}
	if err := registerResponder(putExpected, http.StatusOK, "/roles/role-id-1", http.MethodPut); err != nil {
		suite.T().Fatal(err)
	}
	if err := registerResponder(deleteExpected, http.StatusNoContent, "/roles/role-id-2", http.MethodDelete); err != nil {
		suite.T().Fatal(err)
	}

	tempFile := util.TempAppFile("", "app", testConfigYamlStrForDeleteAllowAutoDelete)
	if tempFile == nil {
		suite.T().Fatal("TestDeployStartJsonSuccess failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	cmd := getConfigApplyCmdWithTmpFile(outWriter, tempFile, "json")
	if err := cmd.Execute(); err != nil {
		suite.T().Fatal(err)
	}
	if _, err := ioutil.ReadAll(outWriter); err != nil {
		suite.T().Fatal(err)
	}
	callCount := httpmock.GetCallCountInfo()
	suite.Equal(1, callCount["DELETE /roles/role-id-2"])
	suite.Equal(1, callCount["GET /roles"])
	suite.Equal(1, callCount["GET /environments"])
	suite.Equal(1, callCount["PUT /roles/role-id-1"])
}

func (suite *ConfigApplyTestSuite) TestConfigApplyDeleteRoleDontAllowAutoDelete() {
	getExpected := []model.RoleConfig{
		{
			ID:     "role-id-1",
			EnvID:  "env-id",
			Name:   "test",
			Tenant: "testTenant",
			Grants: []model.GrantConfig{
				{
					Type:       "api",
					Resource:   "org",
					Permission: "all",
				},
			},
		},
		{
			ID:     "role-id-2",
			EnvID:  "env-id",
			Name:   "test2",
			Tenant: "testTenant",
			Grants: []model.GrantConfig{
				{
					Type:       "api",
					Resource:   "org",
					Permission: "all",
				},
			},
		},
	}
	getEnvironmentsExpected := []configClient.Environment{{
		Name: "testTenant",
		ID:   "env-id",
	}}
	putExpected := model.RoleConfig{
		Name:   "test",
		Tenant: "testTenant",
		Grants: []model.GrantConfig{{
			Type:       "api",
			Resource:   "tenant",
			Permission: "all",
		},
		}}
	if err := registerResponder(getExpected, http.StatusOK, "/roles", http.MethodGet); err != nil {
		suite.T().Fatal(err)
	}
	if err := registerResponder(getEnvironmentsExpected, http.StatusOK, "/environments", http.MethodGet); err != nil {
		suite.T().Fatal(err)
	}
	if err := registerResponder(putExpected, http.StatusOK, "/roles/role-id-1", http.MethodPut); err != nil {
		suite.T().Fatal(err)
	}

	tempFile := util.TempAppFile("", "app", testConfigYamlStrForDeleteDontAllowAutoDelete)
	if tempFile == nil {
		suite.T().Fatal("TestDeployStartJsonSuccess failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	cmd := getConfigApplyCmdWithTmpFile(outWriter, tempFile, "json")
	if err := cmd.Execute(); err != nil {
		suite.T().Fatal(err)
	}
	if _, err := ioutil.ReadAll(outWriter); err != nil {
		suite.T().Fatal(err)
	}
	callCount := httpmock.GetCallCountInfo()
	suite.Equal(0, callCount["DELETE /roles/role-id-2"])
	suite.Equal(1, callCount["GET /roles"])
	suite.Equal(1, callCount["GET /environments"])
	suite.Equal(1, callCount["PUT /roles/role-id-1"])
}

func registerResponder(body any, status int, url, method string) error {
	responder, err := httpmock.NewJsonResponder(status, body)
	if err != nil {
		return err
	}
	httpmock.RegisterResponder(method, url, responder)
	return nil
}

func getConfigApplyCmdWithTmpFile(outWriter io.Writer, tmpFile *os.File, output string) *cobra.Command {
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
	configApplyCmd := NewConfigApplyCmd(configuration)
	configApplyCmd.SetOut(outWriter)
	args := []string{
		"apply",
		"--file=" + tmpFile.Name(),
	}
	configApplyCmd.SetArgs(args)
	return configApplyCmd
}

const testConfigYamlStrForUpdate = `
roles:
  - name: test
    tenant: testTenant
    grants:
      - type: api
        resource: tenant
        permission: full
`
const testConfigYamlStrForCreate = `
roles:
  - name: test
    tenant: testTenant
    grants:
      - type: api
        resource: org
        permission: all
`

const testConfigYamlStrForDeleteAllowAutoDelete = `
allowAutoDelete: true
roles:
  - name: test
    tenant: testTenant
    grants:
      - type: api
        resource: tenant
        permission: all
`
const testConfigYamlStrForDeleteDontAllowAutoDelete = `
allowAutoDelete: false
roles:
  - name: test
    tenant: testTenant
    grants:
      - type: api
        resource: tenant
        permission: all
`

const testConfigYamlStrForDeleteSystemRoles = `
roles: []
`
