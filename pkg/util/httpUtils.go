package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

var (
	ErrFailedToMarshalRequest    = errors.New("failed to marshal request to json")
	ErrFailedToCreateHttpRequest = errors.New("failed to create http request")
)

type HttpRequest struct {
	Body        *map[string]string
	Url         string
	Method      string //one of POST, GET, etc.
	BearerToken *string
	httpClient  http.Client
}

func NewHttpRequest(method string, url string, body map[string]string, bearerToken *string, optionalTimeoutSeconds ...time.Duration) HttpRequest {
	var timeout time.Duration = 10
	if len(optionalTimeoutSeconds) > 0 {
		timeout = optionalTimeoutSeconds[0]
	}
	return HttpRequest{
		Url:         url,
		Body:        &body,
		Method:      method,
		BearerToken: bearerToken,
		httpClient: http.Client{
			Timeout: time.Second * timeout,
		},
	}
}

func (request *HttpRequest) Execute() (*http.Response, error) {
	requestBody, err := json.Marshal(&request.Body)
	if err != nil {
		return nil, ErrFailedToMarshalRequest
	}
	clientRequest, err := http.NewRequest(
		request.Method,
		request.Url,
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		return nil, ErrFailedToCreateHttpRequest
	}
	if request.BearerToken != nil {
		clientRequest.Header.Set("Authorization", "Bearer "+*request.BearerToken)
	}
	clientRequest.Header.Set("Content-Type", "application/json")
	resp, err := request.httpClient.Do(clientRequest)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
