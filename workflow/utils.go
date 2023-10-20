package workflow

import "fmt"

func GetVehicleWorkflowID(vehicleID string) string {
	return fmt.Sprintf("vehicle-%v", vehicleID)
}

func GetOrganizationWorkflowID(orgID string) string {
	return fmt.Sprintf("organization-%v", orgID)
}

func GetGeofenceWorkflowID(name string) string {
	return fmt.Sprintf("geofence-%v", name)
}

func GetNotificationWorkflowID() string {
	return "notification"
}
