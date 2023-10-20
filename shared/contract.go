package shared

import geo "github.com/kellydunn/golang-geo"

type Position struct {
	VehicleId string  `json:"vehicleId"`
	OrgId     string  `json:"orgId"`
	OrgName   string  `json:"orgName"`
	Timestamp int64   `json:"timestamp"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Heading   int32   `json:"heading"`
	DoorsOpen bool    `json:"doorsOpen"`
	Speed     float64 `json:"speed"`
}

type PositionBatch struct {
	Positions []*Position `json:"positions"`
}

type Organization struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type OrganizationDetails struct {
	Id        string      `json:"id"`
	Name      string      `json:"name"`
	Geofences []*Geofence `json:"geofences"`
}

type Geofence struct {
	Name           string   `json:"name"`
	Longitude      float64  `json:"longitude"`
	Latitude       float64  `json:"latitude"`
	RadiusInMeters float64  `json:"radiusInMeters"`
	VehiclesInZone []string `json:"vehiclesInZone"`
}

type Notification struct {
	VehicleId string `json:"vehicleId"`
	OrgId     string `json:"orgId"`
	OrgName   string `json:"orgName"`
	ZoneName  string `json:"zoneName"`
	Event     string `json:"event"`
}

type CircularGeofence struct {
	Name            string
	CentralPoint    geo.Point
	RadiousInMeters float64
}

func (geofence *CircularGeofence) IncludesPosition(latitude float64, longitude float64) bool {
	point := geo.NewPoint(latitude, longitude)
	return geofence.CentralPoint.GreatCircleDistance(point)*1000 < geofence.RadiousInMeters
}
