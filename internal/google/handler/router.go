package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	gauth "github.com/saldeti/saldeti/internal/google/auth"
	"github.com/saldeti/saldeti/internal/google/store"
)

// RegisterRoutes registers all Google Workspace API routes on the given gin engine.
func RegisterRoutes(engine *gin.Engine, st store.Store) {
	// OAuth2 token endpoint (no auth required)
	engine.POST("/oauth2/v1/token", gauth.TokenHandler(st))

	// Admin Directory API
	directory := engine.Group("/admin/directory/v1")
	directory.Use(gauth.RequireAuth())
	{
		directory.GET("", directoryInfoHandler())
		directory.GET("/", directoryInfoHandler())
	}

	// Cloud Identity API: /v1/devices, /v1/groups, etc.
	// Routes will be registered in a future milestone.

	// Admin Reports API: /admin/reports/v1/activity, /admin/reports/v1/usage, etc.
	// Routes will be registered in a future milestone.

	// Admin Data Transfer API: /admin/datatransfer/v1/*
	// Routes will be registered in a future milestone.

	// Workspace Events API: /v1/subscriptions
	// Routes will be registered in a future milestone.
}

func directoryInfoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"kind":      "admin#directory",
			"etag":      "\"placeholder\"",
			"simulator": true,
		})
	}
}