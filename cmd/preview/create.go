package preview

import (
	"context"
	"fmt"
	preview "github.com/armory-io/preview-service/pkg/client"
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	"path"
	"time"
)

const (
	createShort = "Create a network preview"
	createLong  = "Create a network preview. Automatically updates your Kubernetes config file by adding the preview cluster connection and updating the current context"

	clusterPreviewType = "cluster"
)

type (
	createPreviewOptions struct {
		Type     string
		Duration string
		Agent    string
	}
)

func NewCmdCreate(configuration *config.Configuration) *cobra.Command {
	logger := zap.S()
	options := &createPreviewOptions{}

	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{},
		Short:   createShort,
		Long:    createLong,
		RunE: func(cmd *cobra.Command, args []string) error {
			assertEqual(logger, clusterPreviewType, options.Type, fmt.Sprintf("preview type must be %q, got %q", clusterPreviewType, options.Type))

			client := preview.NewClient(func(ctx context.Context) (string, error) {
				return configuration.GetAuthToken(), nil
			}, configuration.GetArmoryCloudAddr().String())

			duration, err := time.ParseDuration(options.Duration)
			assertNil(logger, err, "Provided duration is not a valid Go duration string")

			p, err := client.CreateClusterPreview(cmd.Context(), preview.ClusterPreviewParameters{
				AgentIdentifier: options.Agent,
				Duration:        duration,
			})
			assertNil(logger, err, "Could not create cluster preview")

			home, err := homedir.Dir()
			assertNil(logger, err, "Could not determine $HOME directory")
			assertNil(
				logger,
				client.UpdateKubeconfigWithClusterPreview(*p, path.Join(home, ".kube", "config")),
				"Could not update ~/.kube/config with cluster preview",
			)
			cmd.Println("Your Kubernetes config has been updated with the preview context.")
			return nil
		},
	}

	cmdUtils.SetPersistentFlagsFromEnvVariables(cmd.Commands())

	cmd.Flags().StringVarP(
		&options.Type,
		"type",
		"",
		"",
		"The preview type. Options: [cluster]",
	)
	assertNil(logger, cmd.MarkFlagRequired("type"), "Could not mark flag 'type' as required")

	cmd.Flags().StringVarP(
		&options.Duration,
		"duration",
		"",
		"",
		"The preview duration as a Go duration string. Must be less than 24 hours. Example: 60s, 10m, 1h.",
	)
	assertNil(logger, cmd.MarkFlagRequired("duration"), "Could not mark flag 'duration' as required")

	cmd.Flags().StringVarP(
		&options.Agent,
		"agent",
		"",
		"",
		"The agent identifier to use to create the preview.",
	)
	assertNil(logger, cmd.MarkFlagRequired("agent"), "Could not mark flag 'agent' as required")

	return cmd
}

func assertNil(logger *zap.SugaredLogger, err error, msg string) {
	if err != nil {
		logger.Infof("%s: %s", msg, err)
		os.Exit(1)
	}
}

func assertEqual[T comparable](logger *zap.SugaredLogger, left, right T, msg string) {
	if left != right {
		logger.Info(msg)
		os.Exit(1)
	}
}
