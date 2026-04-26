package handler

import (
	"strings"
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

func getBaseURL(c *gin.Context) string {
	host := c.Request.Host
	if forwarded := c.GetHeader("X-Forwarded-Host"); forwarded != "" {
		host = forwarded
	}
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + host
}

func applySelect(itemMap map[string]interface{}, selects []string) map[string]interface{} {
	if len(selects) == 0 {
		return itemMap
	}
	selectSet := make(map[string]bool, len(selects))
	for _, s := range selects {
		selectSet[s] = true
	}
	result := make(map[string]interface{}, 0)
	for k, v := range itemMap {
		if strings.HasPrefix(k, "@odata.") || selectSet[k] {
			result[k] = v
		}
	}
	return result
}