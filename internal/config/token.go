package config

import (
	"context"
	"errors"
	"fmt"
	"github.com/armory-io/go-cloud-service/pkg/token"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

func GetToken(ctx context.Context, log *logrus.Logger, cmd *cobra.Command) error {
	ct, err := getContext(cmd)
	if err != nil {
		return err
	}
	var identity *token.Identity
	if ct == nil {
		_, identity, err = serviceIdentityFromCliOptions(cmd)
		if err != nil {
			return err
		}
	} else {
		identity = &ct.Identity
	}
	cred := token.GetRPCCredentials(*identity, ctx, log)
	m, err := cred.GetRequestMetadata(ctx, "")
	if err != nil {
		return err
	}
	h, ok := m["authorization"]
	if !ok {
		return errors.New("unable to obtain a token")
	}
	token := strings.TrimPrefix(h, "Bearer ")
	fmt.Println(token)
	return nil
}
