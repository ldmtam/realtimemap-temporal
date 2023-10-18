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

	/*****
		SELECTOR
	*****/
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

	for {
		selector.Select(ctx)
		// we'll continue this workflow as new one when reaching history length and size limit
		if workflow.GetInfo(ctx).GetContinueAsNewSuggested() {
			// after draining all events
			if !selector.HasPending() {
				break
			}
		}
		// if you want to test the logic of continuing workflow as new, please change the condition to
		// "workflow.GetInfo(ctx).GetCurrentHistoryLength() > 100", it'll create new workflow
		// when history length is at least 100
	}

	return nil, workflow.NewContinueAsNewError(ctx, Organization, input)
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
