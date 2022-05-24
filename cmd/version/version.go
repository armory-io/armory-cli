package version

import (
	"github.com/armory/armory-cli/pkg/cmdUtils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Version = "development"

const (
	versionShort   = "Print the version information of Armory Cli"
	versionLong    = "Print the version information of Armory Cli"
	versionExample = "armory version"
)

func NewCmdVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Short:   versionShort,
		Long:    versionLong,
		Example: versionExample,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdUtils.ExecuteParentHooks(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunVersion(cmd)
		},
	}
	return cmd
}

func RunVersion(cmd *cobra.Command) error {
	log.Infof("{\"version\":\"%v\"}", Version)
	return nil
}
