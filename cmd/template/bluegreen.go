package template

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

const (
	templateBlueGreenShort   = "Generate a bluegreen deployment template"
	templateBlueGreenLong    = "Generate a bluegreen deployment template in YAML format"
	templateBlueGreenExample = "armory template bluegreen > bluegreen.yaml"
)

const blueGreenTemplate = `
# A map of deployment strategies, keyed by name.
strategies:

  # Strategy name. Use a descriptive name (e.g., "prod-strategy").
  # Use a strategy by assigning it to a deployment target above.
  <strategy>:

    # Define a blue/green deployment strategy.
    #
    # When using a blue/green strategy, only one version of your software
    # gets exposed to users at a time.
    #
    # First, Borealis deploys the new version without
    # exposing it to the activeService defined below. The new version is
    # then accessible using the previewService (if defined).
    #
    # Second, Borealis executes the "redirectTrafficAfter" steps in parallel.
    # After each step completes, Borealis exposes the new version
    # to the activeService.
    #
    # Finally, Borealis executes the "shutDownOldVersion" steps in parallel.
    # After each step completes, Borealis deletes the old version.
    blueGreen:

      # The name of a Kubernetes Service resource.
      # The activeService must be deployed out-of-band and should be configured
      # to direct traffic to your application.
      activeService: active-service

      # The name of a Kubernetes Service resource. Optional.
      # The previewService must be deployed out-of-band and should be configured
      # to direct traffic to your application. You can use this service to
      # preview the new version of your application before it is exposed to users.
      previewService: preview-service

      # The redirectTrafficAfter steps are pre-conditions for exposing the new
      # version to the activeService. The steps are executed
      # in parallel.
      redirectTrafficAfter:

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

        # A pause step type.
        # The pause type defined below is a judgment-based pause.
        - pause:

            # Pause the deployment until manual approval.
            # You can approve or rollback a deployment in the Cloud Console.
            # Do not provide a "duration" or "unit" value when defining
            # a judgment-based pause.
            untilApproved: true

      # The shutDownOldVersionAfter steps are pre-conditions for deleting the old
      # version of your software. The steps are executed in parallel.
      shutdownOldVersionAfter:
        - pause:
            untilApproved: true
`

type templateBlueGreenOptions struct {
	*templateOptions
}

func NewTemplateBlueGreenCmd(templateOptions *templateOptions) *cobra.Command {
	options := &templateBlueGreenOptions{
		templateOptions: templateOptions,
	}
	cmd := &cobra.Command{
		Use:     "bluegreen",
		Aliases: []string{"bluegreen"},
		Short:   templateBlueGreenShort,
		Long:    templateBlueGreenLong,
		Example: templateBlueGreenExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return blueGreen(cmd, options, args)
		},
	}
	return cmd
}

func blueGreen(cmd *cobra.Command, options *templateBlueGreenOptions, args []string) error {
	template := strings.Join([]string{KubernetesCoreTemplate, blueGreenTemplate},"\n")
	_, err := cmd.OutOrStdout().Write([]byte(template))
	if err != nil {
		return fmt.Errorf("error trying to parse bluegreen template: %s", err)
	}
	return nil
}
