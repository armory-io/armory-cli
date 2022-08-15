package config

import (
	"bytes"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/model"
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
	postExpected := model.RoleConfig{
		Name:   "test",
		Tenant: "testTenant",
		Grants: []model.GrantConfig{{
			Type:       "api",
			Resource:   "org",
			Permission: "all",
		},
		}}
	err := registerResponder(getExpected, http.StatusOK, "/roles", http.MethodGet)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	err = registerResponder(postExpected, http.StatusCreated, "/roles", http.MethodPost)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}

	tempFile := util.TempAppFile("", "app", testConfigYamlStrForCreate)
	if tempFile == nil {
		suite.T().Fatal("TestDeployStartJsonSuccess failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	cmd := getConfigApplyCmdWithTmpFile(outWriter, tempFile, "json")
	err = cmd.Execute()
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	_, err = ioutil.ReadAll(outWriter)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	callCount := httpmock.GetCallCountInfo()
	suite.Equal(1, callCount["POST /roles"])
	suite.Equal(1, callCount["GET /roles"])
}

func (suite *ConfigApplyTestSuite) TestConfigApplyUpdateRole() {
	getExpected := []model.RoleConfig{{
		Name:   "test",
		Tenant: "testTenant",
		Grants: []model.GrantConfig{{
			Type:       "api",
			Resource:   "org",
			Permission: "all",
		}},
	}}
	putExpected := model.RoleConfig{
		Name:   "test",
		Tenant: "testTenantNew",
		Grants: []model.GrantConfig{{
			Type:       "api",
			Resource:   "org",
			Permission: "all",
		},
		}}
	err := registerResponder(getExpected, http.StatusOK, "/roles", http.MethodGet)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	err = registerResponder(putExpected, http.StatusOK, "/roles/test", http.MethodPut)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}

	tempFile := util.TempAppFile("", "app", testConfigYamlStrForUpdate)
	if tempFile == nil {
		suite.T().Fatal("TestDeployStartJsonSuccess failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	cmd := getConfigApplyCmdWithTmpFile(outWriter, tempFile, "json")
	err = cmd.Execute()
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	_, err = ioutil.ReadAll(outWriter)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	callCount := httpmock.GetCallCountInfo()
	suite.Equal(1, callCount["PUT /roles/test"])
	suite.Equal(1, callCount["GET /roles"])
}

func (suite *ConfigApplyTestSuite) TestConfigApplyDeleteRoleAllowAutoDelete() {
	getExpected := []model.RoleConfig{{
		Name:   "test",
		Tenant: "testTenant",
		Grants: []model.GrantConfig{{
			Type:       "api",
			Resource:   "org",
			Permission: "all",
		}},
	},
		{
			Name:   "test2",
			Tenant: "testTenant",
			Grants: []model.GrantConfig{{
				Type:       "api",
				Resource:   "org",
				Permission: "all",
			}},
		},
	}
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
		Tenant: "testTenantNew",
		Grants: []model.GrantConfig{{
			Type:       "api",
			Resource:   "org",
			Permission: "all",
		},
		}}
	err := registerResponder(getExpected, http.StatusOK, "/roles", http.MethodGet)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	err = registerResponder(putExpected, http.StatusOK, "/roles/test", http.MethodPut)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	err = registerResponder(deleteExpected, http.StatusNoContent, "/roles/test2", http.MethodDelete)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}

	tempFile := util.TempAppFile("", "app", testConfigYamlStrForDeleteAllowAutoDelete)
	if tempFile == nil {
		suite.T().Fatal("TestDeployStartJsonSuccess failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	cmd := getConfigApplyCmdWithTmpFile(outWriter, tempFile, "json")
	err = cmd.Execute()
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	_, err = ioutil.ReadAll(outWriter)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	callCount := httpmock.GetCallCountInfo()
	suite.Equal(1, callCount["DELETE /roles/test2"])
	suite.Equal(1, callCount["GET /roles"])
	suite.Equal(1, callCount["PUT /roles/test"])
}

func (suite *ConfigApplyTestSuite) TestConfigApplyDeleteRoleDontAllowAutoDelete() {
	getExpected := []model.RoleConfig{
		{
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
	putExpected := model.RoleConfig{
		Name:   "test",
		Tenant: "testTenantNew",
		Grants: []model.GrantConfig{{
			Type:       "api",
			Resource:   "org",
			Permission: "all",
		},
		}}
	err := registerResponder(getExpected, http.StatusOK, "/roles", http.MethodGet)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	err = registerResponder(putExpected, http.StatusOK, "/roles/test", http.MethodPut)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}

	tempFile := util.TempAppFile("", "app", testConfigYamlStrForDeleteDontAllowAutoDelete)
	if tempFile == nil {
		suite.T().Fatal("TestDeployStartJsonSuccess failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	cmd := getConfigApplyCmdWithTmpFile(outWriter, tempFile, "json")
	err = cmd.Execute()
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	_, err = ioutil.ReadAll(outWriter)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	callCount := httpmock.GetCallCountInfo()
	suite.Equal(0, callCount["DELETE /roles/test2"])
	suite.Equal(1, callCount["GET /roles"])
	suite.Equal(1, callCount["PUT /roles/test"])
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
	configuration := config.New(&config.Input{
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
    tenant: testTenantNew
    grants:
      - type: api
        resource: org
        permission: all
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
    tenant: testTenantNew
    grants:
      - type: api
        resource: org
        permission: all
`
const testConfigYamlStrForDeleteDontAllowAutoDelete = `
allowAutoDelete: false
roles:
  - name: test
    tenant: testTenantNew
    grants:
      - type: api
        resource: org
        permission: all
`
