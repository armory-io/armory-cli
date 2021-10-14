package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pkg/browser"
	log "github.com/sirupsen/logrus"
	"io
	"net/http"
	"strings"
	"time"
)

type DeviceTokenData struct {
	DeviceCode     			string `json:"device_code"`
	UserCode       			string `json:"user_code"`
	VerificationUri 		string `json:"verification_uri"`
	ExpiresIn       		int    `json:"expires_in"`
	Interval                int    `json:"interval"`
	VerificationUriComplete string `json:"verification_uri_complete"`
}

type ErrorResponse struct {
	Error       string `json:"error"`
	Description string `json:"error_description"`
}

type SuccessfulResponse struct {
	// AccessToken Encoded JWT / Bearer Token
	AccessToken             string `json:"access_token"`
	// SecondsUtilTokenExpires the number of seconds until the JWT expires, from when it was created by the Auth Server.
	// The JWT has the exact expiration date time
	SecondsUtilTokenExpires int `json:"expires_in"`
}

type Jwt struct {
	PrincipalMetadata *ArmoryCloudPrincipalMetadata `json:"https://cloud.armory.io/principal"`
	ExpiresAt int64 `json:"exp"`
}

type ArmoryCloudPrincipalMetadata struct {
	Name string `json:"name"`
	Type string `json:"type"`
	OrgName string `json:"orgName"`
	TokenExpiration time.Time
}

var timeout = 5 * time.Second
var httpClient = http.Client{
	Timeout: timeout,
}

func GetDeviceCodeFromAuthorizationServer() (*DeviceTokenData, error) {
	requestBody, err := json.Marshal(map[string]string{
		"client_id": armoryCliStagingClientId,
		"scope": armoryAuthScopes,
		"audience": armoryStagingAudience,
	})
	if err != nil {
		return nil, errors.New("failed to create request body for Armory authorization server")
	}

	getDeviceCodeRequest, err := http.NewRequest(
		"POST",
		"https://auth.staging.cloud.armory.io/oauth/device/code",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		return nil, errors.New("failed to create request for Armory authorization server")
	}

	getDeviceCodeRequest.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(getDeviceCodeRequest)
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(resp.Body)
	var deviceTokenResponse DeviceTokenData
	err = dec.Decode(&deviceTokenResponse)
	if err != nil {
		return nil, err
	}
	return &deviceTokenResponse, nil
}


func PollAuthorizationServerForResponse(deviceTokenResponse *DeviceTokenData, authStartedAt time.Time) (string, error) {
	var secondsAfterAuthStartedAtWhenDeviceFlowExpires = deviceTokenResponse.ExpiresIn * 1000 - 5000
	deviceFlowExpiresTime := authStartedAt.Add(time.Duration(secondsAfterAuthStartedAtWhenDeviceFlowExpires) * time.Second)
	log.Infof("Waiting for user to login")
	for {
		if time.Now().After(deviceFlowExpiresTime) {
			log.Infof("%d", secondsAfterAuthStartedAtWhenDeviceFlowExpires)
			log.Infof(authStartedAt.Local().String())
			log.Infof(deviceFlowExpiresTime.Local().String())
			return "", errors.New("the device flow request has expired")
		}

		fmt.Print(".")
		time.Sleep(time.Duration(deviceTokenResponse.Interval) * time.Second)

		requestBody, err := json.Marshal(map[string]string{
			"client_id": armoryCliStagingClientId,
			"device_code": deviceTokenResponse.DeviceCode,
			"grant_type": "urn:ietf:params:oauth:grant-type:device_code",
		})
		if err != nil {
			return "", errors.New("failed to create request body for Armory authorization server")
		}

		getAuthTokenRequest, err := http.NewRequest(
			"POST",
			"https://auth.staging.cloud.armory.io/oauth/token",
			bytes.NewBuffer(requestBody),
		)
		if err != nil {
			return "", errors.New("failed to create request for Armory authorization server")
		}

		getAuthTokenRequest.Header.Set("Content-Type", "application/json")
		resp, err := httpClient.Do(getAuthTokenRequest)
		if err != nil {
			return "", err
		}

		dec := json.NewDecoder(resp.Body)
		if resp.StatusCode == 200 {
			fmt.Print("\n")
			var authSuccessfulResponse SuccessfulResponse
			err = dec.Decode(&authSuccessfulResponse)
			if err != nil {
				return "", err
			}
			err = resp.Body.Close()
			if err != nil {
				return "", errors.New("failed to close resource")
			}
			return authSuccessfulResponse.AccessToken, nil
		}

		var errorResponse *ErrorResponse
		err = dec.Decode(&errorResponse)
		if err != nil {
			return "", err
		}
		err = resp.Body.Close()
		if err != nil {
			return "", errors.New("failed to close resource")
		}

		if errorResponse.Error != "authorization_pending" {
			return "",fmt.Errorf("there was an error polling for user auth. Err: %s, Desc: %s", errorResponse.Error, errorResponse.Description)
		}
	}
}

func decodeJwtMetadata(encodedJwt string) {
	parts := strings.Split(encodedJwt, ".")
	if len(parts) != 3 {
		log.Fatalln("Expected well-formed JWT")
	}
	jwtMeta := parts[1]

	data, err := base64.StdEncoding.DecodeString(jwtMeta)
	if err != nil {
		log.Debug(err)
		log.Fatalln("Failed to decode JWT metadata")
	}

	var jwt Jwt
	dec := json.NewDecoder(bytes.NewReader(data))
	err = dec.Decode(&jwt)
	if err != nil {
		log.Debug(err)
		log.Fatalln("Failed to deserialize principal claim")
	}

	log.Infof("Welcome %s user: %s, your token expires at: %s", jwt.PrincipalMetadata.OrgName, jwt.PrincipalMetadata.Name, time.Unix(jwt.ExpiresAt, 0).Local().String())
}