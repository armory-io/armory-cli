package version

import (
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/spf13/cobra"
	log "go.uber.org/zap"
)

var Version = "development"

const (
	versionShort   = "Print armory's version"
	versionLong    = "Print armory's version"
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
	log.S().Infof("{\"version\":\"%v\"}", Version)
	return nil
}
