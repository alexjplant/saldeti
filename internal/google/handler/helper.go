package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/google/model"
)

const maxBodyBytes int64 = 1 << 20 // 1MB

func writeJSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}

// writeError writes a Google-style error response.
// status = HTTP status code (also used as error.code)
// reason = Google error reason string (e.g. "notFound", "duplicate", "invalid", "backendError")
// message = human-readable message
func writeError(c *gin.Context, status int, reason string, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    status,
			"message": message,
			"errors": []gin.H{
				{
					"domain":  "global",
					"reason":  reason,
					"message": message,
				},
			},
		},
	})
}

// parseGoogleListOptions extracts Google API list query parameters from the request.
func parseGoogleListOptions(c *gin.Context) model.ListOptions {
	opts := model.ListOptions{}
	if mr := c.Query("maxResults"); mr != "" {
		if v, err := strconv.Atoi(mr); err == nil {
			opts.MaxResults = v
		}
	}
	opts.PageToken = c.Query("pageToken")
	opts.Query = c.Query("query")
	opts.OrderBy = c.Query("orderBy")
	opts.SortOrder = c.Query("sortOrder")
	opts.Customer = c.Query("customer")
	opts.Domain = c.Query("domain")
	opts.Projection = c.Query("projection")
	opts.ViewType = c.Query("viewType")
	return opts
}
