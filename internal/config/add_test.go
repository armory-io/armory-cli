package config

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

const testAddCommandName = "add"

func TestAddSimpleContext(t *testing.T) {
	cases := []struct {
		args  []string
		cfg   string
		check func(t *testing.T, c *Config, err error, output string)
	}{
		// required params
		{
			[]string{
				"--" + ParamAddClientId, "my-client-id",
			},
			"",
			func(t *testing.T, c *Config, err error, output string) {
				assert.NotNil(t, err)
				assert.Equal(t, "required flag(s) \"secret\" not set", err.Error())
			},
		},
		// can add an Armory cloud account
		{
			[]string{
				"--" + ParamAddSecret, "my-secret",
				"--" + ParamAddClientId, "my-client-id",
			},
			"",
			func(t *testing.T, c *Config, err error, output string) {
				assert.Nil(t, err)
				if assert.NotNil(t, c) {
					assert.Equal(t, 1, len(c.Contexts))
					ctx := c.Contexts[0]
					assert.Equal(t, ctx.Name, "default")
					// identity default values should not be written to config
					// just loaded in memory with defaults
					assert.Empty(t, ctx.Identity.Armory.TokenIssuerUrl)
					// but gRPC endpoint should be written to disk
					assert.Equal(t, "deploy.cloud.armory.io:443", ctx.Connection.Grpc)
					assert.Equal(t, "my-secret", ctx.Identity.Armory.Secret)
					assert.Equal(t, "my-client-id", ctx.Identity.Armory.ClientId)
				}
			},
		},
		// can override default values
		{
			[]string{
				"--" + ParamAddSecret, "my-secret",
				"--" + ParamAddClientId, "my-client-id",
				"--" + ParamEndpoint, "my-endpoint:443",
			},
			"",
			func(t *testing.T, c *Config, err error, output string) {
				assert.Nil(t, err)
				if assert.NotNil(t, c) {
					assert.Equal(t, 1, len(c.Contexts))
					ctx := c.Contexts[0]
					assert.Equal(t, "my-endpoint:443", ctx.Connection.Grpc)
				}
			},
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("test-%d", i), func(t2 *testing.T) {
			f, err := ioutil.TempFile(os.TempDir(), "cfg-")
			assert.Nil(t, err)
			// If we start empty - just delete - yes we could have not created it!
			if c.cfg == "" {
				os.Remove(f.Name())
			} else {
				assert.Nil(t, ioutil.WriteFile(f.Name(), []byte(c.cfg), 0600))
			}
			defer os.Remove(f.Name())

			// point config
			assert.Nil(t, os.Setenv("ARMORY_CONFIG", f.Name()))
			defer os.Setenv("ARMORY_CONFIG", "")

			b := bytes.NewBufferString("")
			cmd := fakeAddCmd()
			args := append([]string{testAddCommandName}, c.args...)
			cmd.SetArgs(args)
			cmd.SetOut(b)
			err = cmd.Execute()
			if err != nil {
				c.check(t2, nil, err, b.String())
			} else {
				cfg, err := loadConfig(false)
				c.check(t2, cfg, err, b.String())
			}
		})
	}
}

// fakeAddCmd like deploy-cli.main. This is not an amazing solution and we
// should probably consider moving the cmd definition in this module (and for all commands)
func fakeAddCmd() *cobra.Command {
	root := &cobra.Command{}
	root.PersistentFlags().StringP(ParamContext, "C", "default", "")
	root.PersistentFlags().String(ParamEndpoint, "deploy.cloud.armory.io:443", "")
	root.PersistentFlags().Bool(ParamInsecure, false, "")
	root.PersistentFlags().Bool(ParamPlaintext, false, "")
	root.PersistentFlags().Bool(ParamNoProxy, false, "")
	root.PersistentFlags().String(ParamCacert, "", "")
	root.PersistentFlags().String(ParamCert, "", "")
	root.PersistentFlags().String(ParamKey, "", "")
	root.PersistentFlags().String(ParamKeyPassword, "", "")
	root.PersistentFlags().String(ParamServerName, "", "")
	root.PersistentFlags().String(ParamToken, "", "")
	root.PersistentFlags().Bool(ParamAnonymously, false, "")

	cmd := &cobra.Command{
		Use: testAddCommandName,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Add(cmd)
		},
	}
	cmd.Flags().String(ParamAddClientId, "", "")
	cmd.Flags().String(ParamAddSecret, "", "")
	cmd.Flags().String(ParamAddAudience, "", "")
	cmd.Flags().String(ParamAddTokenIssuerUrl, "", "")

	_ = cmd.MarkFlagRequired(ParamAddSecret)
	_ = cmd.MarkFlagRequired(ParamAddClientId)

	root.AddCommand(cmd)
	return root
}
