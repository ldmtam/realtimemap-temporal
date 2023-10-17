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

	for {
		selector := workflow.NewSelector(ctx)

		selector.AddReceive(workflow.GetSignalChannel(ctx, shared.GeofenceSignal), func(c workflow.ReceiveChannel, more bool) {
			position := &shared.Position{}
			c.Receive(ctx, position)

			_, vehicleIsInZone := vehiclesInZone[position.VehicleId]

			if input.Geofence.IncludesPosition(position.Latitude, position.Longitude) {
				if !vehicleIsInZone {
					vehiclesInZone[position.VehicleId] = struct{}{}
				}
			} else {
				delete(vehiclesInZone, position.VehicleId)
			}

		})

		selector.Select(ctx)
	}
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
	keys := make([]string, len(m))

	i := 0
	for k := range m {
		keys[i] = k
		i++
	}

	return keys
}
