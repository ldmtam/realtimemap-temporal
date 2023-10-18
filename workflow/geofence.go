package workflow

import (
	"context"
	"realtimemap-temporal/data"
	"realtimemap-temporal/shared"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

type GeofenceInput struct {
	Geofence *shared.CircularGeofence
}

type GeofenceOutput struct{}

type GetGeofenceRequest struct{}

type GetGeofenceResponse struct {
	Geofence *shared.Geofence
}

func Geofence(ctx workflow.Context, input *GeofenceInput) (*GeofenceOutput, error) {
	log := workflow.GetLogger(ctx)

	log.Info("Geofence workflow started")
	geofence := input.Geofence
	vehiclesInZone := make(map[string]struct{})

	/*****
		QUERY
	*****/
	err := workflow.SetQueryHandler(ctx, shared.GeofencesQuery, func(request *GetGeofenceRequest) (*GetGeofenceResponse, error) {
		return &GetGeofenceResponse{
			Geofence: &shared.Geofence{
				Name:           geofence.Name,
				RadiusInMeters: geofence.RadiousInMeters,
				Latitude:       geofence.CentralPoint.Lat(),
				Longitude:      geofence.CentralPoint.Lng(),
				VehiclesInZone: getMapKeys(vehiclesInZone),
			},
		}, nil
	})
	if err != nil {
		log.Error("SetQueryHandler failed", "error", err)
		return nil, err
	}

	/*****
		SELECTOR
	*****/
	selector := workflow.NewSelector(ctx)

	selector.AddReceive(workflow.GetSignalChannel(ctx, shared.GeofenceSignal), func(c workflow.ReceiveChannel, more bool) {
		position := &shared.Position{}
		c.Receive(ctx, position)

		_, vehicleIsInZone := vehiclesInZone[position.VehicleId]

		if geofence.IncludesPosition(position.Latitude, position.Longitude) {
			if !vehicleIsInZone {
				vehiclesInZone[position.VehicleId] = struct{}{}
			}
		} else {
			delete(vehiclesInZone, position.VehicleId)
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

	return nil, workflow.NewContinueAsNewError(ctx, Geofence, input)
}

func InitGeofence(ctx context.Context, temporalClient client.Client) error {
	startWorkflowOpts := client.StartWorkflowOptions{
		TaskQueue: shared.RealtimeMapTaskQueue,
	}

	for _, geofence := range data.AllGeofences {
		startWorkflowOpts.ID = GetGeofenceWorkflowID(geofence.Name)
		_, err := temporalClient.ExecuteWorkflow(
			ctx,               // context
			startWorkflowOpts, // start workflow options
			Geofence,          // workflow
			&GeofenceInput{
				Geofence: geofence,
			}, // workflow argument
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func getMapKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
