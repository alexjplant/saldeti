package handler

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const maxBodyBytes int64 = 1 << 20 // 1MB

var configuredBaseURL string

func SetBaseURL(url string) {
	configuredBaseURL = url
}

var trustForwardedHeaders bool

func SetTrustForwardedHeaders(trust bool) {
	trustForwardedHeaders = trust
}

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
	if configuredBaseURL != "" {
		return configuredBaseURL
	}
	host := c.Request.Host
	if trustForwardedHeaders {
		if forwarded := c.GetHeader("X-Forwarded-Host"); forwarded != "" {
			host = forwarded
		}
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if trustForwardedHeaders {
		if c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
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
	// Always include 'id' per MS Graph API behavior
	if val, ok := itemMap["id"]; ok {
		result["id"] = val
	}
	return result
}

func isFilterError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "unable to parse filter expression") ||
		strings.Contains(errStr, "cannot compare values") ||
		strings.Contains(errStr, "operator not supported") ||
		strings.Contains(errStr, "function value must be string") ||
		strings.Contains(errStr, "unknown function") ||
		strings.Contains(errStr, "invalid filter node")
}