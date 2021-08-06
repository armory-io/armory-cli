package config

import (
	"github.com/armory-io/go-cloud-service/pkg/client"
	tls2 "github.com/armory-io/go-cloud-service/pkg/tls"
	"github.com/armory-io/go-cloud-service/pkg/token"
	"github.com/spf13/cobra"
)

const (
	ParamAddClientId       = "client-id"
	ParamAddSecret         = "secret"
	ParamAddTokenIssuerUrl = "token-issuer"
	ParamAddAudience       = "audience"
)

func Add(cmd *cobra.Command) error {
	n, err := cmd.Flags().GetString(ParamContext)
	if err != nil {
		return err
	}

	c, err := loadConfig(false)
	if err != nil {
		return err
	}

	clientId, err := cmd.Flags().GetString(ParamAddClientId)
	if err != nil {
		return err
	}

	secret, err := cmd.Flags().GetString(ParamAddSecret)
	if err != nil {
		return err
	}

	tokenIssuer, err := cmd.Flags().GetString(ParamAddTokenIssuerUrl)
	if err != nil {
		return err
	}

	audience, err := cmd.Flags().GetString(ParamAddAudience)
	if err != nil {
		return err
	}

	ctx := Context{
		Name: n,
		Identity: token.Identity{
			Armory: token.ArmoryCloud{
				ClientId:       clientId,
				Secret:         secret,
				Audience:       audience,
				TokenIssuerUrl: tokenIssuer,
			},
		},
		Connection: client.Service{},
	}
	if ctx.Connection.Grpc, err = cmd.Flags().GetString(ParamEndpoint); err != nil {
		return err
	}
	if ctx.Connection.Insecure, err = cmd.Flags().GetBool(ParamPlaintext); err != nil {
		return err
	}
	if !ctx.Connection.Insecure {
		tls := &tls2.Settings{}
		ctx.Connection.Tls = tls
		if tls.InsecureSkipVerify, err = cmd.Flags().GetBool(ParamInsecure); err != nil {
			return err
		}
		if tls.ClientCertFile, err = cmd.Flags().GetString(ParamCert); err != nil {
			return err
		}
		if tls.CAcertFile, err = cmd.Flags().GetString(ParamCacert); err != nil {
			return err
		}
		if tls.ClientKeyFile, err = cmd.Flags().GetString(ParamKey); err != nil {
			return err
		}
		if tls.ClientKeyPassword, err = cmd.Flags().GetString(ParamKeyPassword); err != nil {
			return err
		}
	}

	if ctx.Connection.NoProxy, err = cmd.Flags().GetBool(ParamNoProxy); err != nil {
		return err
	}

	found := false
	for i := range c.Contexts {
		if c.Contexts[i].Name == n {
			c.Contexts[i] = ctx
			found = true
		}
	}
	if !found {
		c.Contexts = append(c.Contexts, ctx)
	}
	return saveConfig(c)
}
