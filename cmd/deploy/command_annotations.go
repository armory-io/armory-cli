package deploy

import "github.com/spf13/cobra"

// storeCommandResult - commands contain generic purpose Annotations for communicating between chain of
// commands - we can use it to pass simple data around from child to parent commands without need to use
// command's context. storeCommandResult will typically be invoked from child command and will store
// a value with given key in the store.
func storeCommandResult(cmd *cobra.Command, key, value string) {

	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}

	cmd.Annotations[key] = value
}

// fetchCommandResult - typically parent command will fetch value from the store written there by child command(s)
func fetchCommandResult(cmd *cobra.Command, key string) string {
	if cmd.Annotations == nil {
		return ""
	}
	return cmd.Annotations[key]
}
