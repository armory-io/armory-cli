package auth

import (
	"encoding/json"
	clitesting "github.com/armory/armory-cli/pkg/testing"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

type AuthTestSuite struct {
	suite.Suite
}

func (suite *AuthTestSuite) SetupSuite() {
	httpmock.Activate()
}

func (suite *AuthTestSuite) SetupTest() {
	httpmock.Reset()
}

func (suite *AuthTestSuite) TearDownSuite() {
	httpmock.DeactivateAndReset()
}

func (suite *AuthTestSuite) TestTokenAuthSuccess() {
	jwt, err := clitesting.CreateFakeJwt()
	if err != nil {
		suite.T().Fatalf("TestTokenAuthSuccess failed with: %s", err)
	}
	auth := NewAuth("", "", "", "", "", jwt)
	token, err := auth.GetToken()
	suite.Nilf(err, "TestTokenAuthSuccess failed getting token: %s", err)
	suite.Equal(jwt, token, "TestTokenAuthSuccess: Token and Jwt must be equal")
	environment, err := auth.GetEnvironmentId()
	suite.Nilf(err, "TestTokenAuthSuccess failed getting environment: %s", err)
	suite.Equal(environment, "12345", "TestTokenAuthSuccess: Environment and Jwt envId must be equal")
}

func (suite *AuthTestSuite) TestAuthenticationShouldErrorWhenTokenIsProvided() {
	jwt, err := clitesting.CreateFakeJwt()
	if err != nil {
		suite.T().Fatalf("TestTokenAuthSuccess failed with: %s", err)
	}
	authy := NewAuth("", "", "", "", "", jwt)
	_, _, err = authy.authentication()
	suite.NotNil(err, "AuthenticationShouldErrorWhenTokenIsProvided expects an error with authenticating remotely %s", err)
	suite.Equal("do not try to execute remote authentication when a Token has been provided to the command", err.Error(), "AuthenticationShouldErrorWhenTokenIsProvided: expected a specific error but found: %s", err)
}

func (suite *AuthTestSuite) TestAuthSuccess() {
	jwt, err := clitesting.CreateFakeJwt()
	if err != nil {
		suite.T().Fatalf("TestAuthSuccess failed with: %s", err)
	}
	rt := &remoteToken{
		AccessToken: jwt,
	}
	resp, err := json.Marshal(rt)
	if err != nil {
		suite.T().Fatalf("TestAuthSuccess failed with: %s", err)
	}
	httpmock.RegisterResponder("POST", "http://localhost/oauth/token",
		httpmock.NewStringResponder(200, string(resp)))
	auth := NewAuth("test", "pass", "client_credentials", "http://localhost/oauth", "http://localhost", "")
	token, exp, err := auth.authentication()
	suite.Nilf(err, "TestAuthSuccess failed with: %s", err)
	suite.NotNil(exp, "TestAuthSuccess failed with: expiration must not be null")
	suite.Equal(jwt, token, "TestAuthSuccess: Token and Jwt must be equal")
}

func (suite *AuthTestSuite) TestAuthFail() {
	httpmock.RegisterResponder("POST", "http://localhost/oauth/token",
		httpmock.NewStringResponder(401, ""))
	auth := NewAuth("test", "pass", "client_credentials", "http://localhost/oauth", "http://localhost", "")
	_, _, err := auth.authentication()
	suite.NotNil(err, "TestAuthFail failed with: err is null")
	suite.Error(err, "unexpected status code while getting token 401")
}

func (suite *AuthTestSuite) TestAuthFailWithInvalidJwt() {
	rt := &remoteToken{
		AccessToken: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOlsiaHR0cDovL2xvY2FsaG9zdCJdLCJpYXQiOjIzMzQzMTIwMCwic3ViIjoiYXJtb3J5LWNsaSJ9.dK4XAG9sWbA0I8Sd5Wh-UPIkqsgucTt5REc3BdTXe8POrQAZbYf3qXCaQ2DXQyW3YGlgHXXMfOigdiIOKkO06t7B6__7MElCWCBFsJAzroBL2JtImHaXQLqYLJUHmXmHGPfUbAWFZEhvNMhsuYIsmJsM-tJ7dDMi-iEHOuLsGeYmmoMzLFwy0reNbD40gsRlOyuSqrhQJXv5E16m4mKNkDtZsc5Y1pMUEtZrjYbADEtFzojNDmQLEf0vYoh7XgEEP3IEhClI0O_ghnCCN6o0n3ZWJvz6mBorHUs0zXUD_XvBQtpwibQgGmjtuOBu2iEshWYihcAV91Bb52slT3GdU",
	}
	resp, err := json.Marshal(rt)
	if err != nil {
		suite.T().Fatalf("TestAuthFailWithInvalidJwt failed with: %s", err)
	}
	httpmock.RegisterResponder("POST", "http://localhost/oauth/token",
		httpmock.NewStringResponder(200, string(resp)))
	auth := NewAuth("test", "pass", "client_credentials", "http://localhost/oauth", "http://differentaudience/", "")
	_, _, err = auth.authentication()
	suite.NotNil(err, "TestAuthFailWithInvalidJwt failed with: err is null")
}
