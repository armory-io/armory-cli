package org

import (
	"encoding/json"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/suite"
	"net/url"
	"testing"
)

func TestOrgTestSuite(t *testing.T) {
	suite.Run(t, new(OrgTestSuite))
}

type OrgTestSuite struct {
	suite.Suite
}

func (suite *OrgTestSuite) SetupSuite() {
	httpmock.Activate()
}

func (suite *OrgTestSuite) SetupTest() {
	httpmock.Reset()
}

func (suite *OrgTestSuite) TearDownSuite() {
	httpmock.DeactivateAndReset()
}

func (suite *OrgTestSuite) TestGetEnvironmentsSuccess() {
	envs := []Environment{
		{
			Id:    "1",
			Name:  "env1",
			OrgId: "org1",
		},
		{
			Id:    "2",
			Name:  "env2",
			OrgId: "org1",
		},
	}
	resp, err := json.Marshal(envs)
	if err != nil {
		suite.T().Fatalf("TestGetEnvironmentsSuccess failed with: %s", err)
	}
	httpmock.RegisterResponder("GET", "http://localhost/environments",
		httpmock.NewStringResponder(200, string(resp)))

	received, err := GetEnvironments(&url.URL{Scheme: "http", Host: "localhost"}, nil)
	if err != nil {
		suite.T().Fatalf("TestGetEnvironmentsSuccess failed with: %s", err)
	}
	suite.Equal(len(envs), len(received), "size should be the same")
}

func (suite *OrgTestSuite) TestGetEnvironmentsHttpFail() {
	httpmock.RegisterResponder("GET", "http://localhost/environments",
		httpmock.NewStringResponder(500, `{"error_id":"a814890f-e0cf-4ae5-a78a-39040ed51a35","errors":[{"code":"99001","message":"No valid auth credentials found."}]}`))

	received, err := GetEnvironments(&url.URL{Scheme: "http", Host: "localhost"}, nil)
	if err == nil {
		suite.T().Fatal("TestGetEnvironmentsHttpFail failed with: error shouldn't be null")
	}
	if received != nil {
		suite.T().Fatal("TestGetEnvironmentsHttpFail failed with: list of the envs should be null")
	}
	suite.Error(err, "should throw an error")
}
