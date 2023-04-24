// Deprecated: org is another client in this project and should be replaced by the client in the configuration package.
package org

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	errorUtils "github.com/armory/armory-cli/pkg/errors"
	"github.com/armory/armory-cli/pkg/util"
)

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
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) String() string {
	return fmt.Sprintf("[%v] %s", e.Code, e.Message)
}

const (
	CONNECTED_AGENTS_URI string = "/identity/connected-agents"
)

func GetAgents(ArmoryCloudAddr *url.URL, accessToken string) ([]Agent, error) {
	connectedAgentsUrl := &url.URL{
		Scheme: ArmoryCloudAddr.Scheme,
		Host:   ArmoryCloudAddr.Host,
		Path:   CONNECTED_AGENTS_URI,
	}
	request := util.NewHttpRequest("GET", connectedAgentsUrl.String(), nil, &accessToken)
	resp, err := request.Execute()
	if err != nil {
		return nil, ErrNoConnectedAgents
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
		return nil, ErrUnauthorized
	}

	var errorResponse *ApiError
	err = dec.Decode(&errorResponse)
	if err != nil {
		return nil, err
	}
	errContext := fmt.Sprintf(". ErrorId: %s, Desc: %v", errorResponse.ErrorId, errorResponse.Errors)
	return nil, errorUtils.NewErrorWithDynamicContext(ErrRNARetrieval, errContext)
}
