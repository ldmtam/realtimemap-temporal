package main

import (
	"os"

	"log/slog"

	"realtimemap-temporal/shared"
	"realtimemap-temporal/workflow"

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

	w := worker.New(temporalClient, shared.RealtimeMapTaskQueue, worker.Options{})

	w.RegisterWorkflow(workflow.Vehicle)
	w.RegisterWorkflow(workflow.Organization)
	w.RegisterWorkflow(workflow.Geofence)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		panic(err)
	}
}
