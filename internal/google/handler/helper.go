package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const maxBodyBytes int64 = 1 << 20 // 1MB

func writeJSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}

func writeError(c *gin.Context, status int, code string, message string) {
	requestID := uuid.New().String()
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
			"details": gin.H{
				"requestId": requestID,
			},
		},
	})
}