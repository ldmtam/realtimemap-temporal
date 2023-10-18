package workflow

import (
	"context"
	"fmt"
	"realtimemap-temporal/shared"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

const MaxPositionHistory = 200

type VehicleInput struct{}

type VehicleOutput struct{}

type GetPositionHistoryRequest struct{}

type GetPositionHistoryResponse struct {
	Positions *shared.PositionBatch
}

func Vehicle(ctx workflow.Context, input *VehicleInput) (*VehicleOutput, error) {
	log := workflow.GetLogger(ctx)

	log.Info("Vehicle workflow started")
	positionHistory := make([]*shared.Position, 0)

	/*****
		QUERY
	*****/
	err := workflow.SetQueryHandler(ctx, shared.VehiclePositionHistoryQuery, func(request *GetPositionHistoryRequest) (*GetPositionHistoryResponse, error) {
		return &GetPositionHistoryResponse{
			Positions: &shared.PositionBatch{Positions: positionHistory},
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

	selector.AddReceive(workflow.GetSignalChannel(ctx, shared.VehicleSignal), func(c workflow.ReceiveChannel, more bool) {
		position := &shared.Position{}
		c.Receive(ctx, position)

		if len(positionHistory) > MaxPositionHistory {
			positionHistory = positionHistory[1:]
		}
		positionHistory = append(positionHistory, position)

		workflow.SignalExternalWorkflow(
			ctx, // context
			GetOrganizationWorkflowID(position.OrgId), // workflow id
			"",                        // run id
			shared.OrganizationSignal, // signal name
			position,                  // signal argument
		)
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

	return nil, workflow.NewContinueAsNewError(ctx, Vehicle, input)
}

func InitVehicle(ctx context.Context, temporalClient client.Client, position *shared.Position) error {
	workflowID := fmt.Sprintf("vehicle-%v", position.VehicleId)
	startWorkflowOpts := client.StartWorkflowOptions{
		TaskQueue: shared.RealtimeMapTaskQueue,
	}

	_, err := temporalClient.SignalWithStartWorkflow(
		ctx,                  // context
		workflowID,           // workflow id
		shared.VehicleSignal, // signal name
		position,             // signal argument
		startWorkflowOpts,    // start workflow options
		Vehicle,              // workflow
		&VehicleInput{},      // workflow arguments
	)
	if err != nil {
		return err
	}

	return nil
}
