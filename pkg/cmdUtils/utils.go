package cmdUtils

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
	"strings"
)

// ExecuteParentHooks runs parents PersistentPreRun hooks
func ExecuteParentHooks(cmd *cobra.Command, args []string) {
	for cmd.HasParent() {
		cmd = cmd.Parent()
		if cmd.PersistentPreRun != nil {
			cmd.PersistentPreRun(cmd, args)
		}
	}
}

func SetPersistentFlagsFromEnvVariables(commands []*cobra.Command) {
	for _, cmd := range commands {
		cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
			envVar := flagToEnvVarNameKebabCaseFlags(f)
			if val, present := os.LookupEnv(envVar); present {
				cmd.PersistentFlags().Set(f.Name, val)
			}
			envVar = flagToEnvVarNameCamelCaseFlags(f)
			if val, present := os.LookupEnv(envVar); present {
				cmd.PersistentFlags().Set(f.Name, val)
			}
		})
	}
}

// flagToEnvVarNameKebabCaseFlags
// Since the original plan was to have flags be kebab case flagToEnvVarNameKebabCaseFlags converts kebab flags to env vars.
// but since all the flags where camel case this results in env vars that aren't readable clientId becomes ARMORY_CLIENTID
// flagToEnvVarNameCamelCaseFlags adds support for ARMORY_CLIENT_ID
func flagToEnvVarNameKebabCaseFlags(f *pflag.Flag) string {
	return fmt.Sprintf("ARMORY_%s", strings.Replace(strings.ToUpper(f.Name), "-", "_", -1))
}

// flagToEnvVarNameCamelCaseFlags
// Since the original plan was to have flags be kebab case flagToEnvVarNameKebabCaseFlags converts kebab flags to env vars.
// but since all the flags where camel case this results in env vars that aren't readable clientId becomes ARMORY_CLIENTID
// this method converts the flag to kebab case then into an env var, so that ARMORY_CLIENT_ID will be supported.
func flagToEnvVarNameCamelCaseFlags(f *pflag.Flag) string {
	flagAsKebab := strcase.ToKebab(f.Name)
	return fmt.Sprintf("ARMORY_%s", strings.Replace(strings.ToUpper(flagAsKebab), "-", "_", -1))
}
