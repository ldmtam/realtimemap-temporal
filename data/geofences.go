package data

import (
	"realtimemap-temporal/shared"

	geo "github.com/kellydunn/golang-geo"
)

var AllGeofences = map[string]*shared.CircularGeofence{
	"Airport":           Airport,
	"Downtown":          Downtown,
	"RailwaySquare":     RailwaySquare,
	"LauttasaariIsland": LauttasaariIsland,
	"LaajasaloIsland":   LaajasaloIsland,
	"KallioDistrict":    KallioDistrict,
}

var (
	Airport = &shared.CircularGeofence{
		Name:            "Airport",
		CentralPoint:    *geo.NewPoint(60.31146, 24.96907),
		RadiousInMeters: 2000,
	}

	Downtown = &shared.CircularGeofence{
		Name:            "Downtown",
		CentralPoint:    *geo.NewPoint(60.16422983026082, 24.941068845053014),
		RadiousInMeters: 1700,
	}

	RailwaySquare = &shared.CircularGeofence{
		Name:            "Railway Square",
		CentralPoint:    *geo.NewPoint(60.171285, 24.943936),
		RadiousInMeters: 150,
	}

	LauttasaariIsland = &shared.CircularGeofence{
		Name:            "Lauttasaari island",
		CentralPoint:    *geo.NewPoint(60.158536, 24.873788),
		RadiousInMeters: 1400,
	}

	LaajasaloIsland = &shared.CircularGeofence{
		Name:            "Laajasalo island",
		CentralPoint:    *geo.NewPoint(60.16956184470527, 25.052851825093114),
		RadiousInMeters: 2200,
	}

	KallioDistrict = &shared.CircularGeofence{
		Name:            "Kallio district",
		CentralPoint:    *geo.NewPoint(60.18260263288996, 24.953588638997264),
		RadiousInMeters: 600,
	}
)
