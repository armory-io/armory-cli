package schema

import (
	"encoding/json"
	"fmt"
	"github.com/armory/armory-cli/cmd"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/danielgtaylor/huma/schema"
	"github.com/spf13/cobra"
	"reflect"
)

func NewSchemaCmd(options *cmd.RootOptions) *cobra.Command {
	command := &cobra.Command{
		Use:   "schema",
		Short: "Generate JSON schema for various objects that the CLI can consume",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateSchema(cmd, options, args)
		},
	}
	return command
}

func generateSchema(cmd *cobra.Command, options *cmd.RootOptions, args []string) error {
	orchSchema, err := schema.Generate(reflect.TypeOf(model.OrchestrationConfigV2{}))
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(orchSchema, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}
