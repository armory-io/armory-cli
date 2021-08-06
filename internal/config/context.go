package config

import (
	"fmt"
	"github.com/armory-io/go-cloud-service/pkg/client"
	tls2 "github.com/armory-io/go-cloud-service/pkg/tls"
	"github.com/armory-io/go-cloud-service/pkg/token"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

const (
	ParamContext     = "context"
	ParamEndpoint    = "endpoint"
	ParamInsecure    = "insecure"
	ParamPlaintext   = "plaintext"
	ParamNoProxy     = "no-proxy"
	ParamCacert      = "cacert"
	ParamCert        = "cert"
	ParamKey         = "key"
	ParamKeyPassword = "key-password"
	ParamServerName  = "server-name"
	ParamToken       = "token"
	ParamAnonymously = "anonymously"
)

func getContext(cmd *cobra.Command) (*Context, error) {
	c, err := loadConfig(true)
	if err != nil {
		return nil, err
	}

	ctx, err := cmd.Flags().GetString(ParamContext)
	if err != nil {
		return nil, err
	}
	if ctx == "" {
		ctx = c.CurrentContext
	}
	for _, ct := range c.Contexts {
		if ct.Name == ctx {
			return &ct, nil
		}
	}
	return nil, nil
}

func GetClientConnection(log *logrus.Logger, cmd *cobra.Command) (*client.Connection, error) {
	ctx, err := getContext(cmd)
	if err != nil {
		return nil, err
	}
	if ctx == nil {
		return cliClientOptions(log, cmd)
	}
	conn := ctx.NewConnection(log)
	return &conn, nil
}

func serviceIdentityFromCliOptions(cmd *cobra.Command) (*client.Service, *token.Identity, error) {
	svc := client.Service{}
	identity := token.Identity{}
	var err error

	if svc.Grpc, err = cmd.Flags().GetString(ParamEndpoint); err != nil {
		return nil, nil, err
	}

	if svc.Insecure, err = cmd.Flags().GetBool(ParamPlaintext); err != nil {
		return nil, nil, err
	}

	if !svc.Insecure {
		tls := &tls2.Settings{}
		svc.Tls = tls
		if tls.InsecureSkipVerify, err = cmd.Flags().GetBool(ParamInsecure); err != nil {
			return nil, nil, err
		}
		if tls.ClientCertFile, err = cmd.Flags().GetString(ParamCert); err != nil {
			return nil, nil, err
		}

		if tls.CAcertFile, err = cmd.Flags().GetString(ParamCacert); err != nil {
			return nil, nil, err
		}

		if tls.ClientKeyFile, err = cmd.Flags().GetString(ParamKey); err != nil {
			return nil, nil, err
		}

		if tls.ClientKeyPassword, err = cmd.Flags().GetString(ParamKeyPassword); err != nil {
			return nil, nil, err
		}
	}

	if svc.NoProxy, err = cmd.Flags().GetBool(ParamNoProxy); err != nil {
		return nil, nil, err
	}

	if identity.Token, err = obtainToken(cmd); err != nil {
		return nil, nil, err
	}
	return &svc, &identity, nil
}

func cliClientOptions(log *logrus.Logger, cmd *cobra.Command) (*client.Connection, error) {
	svc, identity, err := serviceIdentityFromCliOptions(cmd)
	if err != nil {
		return nil, err
	}
	conn := client.New(*svc, identity, log)
	return &conn, nil
}

func obtainToken(cmd *cobra.Command) (string, error) {
	// If anonymous, skip the verification, token will be empty
	if anon, err := cmd.Flags().GetBool(ParamAnonymously); err != nil || anon {
		return "", err
	}
	token, err := cmd.Flags().GetString(ParamToken)
	if err != nil {
		token = os.Getenv(deployCliEnvVarPrefix + "_TOKEN")
	}
	if token == "" {
		return "", fmt.Errorf("Unable to locate authentication information. Use --%s or %s_TOKEN environment variable.", ParamToken, deployCliEnvVarPrefix)
	}
	return token, nil
}
