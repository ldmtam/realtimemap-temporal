package workflow

import (
	"context"
	"encoding/json"
	"realtimemap-temporal/shared"
	"time"

	"github.com/redis/go-redis/v9"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

type NotificationInput struct{}

type NotificationOutput struct{}

type NotifyActivities struct {
	RedisCli *redis.Client
}

func Notification(ctx workflow.Context, input *NotificationInput) (*NotificationOutput, error) {
	log := workflow.GetLogger(ctx)

	log.Info("Notification workflow started")

	/*****
		ACTIVITIES
	*****/
	var a *NotifyActivities

	/*****
		SELECTOR
	*****/
	selector := workflow.NewSelector(ctx)

	selector.AddReceive(workflow.GetSignalChannel(ctx, shared.NotificationSignal), func(c workflow.ReceiveChannel, more bool) {
		notification := &shared.Notification{}
		c.Receive(ctx, notification)

		ao := workflow.ActivityOptions{
			TaskQueue:           shared.RealtimeMapTaskQueue,
			StartToCloseTimeout: 10 * time.Second,
		}

		if err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, ao),
			a.Notify,
			notification,
		).Get(ctx, nil); err != nil {
			return
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

	return nil, workflow.NewContinueAsNewError(ctx, Notification, input)
}

func (a *NotifyActivities) Notify(notification *shared.Notification) error {
	notificationBytes, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	err = a.RedisCli.Publish(
		context.Background(),
		shared.GeofenceNotificationChannel,
		string(notificationBytes)).Err()
	if err != nil {
		return err
	}

	return nil
}

func InitNotification(ctx context.Context, temporalClient client.Client) error {
	startWorkflowOpts := client.StartWorkflowOptions{
		TaskQueue: shared.RealtimeMapTaskQueue,
	}

	startWorkflowOpts.ID = GetNotificationWorkflowID()
	_, err := temporalClient.ExecuteWorkflow(
		ctx,                  // context
		startWorkflowOpts,    // start workflow options
		Notification,         // workflow
		&NotificationInput{}, // workflow argument
	)
	if err != nil {
		return err
	}

	return nil
}
