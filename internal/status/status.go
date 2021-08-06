package status

import (
	"context"
	"errors"
	"fmt"
	"github.com/armory/armory-cli/internal/deng/protobuff"
	"github.com/armory/armory-cli/internal/helpers"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/juju/ansiterm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"os"
	"time"
)

const (
	ParameterWatch      = "watch"
	ParameterShowEvents = "show-events"
)

func Execute(ctx context.Context, cmd *cobra.Command, client protobuff.DeploymentServiceClient, args []string) error {
	if len(args) == 0 {
		return errors.New("please provide deployment ID")
	}

	watch, err := cmd.Flags().GetBool(ParameterWatch)
	if err != nil {
		return err
	}

	showEvents, err := cmd.Flags().GetBool(ParameterShowEvents)
	if err != nil {
		return err
	}

	depId := args[0]
	return ShowStatus(ctx, depId, client, watch, showEvents)
}

func ShowStatus(ctx context.Context, deploymentId string, client protobuff.DeploymentServiceClient, watch, showEvents bool) error {
	// Get the status
	// Prepare the request
	req := &protobuff.GetStatusRequest{
		DeploymentId: deploymentId,
	}

	desc, err := client.Status(ctx, req)
	if err != nil {
		log.WithError(err).Error("unable to get deployment information")
		return err
	}

	// Later, results could be sent somewhere else
	w := os.Stdout
	if watch {
		return Watch(ctx, w, desc, showEvents, client)
	}

	// Display the whole status
	printStatus(w, desc)
	if showEvents {
		o := onceEventGetter{
			client:       client,
			deploymentId: deploymentId,
		}
		printEvents(w, o.getEvents)
	}
	return nil
}

func PrintStatus(w io.Writer, descriptor *protobuff.Descriptor) {
	printStatus(w, descriptor)
}

func printStatus(w io.Writer, descriptor *protobuff.Descriptor) {
	wt := ansiterm.NewTabWriter(w, 0, 0, 2, ' ', 0)

	dt := descriptor.StartedAt
	started, err := time.Parse(time.RFC3339, descriptor.StartedAt)
	if err == nil {
		dt = started.Format(time.RFC822)
	}
	_, _ = fmt.Fprintf(wt, "Deployment ID:\t%s\n", descriptor.Id)
	_, _ = fmt.Fprintf(wt, "Started At:\t%s\n", dt)
	_, _ = fmt.Fprintf(wt, "Started By:\t%s\n", descriptor.InitiatedBy)
	_, _ = fmt.Fprintf(wt, "Environment:\t%s (%s)\n", descriptor.Env.Account, descriptor.Env.Provider)
	printKubernetesOptions(wt, descriptor.Env.GetKubernetes())
	_, _ = fmt.Fprintf(wt, "Status:\t%s\n", helpers.Status(descriptor.Status))
	_ = wt.Flush()

	// Kubernetes state
	printKubernetesState(w, descriptor.State)
}

func printEvents(w io.Writer, getter eventGetter) {
	we := ansiterm.NewTabWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprint(we, "\nEVENT\tTYPE\tSTATUS\tDATE\n")

	events, eventsErr := getter(context.TODO())
	if eventsErr != nil {
		_, _ = fmt.Fprintf(w, helpers.AnsiFormat("Error obtaining events\n", helpers.FgRed))
	} else {
		for _, e := range events {
			tm := "?"
			dt, err := time.Parse(time.RFC3339, e.OccurredAt)
			if err == nil {
				tm = dt.Format("15:04:05")
			}
			_, _ = fmt.Fprintf(we, "%s\t%s\t%s\t%s\n", e.Message, e.Type, e.Status, tm)
		}
	}

	_ = we.Flush()
}

func printKubernetesOptions(w io.Writer, qualifier *protobuff.KubernetesQualifier) {
	if qualifier == nil {
		return
	}
	_, _ = fmt.Fprintf(w, "Namespace:\t%s\n", qualifier.Namespace)
}

func printKubernetesState(w io.Writer, state *any.Any) {
	if state == nil {
		return
	}
	d := protobuff.KubernetesDeployment{}
	if err := state.UnmarshalTo(&d); err != nil {
		// Ignore
		return
	}
	if len(d.Atomic) == 0 {
		_, _ = fmt.Fprintf(w, helpers.AnsiFormat("\nNo objects deployed\n", helpers.Bold))
		return
	}

	_, _ = fmt.Fprintf(w, "\nOBJECTS DEPLOYED\n")
	wt := ansiterm.NewTabWriter(w, 0, 0, 2, ' ', 0)
	for _, c := range d.Atomic {
		_, _ = fmt.Fprintf(wt, "Name:\t%s (%s)\n", c.Name, c.Type)
		_, _ = fmt.Fprintf(wt, "Status:\t%s\n", c.Status.String())
		if c.State != nil {
			_, _ = fmt.Fprintf(wt, "Replicas:\t%d\n", c.State.Replicas)
		}
	}
	_ = wt.Flush()
}

func Watch(ctx context.Context, w io.Writer, descriptor *protobuff.Descriptor, showEvents bool, client protobuff.DeploymentServiceClient) error {
	// TODO make it a little smarter than a timer
	// We don't need to check every 5s
	timer := time.NewTimer(5 * time.Second)
	wg := watchGetter{client: client, descriptor: descriptor}
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			// TODO don't ignore error
			_ = wg.tick(ctx)
			helpers.Clear(w)
			printStatus(w, descriptor)
			if showEvents {
				printEvents(w, wg.getEvents)
			}
			if wg.descriptor.Status.IsFinal() {
				// We're done - stop
				return nil
			}
		}
	}
}
