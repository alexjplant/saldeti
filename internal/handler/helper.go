package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func writeJSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}

func writeError(c *gin.Context, status int, code string, message string) {
	requestID := uuid.New().String()
	clientRequestID := c.GetHeader("client-request-id")
	if clientRequestID == "" {
		clientRequestID = requestID
	}

	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
			"innerError": gin.H{
				"date":             time.Now().Format(time.RFC3339),
				"request-id":       requestID,
				"client-request-id": clientRequestID,
			},
		},
	})
}