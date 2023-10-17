package main

import (
	"context"
	"os"
	"os/signal"
	"realtimemap-temporal/ingress"
	"realtimemap-temporal/server"
	"realtimemap-temporal/shared"
	"realtimemap-temporal/workflow"

	"log/slog"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	stopOnSignals(cancel)

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

	srv := server.NewHttpServer(ctx, temporalClient)
	srvDone := srv.ListenAndServe()

	err = workflow.InitOrganization(ctx, temporalClient)
	if err != nil {
		panic(err)
	}

	err = workflow.InitGeofence(ctx, temporalClient)
	if err != nil {
		panic(err)
	}

	ingressDone := ingress.ConsumeVehicleEvents(func(e *ingress.Event) {
		position := mapToPosition(e)
		if position != nil {
			err := workflow.InitVehicle(ctx, temporalClient, position)
			if err != nil {
				panic(err)
			}
		}
	}, ctx)

	<-ingressDone
	<-srvDone
}

func stopOnSignals(cancel func()) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	go func() {
		<-sigs
		slog.Info("*** STOPPING ***")
		cancel()
	}()
}

func mapToPosition(e *ingress.Event) *shared.Position {
	var payload *ingress.Payload

	if e.VehiclePosition != nil {
		payload = e.VehiclePosition
	} else if e.DoorOpen != nil {
		payload = e.DoorOpen
	} else if e.DoorClosed != nil {
		payload = e.DoorClosed
	} else {
		return nil
	}

	if !payload.HasValidPosition() {
		return nil
	}

	return &shared.Position{
		VehicleId: e.VehicleId,
		OrgId:     e.OperatorId,
		Latitude:  *payload.Latitude,
		Longitude: *payload.Longitude,
		Heading:   *payload.Heading,
		Timestamp: (*payload.Timestamp).UnixMilli(),
		Speed:     *payload.Speed,
	}
}
