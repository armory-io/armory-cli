package aws

import (
	"fmt"
	"github.com/armory/armory-cli/pkg/config"
	clitesting "github.com/armory/armory-cli/pkg/testing"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func getStdTestConfig(token string, outFmt string) *config.Configuration {
	addr := "https://localhost"
	clientId := ""
	clientSecret := ""
	isTest := true
	return config.New(&config.Input{
		AccessToken:  &token,
		ApiAddr:      &addr,
		ClientId:     &clientId,
		ClientSecret: &clientSecret,
		OutFormat:    &outFmt,
		IsTest:       &isTest,
	})
}

func TestCreateRoleNoResponse(t *testing.T) {
	token, err := clitesting.CreateFakeJwt()
	assert.NoError(t, err)
	r, w, err := os.Pipe()
	if err != nil {
		t.Errorf("Cannot create stdin pipe: %v", err)
	}
	if _, err := w.WriteString(fmt.Sprintf("%v\n", "N")); err != nil {
		t.Errorf("Cannot write output to prompt: %v", err)
	}
	cmd := NewCreateRoleCmd(getStdTestConfig(token, "text"), r)
	err = cmd.Execute()
	assert.EqualError(t, err, "", "The prompt should exit, but we should not encounter an error prior to it")
}
