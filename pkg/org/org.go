package org

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/armory/armory-cli/pkg/util"
	"net/url"
)

type Environment struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	OrgId string `json:"orgId"`
}

type ApiError struct {
	//{"error_id":"a814890f-e0cf-4ae5-a78a-39040ed51a35","errors":[{"code":99001,"message":"No valid auth credentials found."}]}
	ErrorId string      `json:"error_id"`
	Errors  *[]AppError `json:"errors"`
}

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

const (
	ENVIRONMENT_URI string = "/environments"
)

func GetEnvironments(ArmoryCloudAddr *url.URL, accessToken *string) ([]Environment, error) {
	environmentUrl := &url.URL{
		Scheme: ArmoryCloudAddr.Scheme,
		Host:   ArmoryCloudAddr.Host,
		Path:   ENVIRONMENT_URI,
	}
	request := util.NewHttpRequest("GET", environmentUrl.String(), nil, accessToken)
	request.BearerToken = accessToken
	resp, err := request.Execute()
	if err != nil {
		return nil, errors.New("unable to retrieve environments to login to; please try again")
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	if resp.StatusCode == 200 {
		var environments []Environment
		err = dec.Decode(&environments)
		if err != nil {
			return nil, err
		}
		return environments, nil
	}

	var errorResponse *ApiError
	err = dec.Decode(&errorResponse)
	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("error retrieving environment to login to. ErrorId: %s, Desc: %s", errorResponse.ErrorId, (*errorResponse.Errors)[0].Message)
}
