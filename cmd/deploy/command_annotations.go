package deploy

import "github.com/spf13/cobra"

func storeCommandResult(cmd *cobra.Command, key, value string) {

	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}

	cmd.Annotations[key] = value
}

func fetchCommandResult(cmd *cobra.Command, key string) string {
	if cmd.Annotations == nil {
		return ""
	}
	return cmd.Annotations[key]
}
