package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

type BatchRequest struct {
	Requests []BatchSubRequest `json:"requests"`
}

type BatchSubRequest struct {
	ID      string                 `json:"id"`
	Method  string                 `json:"method"`
	URL     string                 `json:"url"`
	Headers map[string]string      `json:"headers,omitempty"`
	Body    map[string]interface{} `json:"body,omitempty"`
}

type BatchResponse struct {
	Responses []BatchSubResponse `json:"responses"`
}

type BatchSubResponse struct {
	ID      string                 `json:"id"`
	Status  int                    `json:"status"`
	Headers map[string]string     `json:"headers,omitempty"`
	Body    map[string]interface{} `json:"body,omitempty"`
}

func batchHandler(engine *gin.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req BatchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
				"code":    "BadRequest",
				"message": "Invalid batch request format.",
			}})
			return
		}

		responses := make([]BatchSubResponse, 0, len(req.Requests))
		for _, sub := range req.Requests {
			// Create internal request
			var bodyBytes []byte
			if sub.Body != nil {
				bodyBytes, _ = json.Marshal(sub.Body)
			}
			req := httptest.NewRequest(sub.Method, sub.URL, bytes.NewReader(bodyBytes))
			// Copy auth header
			if auth := c.GetHeader("Authorization"); auth != "" {
				req.Header.Set("Authorization", auth)
			}
			for k, v := range sub.Headers {
				req.Header.Set(k, v)
			}
			if sub.Body != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			var responseBody map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &responseBody)

			responses = append(responses, BatchSubResponse{
				ID:     sub.ID,
				Status: w.Code,
				Body:   responseBody,
			})
		}

		c.JSON(http.StatusOK, BatchResponse{Responses: responses})
	}
}
