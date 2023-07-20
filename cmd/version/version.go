package version

import (
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/console"
	"github.com/spf13/cobra"
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
		Run: func(cmd *cobra.Command, args []string) {
			RunVersion(cmd)
		},
	}
	return cmd
}

func RunVersion(cmd *cobra.Command) {
	console.Stdoutf("{\"version\":\"%v\"}", Version)
}
