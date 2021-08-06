package config

import (
	"github.com/armory-io/go-cloud-service/pkg/token"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestLocations(t *testing.T) {
	h, err := os.UserHomeDir()
	assert.Nil(t, err)
	assert.Equal(t, path.Join(h, ".armory"), getHome())
	assert.Equal(t, path.Join(h, ".armory/config.yaml"), getConfigFile())

	assert.Nil(t, os.Setenv("ARMORY_CONFIG", "/tmp/my-test"))
	defer os.Setenv("ARMORY_CONFIG", "")

	assert.Equal(t, "/tmp/my-test", getConfigFile())
}

func TestConfigNotExist(t *testing.T) {
	assert.Nil(t, os.Setenv("ARMORY_CONFIG", "/tmp/my-test"))
	defer os.Setenv("ARMORY_CONFIG", "")

	c, err := loadConfig(true)
	assert.Nil(t, err)
	assert.Equal(t, &Config{}, c)

	f, err := ioutil.TempFile(os.TempDir(), "cfg-")
	defer os.Remove(f.Name())
	assert.Nil(t, err)
	assert.Nil(t, ioutil.WriteFile(f.Name(), []byte(`
contexts:
- name: test`), 0600))

	assert.Nil(t, os.Setenv("ARMORY_CONFIG", f.Name()))
	defer os.Setenv("ARMORY_CONFIG", "")

	c, err = loadConfig(true)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(c.Contexts))
	ctx := c.Contexts[0]
	assert.Equal(t, "test", ctx.Name)
	identity := token.DefaultIdentity()
	svc := defaultService()
	assert.Equal(t, identity.Armory.TokenIssuerUrl, ctx.Identity.Armory.TokenIssuerUrl)
	assert.Equal(t, svc.Grpc, ctx.Connection.Grpc)
}

func TestConfigNoDefault(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "cfg-")
	defer os.Remove(f.Name())
	assert.Nil(t, err)
	assert.Nil(t, ioutil.WriteFile(f.Name(), []byte(`
contexts:
- name: test`), 0600))

	assert.Nil(t, os.Setenv("ARMORY_CONFIG", f.Name()))
	defer os.Setenv("ARMORY_CONFIG", "")
	// Load without defaults
	c, err := loadConfig(false)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(c.Contexts))
	ctx := c.Contexts[0]
	assert.Equal(t, "test", ctx.Name)
	assert.Empty(t, ctx.Identity.Armory.TokenIssuerUrl)
	assert.Empty(t, ctx.Connection.Grpc)
}
