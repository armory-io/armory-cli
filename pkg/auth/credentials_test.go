package auth

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWriteAndLoadCredentialsSuccess(t *testing.T){
	credentials := Credentials{
		ClientId: "123",
		Source: "test",
		Audience: "http://armory-deployments",
	}
	tempPath := t.TempDir() + "/credentials"
	credentials.WriteCredentials(tempPath)
	assert.FileExists(t, tempPath)
	received, err := LoadCredentials(tempPath)
	if err != nil {
		t.Fatalf("TestLoadCredentialsSuccess failed with %s", err)
	}
	assert.EqualValues(t, credentials, received)
}

func TestGetEnvironmentSuccess(t *testing.T){
	token, err := createFakeJwt()
	if err != nil {
		t.Fatalf("TestGetEnvironmentSuccess failed with %s", err)
	}
	credentials := Credentials{
		Token: token,
	}
	env, err := credentials.GetEnvironment()
	if err != nil {
		t.Fatalf("TestGetEnvironmentSuccess failed with %s", err)
	}
	assert.EqualValues(t, "12345", env)
}