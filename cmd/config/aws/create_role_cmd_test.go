package aws

import (
	"github.com/armory/armory-cli/pkg/config"
	clitesting "github.com/armory/armory-cli/pkg/testing"
	"github.com/stretchr/testify/assert"
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

func TestCreateRoleErrors(t *testing.T) {
	token, err := clitesting.CreateFakeJwt()
	assert.NoError(t, err)
	cmd := NewCreateRoleCmd(getStdTestConfig(token, "text"))
	err = cmd.Execute()
	assert.EqualError(t, err, "^D", "The prompt should exit, but we should not encounter an error prior to it")
}
