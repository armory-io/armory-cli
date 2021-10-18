package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lestrrat-go/jwx/jwt"
	log "github.com/sirupsen/logrus"
	"net/http"
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

var timeout = 5 * time.Second
var httpClient = http.Client{
	Timeout: timeout,
}

func GetDeviceCodeFromAuthorizationServer(clientId, scope, audience, authUrl string) (*DeviceTokenData, error) {
	requestBody, err := json.Marshal(map[string]string{
		"client_id": clientId,
		"scope": scope,
		"audience": audience,
	})
	if err != nil {
		return nil, errors.New("failed to create request body for Armory authorization server")
	}

	getDeviceCodeRequest, err := http.NewRequest(
		"POST",
		authUrl + "/device/code",
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


func PollAuthorizationServerForResponse(clientId, authUrl string, deviceTokenResponse *DeviceTokenData, authStartedAt time.Time) (string, error) {
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
			"client_id": clientId,
			"device_code": deviceTokenResponse.DeviceCode,
			"grant_type": "urn:ietf:params:oauth:grant-type:device_code",
		})
		if err != nil {
			return "", errors.New("failed to create request body for Armory authorization server")
		}

		getAuthTokenRequest, err := http.NewRequest(
			"POST",
			authUrl + "/token",
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

func ValidateJwt(encodedJwt string) (jwt.Token, error) {
	token, err := jwt.Parse([]byte(encodedJwt), jwt.WithValidate(true))
	if err != nil {
		return nil, err
	}
	return token, nil
}