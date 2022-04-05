package assembler

import (
	"fmt"
	"github.com/armory/armory-cli/cmd"
	"github.com/armory/armory-cli/cmd/deploy"
	"github.com/armory/armory-cli/cmd/login"
	"github.com/armory/armory-cli/cmd/logout"
	"github.com/armory/armory-cli/cmd/schema"
	"github.com/armory/armory-cli/cmd/template"
	"github.com/armory/armory-cli/cmd/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
	"strings"
)

func AddSubCommands(rootCmd *cobra.Command, rootOpts *cmd.RootOptions) {
	rootCmd.AddCommand(version.NewCmdVersion())
	rootCmd.AddCommand(deploy.NewDeployCmd(rootOpts))
	rootCmd.AddCommand(template.NewTemplateCmd(rootOpts))
	rootCmd.AddCommand(schema.NewSchemaCmd(rootOpts))
	rootCmd.AddCommand(login.NewLoginCmd(rootOpts))
	rootCmd.AddCommand(logout.NewLogoutCmd(rootOpts))
	setPersistentFlagsFromEnvVariables(rootCmd.Commands())
}

func setPersistentFlagsFromEnvVariables(commands []*cobra.Command) {
	for _, cmd := range commands {
		cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
			envVar := FlagToEnvVarName(f)
			if val, present := os.LookupEnv(envVar); present {
				cmd.PersistentFlags().Set(f.Name, val)
			}
		})
	}
}

func FlagToEnvVarName(f *pflag.Flag) string {
	return fmt.Sprintf("ARMORY_%s", strings.Replace(strings.ToUpper(f.Name), "-", "_", -1))
}
