package template

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

const (
	templateCanaryShort   = "Generate a canary deployment template"
	templateCanaryLong    = "Generate a canary deployment template in YAML format"
	templateCanaryExample = "armory template canary > canary.yaml"
)

const canaryTemplate = `
# A map of deployment strategies, keyed by name.
strategies:

  # Strategy name. Use a descriptive name (e.g., "prod-strategy").
  # Use a strategy by assigning it to a deployment target above.
  <strategy>:

    # Define a progressive canary deployment strategy.
    # Use this strategy to incrementally shift traffic to a new version of
    # your software.
    canary:

      # A set of deployment steps. The steps are executed in order.
      steps:

        # A pause step type.
        # The deployment stops until the pause behavior is complete.
        # The pause type defined below is a duration-based pause.
        - pause:

            # Pause the deployment for <duration> <unit> (e.g., pause for 5 minutes).
            # A duration-based pause should omit the "untilApproved" flag.
            duration: 1

            # The pause's time unit. One of seconds, minutes, or hours.
            # Required if duration is set.
            unit: seconds

        # The percentage of pods that should be running the new version
        # of your software.
        # Use an integer between 0 and 100, inclusive.
        - setWeight:
            weight: 33

        # A pause step type.
        # The pause type defined below is a judgment-based pause.
        - pause:

            # Pause the deployment until manual approval.
            # You can approve or rollback a deployment in the Cloud Console.
            # Do not provide a "duration" or "unit" value when defining
            # a judgment-based pause.
            untilApproved: true
`

type templateCanaryOptions struct {
	*templateOptions
}

func NewTemplateCanaryCmd(templateOptions *templateOptions) *cobra.Command {
	options := &templateCanaryOptions{
		templateOptions: templateOptions,
	}
	cmd := &cobra.Command{
		Use:     "canary",
		Aliases: []string{"canary"},
		Short:   templateCanaryShort,
		Long:    templateCanaryLong,
		Example: templateCanaryExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return canary(cmd, options, args)
		},
	}
	return cmd
}

func canary(cmd *cobra.Command, options *templateCanaryOptions, args []string) error {
	template := strings.Join([]string{KubernetesCoreTemplate, canaryTemplate},"\n")
	_, err := cmd.OutOrStdout().Write([]byte(template))
	if err != nil {
		return fmt.Errorf("error trying to parse canary template: %s", err)
	}
	return nil
}
