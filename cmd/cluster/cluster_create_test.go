package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/armory/armory-cli/cmd/agent"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestClusterCreateSuite(t *testing.T) {
	suite.Run(t, new(ClusterCreateTestSuite))
}

type ClusterCreateTestSuite struct {
	suite.Suite
}

func (suite *ClusterCreateTestSuite) SetupSuite() {
	assert.NoError(suite.T(), os.Setenv("ARMORY_CLI_TEST", "true"))
	httpmock.Activate()
}

func (suite *ClusterCreateTestSuite) SetupTest() {
	httpmock.Reset()
}

func (suite *ClusterCreateTestSuite) TearDownSuite() {
	assert.NoError(suite.T(), os.Unsetenv("ARMORY_CLI_TEST"))
	httpmock.DeactivateAndReset()
}

func (suite *ClusterCreateTestSuite) TestCreateAndRunCommand() {
	credential := model.Credential{
		ID:           "my-agent-identifier",
		ClientSecret: "my-secret",
		ClientId:     "my-id",
	}

	assert.NoError(suite.T(), registerResponder(credential, http.StatusCreated, "/credentials", http.MethodPost))
	assert.NoError(suite.T(), registerResponder(rolesFromGrantStrings("my-role-id", []string{"api:agentHub:full"}, true), http.StatusOK, "/roles", http.MethodGet))
	assert.NoError(suite.T(), registerResponder([]model.RoleConfig{}, http.StatusOK, "/credentials/my-agent-identifier/roles", http.MethodPut))
	assert.NoError(suite.T(), registerResponder(model.CreateSandboxResponse{ClusterId: "cluster-id"}, 200, "/sandbox/clusters", http.MethodPost))
	assert.NoError(suite.T(), registerResponder(model.SandboxCluster{PercentComplete: 100}, 200, "/sandbox/clusters/cluster-id", http.MethodGet))

	cmd := NewClusterCmd(getDefaultAppConfiguration(), getSandboxFileStore())
	cmd.SetOut(io.Discard)
	cmd.SetArgs([]string{
		"create",
	})

	err := cmd.Execute()
	assert.NoError(suite.T(), err)
}

func (suite *ClusterCreateTestSuite) TestGETClusterTolerates404s() {
	credential := model.Credential{
		ID:           "my-agent-identifier",
		ClientSecret: "my-secret",
		ClientId:     "my-id",
	}

	assert.NoError(suite.T(), registerResponder(credential, http.StatusCreated, "/credentials", http.MethodPost))
	assert.NoError(suite.T(), registerResponder(rolesFromGrantStrings("my-role-id", []string{"api:agentHub:full"}, true), http.StatusOK, "/roles", http.MethodGet))
	assert.NoError(suite.T(), registerResponder([]model.RoleConfig{}, http.StatusOK, "/credentials/my-agent-identifier/roles", http.MethodPut))
	assert.NoError(suite.T(), registerResponder(model.CreateSandboxResponse{ClusterId: "cluster-id"}, 200, "/sandbox/clusters", http.MethodPost))
	httpmock.RegisterResponder(http.MethodGet, "/sandbox/clusters/cluster-id",
		httpmock.NewStringResponder(404, "{ \"error\": \"not found\"}").Times(4),
	)

	cmd := NewClusterCmd(getDefaultAppConfiguration(), getSandboxFileStore())
	cmd.SetOut(io.Discard)
	cmd.SetArgs([]string{
		"create",
	})

	err := cmd.Execute()
	suite.Equal("{ \"error\": \"not found\"}", err.Error())

	info := httpmock.GetCallCountInfo()
	fmt.Print(info)
	suite.Equal(4, info["GET /sandbox/clusters/cluster-id"], "GET should return on its fourth 404")

	httpmock.ZeroCallCounters()
	// should fail immediately on a 500
	httpmock.RegisterResponder(http.MethodGet, "/sandbox/clusters/cluster-id",
		httpmock.NewStringResponder(500, "{ \"error\": \"unexpected err\"}").Times(1),
	)
	err = cmd.Execute()
	suite.Equal("{ \"error\": \"unexpected err\"}", err.Error())
	info = httpmock.GetCallCountInfo()
	suite.Equal(1, info["GET /sandbox/clusters/cluster-id"], "GET should fail immediately on 500")

}

func (suite *ClusterCreateTestSuite) TestUpdateProgressBar() {
	assert.NoError(suite.T(), registerResponder(model.SandboxCluster{PercentComplete: 20}, 200, "/sandbox/clusters/abcd", "GET"))
	o := getDefaultCreateOptions()
	// The output of the progressbar interferes with the ability for the test reader to determine success or failure. Discarding output fixes it.
	o.InitializeProgressBar(io.Discard)
	o.saveData = getSandboxFileStore()
	o.saveData.setCreateSandboxResponse(model.CreateSandboxResponse{ClusterId: "abcd"})

	for i := 0; i <= 99; i += 10 {
		done, err := o.UpdateProgressBar(&model.SandboxCluster{
			ID:                  "some-id",
			IP:                  "0.0.0.0",
			DNS:                 "localhost",
			Status:              "The cluster is being created",
			CreatedAt:           "",
			ExpiresAt:           "",
			PercentComplete:     float32(i),
			NextPercentComplete: float32(i + 1),
		})
		assert.NoError(suite.T(), err)
		assert.False(suite.T(), done)

		loadedData, err := o.saveData.readSandboxFromFile()
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), float32(i), loadedData.SandboxCluster.PercentComplete)
	}

	done, err := o.UpdateProgressBar(&model.SandboxCluster{
		ID:                  "some-id",
		IP:                  "0.0.0.0",
		DNS:                 "localhost",
		Status:              "The cluster is being created",
		CreatedAt:           "",
		ExpiresAt:           "",
		PercentComplete:     100,
		NextPercentComplete: 100,
	})

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), done)

}

func (suite *ClusterCreateTestSuite) TestCreateNamedCredential() {
	o := getDefaultCreateOptions()
	credential := o.createNamedCredential("xyz")
	assert.Equal(suite.T(), "xyz-temp-cluster-credentials", credential.Name)
}

func (suite *ClusterCreateTestSuite) TestRandomString() {
	assert.Len(suite.T(), randomString(6), 6)
	assert.Len(suite.T(), randomString(11), 11)
}

func (suite *ClusterCreateTestSuite) TestAssignCredentialRNARole() {
	ctx := context.Background()
	cases := []struct {
		name                  string
		existingGrants        []string
		rolesAreSystemDefined bool
		expectedErr           error
	}{
		{
			name:                  "happy path, role found and assigned",
			existingGrants:        []string{"api:agentHub:full", "api:deployments:full"},
			rolesAreSystemDefined: true,
			expectedErr:           nil,
		},
		{
			name:                  "role not system defined",
			existingGrants:        []string{"api:agentHub:full", "api:deployments:full"},
			rolesAreSystemDefined: false,
			expectedErr:           agent.ErrRoleMissing,
		},
		{
			name:           "role not found",
			existingGrants: []string{"api:not:full", "api:deployments:full"},
			expectedErr:    agent.ErrRoleMissing,
		},
	}

	for _, c := range cases {
		suite.Run(c.name, func() {
			o := getDefaultCreateOptions()
			credential := model.Credential{
				ID: "my-agent-identifier",
			}
			reqBytes, err := json.Marshal([]string{})
			assert.NoError(suite.T(), err)
			assert.NoError(suite.T(), registerResponder(rolesFromGrantStrings("my-role-id", c.existingGrants, c.rolesAreSystemDefined), http.StatusOK, "/roles", http.MethodGet))
			assert.NoError(suite.T(), registerResponder(reqBytes, http.StatusOK, "/credentials/my-agent-identifier/roles", http.MethodPut))

			err = AssignCredentialRNARole(ctx, &credential, o.ArmoryClient, "x")
			callCount := httpmock.GetCallCountInfo()
			suite.Equal(1, callCount["GET /roles"])
			if c.expectedErr != nil {
				suite.ErrorIs(err, agent.ErrRoleMissing)
				suite.Equal(0, callCount["PUT /credentials/my-agent-identifier/roles"])
			} else {
				suite.Equal(1, callCount["PUT /credentials/my-agent-identifier/roles"])
			}
		})

	}
}

func (suite *ClusterCreateTestSuite) TestCreateSandboxRequest() {
	o := getDefaultCreateOptions()
	request := o.createSandboxRequest("a12x3v", &model.Credential{
		ClientSecret: "super-secret",
		ClientId:     "my-id",
	})
	assert.Equal(suite.T(), "a12x3v-sandbox-rna", request.AgentIdentifier)
	assert.Equal(suite.T(), "my-id", request.ClientId)
	assert.Equal(suite.T(), "super-secret", request.ClientSecret)
}

func getDefaultCreateOptions() *CreateOptions {
	o := NewCreateOptions(getSandboxFileStore())
	o.InitializeConfiguration(getDefaultAppConfiguration())
	return o
}

func getSandboxFileStore() SandboxStorage {
	if os.Getenv("CI") == "true" {
		return &InMemorySandboxStorage{}
	}
	return &SandboxClusterFileStore{}
}

func getDefaultAppConfiguration() *config.Configuration {
	token := "some-token"
	addr := "https://localhost"
	clientId := ""
	clientSecret := ""
	output := "text"
	isTest := true
	return config.New(&config.Input{
		AccessToken:  &token,
		ApiAddr:      &addr,
		ClientId:     &clientId,
		ClientSecret: &clientSecret,
		OutFormat:    &output,
		IsTest:       &isTest,
	})
}

func registerResponder(body any, status int, url, method string) error {
	responder, err := httpmock.NewJsonResponder(status, body)
	if err != nil {
		return err
	}
	httpmock.RegisterResponder(method, url, responder)
	return nil
}

func rolesFromGrantStrings(roleId string, grants []string, systemDefined bool) []model.RoleConfig {
	var roles []model.RoleConfig
	for _, grant := range grants {
		g := strings.Split(grant, ":")
		roles = append(roles, model.RoleConfig{
			ID:            roleId,
			EnvID:         "",
			Name:          "",
			Tenant:        "",
			SystemDefined: systemDefined,
			Grants: []model.GrantConfig{
				{
					Type:       g[0],
					Resource:   g[1],
					Permission: g[2],
				},
			},
		})
	}
	return roles
}

type InMemorySandboxStorage struct {
	SandboxClusterFileStore
}

func (s *InMemorySandboxStorage) writeToSandboxFile() error {
	return nil
}

func (s *InMemorySandboxStorage) readSandboxFromFile() (*model.SandboxSaveData, error) {
	return &s.saveData, nil
}
