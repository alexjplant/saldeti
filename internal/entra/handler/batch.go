package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

const maxBatchRequests = 20 // Maximum number of individual requests allowed in a single $batch

type BatchRequest struct {
	Requests []BatchSubRequest `json:"requests"`
}

type BatchSubRequest struct {
	ID      string            `json:"id"`
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    map[string]any    `json:"body,omitempty"`
}

type BatchResponse struct {
	Responses []BatchSubResponse `json:"responses"`
}

type BatchSubResponse struct {
	ID      string            `json:"id"`
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    map[string]any    `json:"body,omitempty"`
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

		if len(req.Requests) > maxBatchRequests {
			writeError(c, http.StatusBadRequest, "BadRequest", fmt.Sprintf("Batch request exceeds maximum of %d individual requests", maxBatchRequests))
			return
		}

		responses := processBatchSubRequests(c, engine, req.Requests)
		c.JSON(http.StatusOK, BatchResponse{Responses: responses})
	}
}

// processBatchSubRequests executes each sub-request against the engine and
// returns the collected sub-responses. Sub-requests whose body cannot be
// marshaled to JSON receive a 500 error sub-response and are skipped.
func processBatchSubRequests(c *gin.Context, engine *gin.Engine, requests []BatchSubRequest) []BatchSubResponse {
	responses := make([]BatchSubResponse, 0, len(requests))
	for _, sub := range requests {
		var bodyBytes []byte
		if sub.Body != nil {
			var err error
			bodyBytes, err = json.Marshal(sub.Body)
			if err != nil {
				log.Error().
					Err(err).
					Str("sub_id", sub.ID).
					Str("sub_url", sub.URL).
					Msg("Failed to marshal batch sub-request body")
				responses = append(responses, BatchSubResponse{
					ID:     sub.ID,
					Status: http.StatusInternalServerError,
					Body: map[string]any{
						"error": gin.H{
							"code":    "InternalError",
							"message": "Failed to marshal request body",
						},
					},
				})
				continue
			}
		}

		url := normalizeBatchURL(sub.URL)
		req := httptest.NewRequest(sub.Method, url, bytes.NewReader(bodyBytes))
		req.Host = c.Request.Host
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

		var responseBody map[string]any
		if w.Body.Len() > 0 {
			if err := json.Unmarshal(w.Body.Bytes(), &responseBody); err != nil {
				log.Error().
					Err(err).
					Str("sub_id", sub.ID).
					Str("sub_url", sub.URL).
					Int("status", w.Code).
					Msg("Failed to unmarshal batch sub-response body")
				responseBody = map[string]any{
					"error": gin.H{
						"code":    "InternalError",
						"message": "Failed to parse sub-response body",
					},
				}
			}
		}

		responses = append(responses, BatchSubResponse{
			ID:     sub.ID,
			Status: w.Code,
			Body:   responseBody,
		})
	}
	return responses
}

func normalizeBatchURL(url string) string {
	if strings.HasPrefix(url, "/v1.0/") || strings.HasPrefix(url, "/beta/") {
		return url
	}
	if !strings.HasPrefix(url, "/") {
		return "/v1.0/" + url
	}
	return "/v1.0" + url
}
