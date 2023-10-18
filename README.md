# Real-time Map

Real-time Map displays real-time positions of public transport vehicles in Helsinki. It's a showcase for using Temporal Workflow as Actors

The app features:
- Real-time positions of vehicles.
- Vehicle trails.
- Geofencing notifications (vehicle entering and exiting the area).
- Vehicles in geofencing areas per public transport company.
- Horizontal scaling.

The goals of this app are:
- Showing what Temporal Workflow can do.
- Presenting a semi-real-world use case of the distributed actor model.
- Learning how to use Temporal Workflow

## Running the app
Prerequisites:
- [Golang](https://go.dev/)
- [Temporal Workflow](https://learn.temporal.io/getting_started/go/dev_environment/)

Run Temporal Workflow
```
temporal server start-dev
```
Start worker
```
go run worker/main.go
```

Start backend
```
go run main.go
```

Check out the Temporal Workflow UI by navigating to [localhost:8233](http://localhost:8233)

## What does it work?
We'll have 3 types of Workflow in the application
- Vehicle: receive position update message from MQTT, send signal the **organization** Workflow, maintain vehicle position history and response to **get vehicle history request** from **server**
- Organization: receive signal from **vehicle** Workflow and send signal to corresponding **geofence** Workflow
- Geofence: receive signal from **organization** Workflow, maintain which vehicles are currently in this geofence and response to **get geofence request** from **server**

## cURL
List all **organizations** that have geofences setup
```
curl --location 'localhost:12345/api/v1/organization'
```

List all **geofences** and **vehicles** currently inside those geofences of an **organization**
```
curl --location 'localhost:12345/api/v1/organization/0030'
```

List all position changes history of an **vehicle**
```
curl --location 'localhost:12345/api/v1/trail/0012.02212'
```

## How does it work?
Please refer to the [.NET version using Proto.Actor](https://github.com/asynkron/realtimemap-dotnet) README for a detailed description of the architecture.

## TODO
-  [ ] Support Geofencing notifications (vehicle entering and exiting the area).
-  [x] Support Continue-As-New for Workflows when Temporal history events length and size limit reached.
