package main

import (
	"os"

	"log/slog"

	"realtimemap-temporal/shared"
	"realtimemap-temporal/workflow"

	"github.com/redis/go-redis/v9"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"
)

func main() {
	clientOptions := client.Options{
		Logger: log.NewStructuredLogger(
			slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
			}))),
	}
	temporalClient, err := client.Dial(clientOptions)
	if err != nil {
		panic(err)
	}
	defer temporalClient.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: ":6379",
	})
	defer redisClient.Close()

	w := worker.New(temporalClient, shared.RealtimeMapTaskQueue, worker.Options{})

	w.RegisterWorkflow(workflow.Vehicle)
	w.RegisterWorkflow(workflow.Organization)
	w.RegisterWorkflow(workflow.Geofence)
	w.RegisterWorkflow(workflow.Notification)

	notifyActivities := &workflow.NotifyActivities{
		RedisCli: redisClient,
	}
	w.RegisterActivity(notifyActivities)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		panic(err)
	}
}
