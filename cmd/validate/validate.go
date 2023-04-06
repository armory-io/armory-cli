package validate

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueerrors "cuelang.org/go/cue/errors"
	cueyaml "cuelang.org/go/encoding/yaml"
	_ "embed"
	"github.com/armory/armory-cli/cmd/utils"
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

const (
	validateShort = "Validate deployment yaml"
	validateLong  = "Validate deployment yaml\n\n" +
		"For deployment configuration YAML documentation, visit https://docs.armory.io/cd-as-a-service/reference/ref-deployment-file"
	validateExample = "armory deploy validate [options]"
)

//go:embed resources/pipelineRequest.cue
var schemaFile []byte

type validateOptions struct {
	deploymentFile string
}

func NewValidateCmd(configuration *config.Configuration) *cobra.Command {
	options := &validateOptions{}
	cmd := &cobra.Command{
		Use:     "validate --file [<path to file>]",
		Aliases: []string{"validate"},
		Short:   validateShort,
		Long:    validateLong,
		Example: validateExample,
		GroupID: "deployment",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdUtils.ExecuteParentHooks(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return validate(cmd, configuration, options)
		},
	}
	cmd.Flags().StringVarP(&options.deploymentFile, "file", "f", "", "path to the deployment file")
	return cmd
}

func validate(cmd *cobra.Command, configuration *config.Configuration, options *validateOptions) error {
	if *configuration.GetIsTest() {
		utils.ConfigureLoggingForTesting(cmd)
	}
	file, err := os.ReadFile(options.deploymentFile)
	if err != nil {
		return err
	}
	cueContext := cuecontext.New()
	v := cueContext.CompileBytes(schemaFile)
	schema := v.LookupPath(cue.ParsePath("#PipelineRequest"))
	err = cueyaml.Validate(file, schema)
	errList := cueerrors.Errors(err)
	validationFailures := lo.Map[cueerrors.Error, string](errList, func(e cueerrors.Error, _ int) string { return e.Error() })
	if len(validationFailures) > 0 {
		_, err = cmd.OutOrStdout().Write([]byte("YAML is NOT valid. See the following errors:\n\n"))
		_, err = cmd.OutOrStdout().Write([]byte(strings.Join(validationFailures, "\n\n") + "\n"))
	} else {
		_, err = cmd.OutOrStdout().Write([]byte("YAML is valid.\n"))
	}
	return err
}
