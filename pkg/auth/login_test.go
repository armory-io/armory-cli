package auth

import (
	"encoding/json"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

func TestLoginTestSuite(t *testing.T) {
	suite.Run(t, new(LoginTestSuite))
}

type LoginTestSuite struct {
	suite.Suite
}

func (suite *LoginTestSuite) SetupSuite() {
	httpmock.Activate()
}

func (suite *LoginTestSuite) SetupTest() {
	httpmock.Reset()
}

func (suite *LoginTestSuite) TearDownSuite() {
	httpmock.DeactivateAndReset()
}

func (suite *LoginTestSuite) TestGetDeviceCodeSuccess() {
	device := &DeviceTokenData{
		DeviceCode: "123",
		UserCode: "ABCDE",
		VerificationUri: "http://localhost/activate",
		ExpiresIn: 900,
		Interval: 5,
		VerificationUriComplete: "http://localhost/activate?user_code=ABCDE",

	}
	resp, err := json.Marshal(device)
	if err != nil {
		suite.T().Fatalf("TestGetDeviceCodeSuccess failed with: %s", err)
	}
	httpmock.RegisterResponder("POST", "http://localhost/oauth/device/code",
		httpmock.NewStringResponder(200, string(resp)))
	auth, err := GetDeviceCodeFromAuthorizationServer("123", "add", "http://localhost", "http://localhost/oauth")
	if err != nil {
		suite.T().Fatalf("TestGetDeviceCodeSuccess failed with: %s", err)
	}
	if auth == nil {
		suite.T().Fatal("TestGetDeviceCodeSuccess failed with: token is Empty")
	}
	suite.EqualValues(device, auth, "both objects must be equal")
}

func (suite *LoginTestSuite) TestGetDeviceCodeFail() {
	httpmock.RegisterResponder("POST", "http://localhost/oauth/device/code",
		httpmock.NewStringResponder(200, `{"device_code": 123}`))
	result, err := GetDeviceCodeFromAuthorizationServer("123", "add", "http://localhost", "http://localhost/oauth")
	if result != nil {
		suite.T().Fatal("TestGetDeviceCodeFail failed with: result must be null")
	}
	suite.Error(err)
}

func (suite *LoginTestSuite) TestRefreshAuthTokenSuccess() {
	success := &SuccessfulResponse{
		AccessToken: "XYV",
		RefreshToken: "ABCDE",
		SecondsUtilTokenExpires: 900,
	}
	resp, err := json.Marshal(success)
	if err != nil {
		suite.T().Fatalf("TestRefreshAuthTokenSuccess failed with: %s", err)
	}
	httpmock.RegisterResponder("POST", "http://localhost/oauth/token",
		httpmock.NewStringResponder(200, string(resp)))
	received, err := RefreshAuthToken("123", "http://localhost/oauth", "ABCDE", "test")
	if err != nil {
		suite.T().Fatalf("TestRefreshAuthTokenSuccess failed with: %s", err)
	}
	if received == nil {
		suite.T().Fatal("TestRefreshAuthTokenSuccess failed with: token is Empty")
	}
	suite.EqualValues(success, received, "both objects must be equal")
}

func (suite *LoginTestSuite) TestRefreshAuthTokenFail() {
	failed := &ErrorResponse{
		Error: "Invalid request",
		Description: "This is a test",
	}
	resp, err := json.Marshal(failed)
	if err != nil {
		suite.T().Fatalf("TestRefreshAuthTokenFail failed with: %s", err)
	}
	httpmock.RegisterResponder("POST", "http://localhost/oauth/token",
		httpmock.NewStringResponder(400, string(resp)))
	result, err := RefreshAuthToken("123", "http://localhost/oauth", "ABCDE", "test")
	if result != nil {
		suite.T().Fatal("TestRefreshAuthTokenFail failed with: result must be null")
	}
	suite.EqualError(err, "there was an error authorizing for the requested environment. Err: Invalid request, Desc: This is a test")
}

func (suite *LoginTestSuite) TestPollAuthorizationServerForResponseDeviceExpired() {
	device := &DeviceTokenData{
		DeviceCode: "12345",
		UserCode: "ABCDE",
		VerificationUri: "http://localhost/activate",
		ExpiresIn: 0,
		Interval: 0,
		VerificationUriComplete: "http://localhost/activate?user_code=ABCDE",
	}
	authStartedAt := time.Date(1990, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	result, err := PollAuthorizationServerForResponse("123","http://localhost/oauth", device, authStartedAt)
	suite.Nil(result)
	suite.NotNil(err)
	suite.EqualError(err, "the device flow request has expired")
}

func (suite *LoginTestSuite) TestPollAuthorizationServerForResponseHttpFailed() {
	failed := &ErrorResponse{
		Error: "Invalid request",
		Description: "This is a test",
	}
	resp, err := json.Marshal(failed)
	if err != nil {
		suite.T().Fatalf("TestPollAuthorizationServerForResponseHttpFailed failed with: %s", err)
	}
	httpmock.RegisterResponder("POST", "http://localhost/oauth/token",
		httpmock.NewStringResponder(400, string(resp)))
	device := &DeviceTokenData{
		DeviceCode: "12345",
		UserCode: "ABCDE",
		VerificationUri: "http://localhost/activate",
		ExpiresIn: 0,
		Interval: 0,
		VerificationUriComplete: "http://localhost/activate?user_code=ABCDE",
	}
	authStartedAt := time.Date(2999, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	result, err := PollAuthorizationServerForResponse("123","http://localhost/oauth", device, authStartedAt)
	suite.Nil(result)
	suite.NotNil(err)
	suite.EqualError(err, "there was an error polling for user auth. Err: Invalid request, Desc: This is a test")
}

func (suite *LoginTestSuite) TestPollAuthorizationServerForResponseSuccess() {
	success := &SuccessfulResponse{
		AccessToken: "XYV",
		RefreshToken: "ABCDE",
		SecondsUtilTokenExpires: 900,
	}
	resp, err := json.Marshal(success)
	if err != nil {
		suite.T().Fatalf("TestPollAuthorizationServerForResponseSuccess failed with: %s", err)
	}
	httpmock.RegisterResponder("POST", "http://localhost/oauth/token",
		httpmock.NewStringResponder(200, string(resp)))
	device := &DeviceTokenData{
		DeviceCode: "12345",
		UserCode: "ABCDE",
		VerificationUri: "http://localhost/activate",
		ExpiresIn: 0,
		Interval: 0,
		VerificationUriComplete: "http://localhost/activate?user_code=ABCDE",
	}
	authStartedAt := time.Date(2999, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	received, err := PollAuthorizationServerForResponse("123","http://localhost/oauth", device, authStartedAt)
	suite.Nil(err)
	suite.NotNil(received)
	suite.EqualValues(received, success)
}
