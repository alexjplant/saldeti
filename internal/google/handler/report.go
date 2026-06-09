package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/google/model"
	"github.com/saldeti/saldeti/internal/google/store"
)

// listActivitiesHandler handles GET /admin/reports/v1/activity/users/:userKey/applications/:applicationName
func listActivitiesHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := c.Param("userKey")
		applicationName := c.Param("applicationName")
		activities, err := st.ListActivities(c.Request.Context(), userKey, applicationName)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to list activities")
			return
		}
		if activities == nil {
			activities = []model.Activity{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"kind":  "admin#reports#activities",
			"etag":  "\"placeholder\"",
			"items": activities,
		})
	}
}

// watchActivitiesHandler handles POST /admin/reports/v1/activity/users/:userKey/applications/:applicationName/watch
func watchActivitiesHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		writeJSON(c, http.StatusOK, gin.H{
			"kind": "admin#reports#channel",
			"id":   "",
			"etag": "\"placeholder\"",
		})
	}
}

// getCustomerUsageReportHandler handles GET /admin/reports/v1/usage/dates/:date
func getCustomerUsageReportHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		date := c.Param("date")
		reports, err := st.ListUsageReports(c.Request.Context(), date, "", "", "")
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to get customer usage report")
			return
		}
		if reports == nil {
			reports = []model.UsageReport{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"kind":  "admin#reports#usageReports",
			"etag":  "\"placeholder\"",
			"usageReports": reports,
		})
	}
}

// getUserUsageReportHandler handles GET /admin/reports/v1/usage/users/:userKey/dates/:date
func getUserUsageReportHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		date := c.Param("date")
		userKey := c.Param("userKey")
		reports, err := st.ListUsageReports(c.Request.Context(), date, userKey, "", "")
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to get user usage report")
			return
		}
		if reports == nil {
			reports = []model.UsageReport{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"kind":  "admin#reports#usageReports",
			"etag":  "\"placeholder\"",
			"usageReports": reports,
		})
	}
}

// getEntityUsageReportHandler handles GET /admin/reports/v1/usage/:entityType/:entityKey/dates/:date
func getEntityUsageReportHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		date := c.Param("date")
		entityType := c.Param("entityType")
		entityKey := c.Param("entityKey")
		reports, err := st.ListUsageReports(c.Request.Context(), date, "", entityType, entityKey)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to get entity usage report")
			return
		}
		if reports == nil {
			reports = []model.UsageReport{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"kind":  "admin#reports#usageReports",
			"etag":  "\"placeholder\"",
			"usageReports": reports,
		})
	}
}