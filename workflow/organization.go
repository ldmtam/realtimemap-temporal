package workflow

import (
	"context"
	"realtimemap-temporal/data"
	"realtimemap-temporal/shared"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

type OrganizationInput struct {
	Id        string
	Name      string
	Geofences []*shared.CircularGeofence
}

type OrganizationOutput struct{}

func Organization(ctx workflow.Context, input *OrganizationInput) (*OrganizationOutput, error) {
	log := workflow.GetLogger(ctx)

	log.Info("Organization workflow started")

	geofences := input.Geofences

	for {
		selector := workflow.NewSelector(ctx)

		selector.AddReceive(workflow.GetSignalChannel(ctx, shared.OrganizationSignal), func(c workflow.ReceiveChannel, more bool) {
			position := &shared.Position{}
			c.Receive(ctx, position)

			for _, geofence := range geofences {
				workflow.SignalExternalWorkflow(
					ctx,                                  // context
					GetGeofenceWorkflowID(geofence.Name), // workflow id
					"",                                   // run id
					shared.GeofenceSignal,                // signal name
					position,                             // signal argument
				)
			}
		})

		selector.Select(ctx)
	}
}

func InitOrganization(ctx context.Context, temporalClient client.Client) error {
	startWorkflowOpts := client.StartWorkflowOptions{
		TaskQueue: shared.RealtimeMapTaskQueue,
	}

	for _, org := range data.AllOrganizations {
		startWorkflowOpts.ID = GetOrganizationWorkflowID(org.Id)
		_, err := temporalClient.ExecuteWorkflow(
			ctx,               // context
			startWorkflowOpts, // start workflow options
			Organization,      // workflow
			&OrganizationInput{
				Id:        org.Id,
				Name:      org.Name,
				Geofences: org.Geofences,
			}, // workflow argument
		)
		if err != nil {
			return err
		}
	}

	return nil
}
