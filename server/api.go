package server

import (
	"fmt"
	"net/http"
	"realtimemap-temporal/data"
	"realtimemap-temporal/shared"
	"realtimemap-temporal/workflow"
	"sort"

	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/client"
)

func serveAPI(router *gin.Engine, temporalClient client.Client) {
	router.GET("/api/v1/organization", func(c *gin.Context) {
		result := make([]*shared.Organization, 0, len(data.AllOrganizations))

		for _, org := range data.AllOrganizations {
			if len(org.Geofences) > 0 {
				result = append(result, &shared.Organization{
					Id:   org.Id,
					Name: org.Name,
				})
			}
		}

		sort.Slice(result, func(i, j int) bool {
			return result[i].Name < result[j].Name
		})

		c.JSON(http.StatusOK, result)
	})

	router.GET("/api/v1/organization/:id", func(c *gin.Context) {
		orgID := c.Param("id")
		org, ok := data.AllOrganizations[orgID]
		if !ok {
			c.JSON(http.StatusNotFound, map[string]any{"message": fmt.Sprintf("Organization %v not found", orgID)})
			return
		}

		geofenceSet := make(map[string]struct{})
		for _, geofence := range org.Geofences {
			geofenceSet[geofence.Name] = struct{}{}
		}

		geofences := make([]*shared.Geofence, 0)
		for geofence := range geofenceSet {
			resp, err := temporalClient.QueryWorkflow(
				c.Request.Context(),                      // context
				workflow.GetGeofenceWorkflowID(geofence), // workflow id
				"",                                       // run id
				shared.GeofencesQuery,                    // query type
				&workflow.GetGeofenceRequest{},           // query input
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, map[string]any{"error": err})
				return
			}

			geofenceResp := &workflow.GetGeofenceResponse{}
			err = resp.Get(geofenceResp)
			if err != nil {
				c.JSON(http.StatusInternalServerError, map[string]error{"error": err})
				return
			}

			geofences = append(geofences, geofenceResp.Geofence)
		}

		sort.Slice(geofences, func(i, j int) bool {
			return geofences[i].Name < geofences[j].Name
		})

		c.JSON(http.StatusOK, &shared.OrganizationDetails{
			Id:        org.Id,
			Name:      org.Name,
			Geofences: geofences,
		})
	})

	router.GET("/api/v1/trail/:id", func(c *gin.Context) {
		vehicleID := c.Param("id")

		resp, err := temporalClient.QueryWorkflow(
			c.Request.Context(),                      // context
			workflow.GetVehicleWorkflowID(vehicleID), // workflow id
			"",                                       // run id
			shared.VehiclePositionHistoryQuery,       // query type
			&workflow.GetPositionHistoryRequest{},    // query input
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err})
			return
		}

		historyResp := &workflow.GetPositionHistoryResponse{}
		err = resp.Get(historyResp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err})
			return
		}

		c.JSON(http.StatusOK, historyResp.Positions)
	})
}
