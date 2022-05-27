package org

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/armory/armory-cli/pkg/util"
	"net/url"
	"time"
)

type Environment struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	OrgId string `json:"orgId"`
}

type Agent struct {
	AgentIdentifier       string    `json:"agentIdentifier"`
	OrgId                 string    `json:"orgId"`
	EnvId                 string    `json:"envId"`
	AgentInstanceUuid     string    `json:"agentInstanceUuid"`
	StreamId              string    `json:"streamId"`
	IpAddress             string    `json:"ipAddress"`
	NodeIp                string    `json:"nodeIp"`
	AgentVersion          string    `json:"agentVersion"`
	K8SClusterRoleSupport bool      `json:"k8sClusterRoleSupport"`
	ClientId              string    `json:"clientId"`
	ConnectedAtIso8601    time.Time `json:"connectedAtIso8601"`
}

type ApiError struct {
	//{"error_id":"a814890f-e0cf-4ae5-a78a-39040ed51a35","errors":[{"code":99001,"message":"No valid auth credentials found."}]}
	ErrorId string      `json:"error_id"`
	Errors  *[]AppError `json:"errors"`
}

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) String() string {
	return fmt.Sprintf("[%v] %s", e.Code, e.Message)
}

const (
	ENVIRONMENT_URI      string = "/environments"
	CONNECTED_AGENTS_URI string = "/identity/connected-agents"
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
	return nil, fmt.Errorf("error retrieving environment to login to. ErrorId: %s, Desc: %s", errorResponse.ErrorId, errorResponse.Errors)
}

func GetAgents(ArmoryCloudAddr *url.URL, accessToken string) ([]Agent, error) {
	connectedAgentsUrl := &url.URL{
		Scheme: ArmoryCloudAddr.Scheme,
		Host:   ArmoryCloudAddr.Host,
		Path:   CONNECTED_AGENTS_URI,
	}
	request := util.NewHttpRequest("GET", connectedAgentsUrl.String(), nil, &accessToken)
	resp, err := request.Execute()
	if err != nil {
		return nil, errors.New("unable to retrieve agents to connect with; please ensure an agent is connected and try again")
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	if resp.StatusCode == 200 {
		var connectedAgents []Agent
		err = dec.Decode(&connectedAgents)
		if err != nil {
			return nil, err
		}
		return connectedAgents, nil
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("error: Unauthorized. Try `armory login` to ensure you're using the proper environment")
	}

	var errorResponse *ApiError
	err = dec.Decode(&errorResponse)
	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("error retrieving agents to connect with. ErrorId: %s, Desc: %s", errorResponse.ErrorId, errorResponse.Errors)
}
