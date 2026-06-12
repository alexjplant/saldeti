package handler

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/saldeti/saldeti/internal/entra/model"
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

func applySelect(itemMap map[string]interface{}, selects []string, extraPreserve ...map[string]bool) map[string]interface{} {
	if len(selects) == 0 {
		return itemMap
	}
	selectSet := make(map[string]bool, len(selects))
	for _, s := range selects {
		selectSet[s] = true
	}
	// Merge any extra keys to preserve (e.g. $expand properties)
	if len(extraPreserve) > 0 {
		for k := range extraPreserve[0] {
			selectSet[k] = true
		}
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

// buildEntityResponse creates a response map with @odata.context and merges in
// all fields from the entity by marshaling it to JSON and back.
func buildEntityResponse(odataCtx string, entity any) (map[string]interface{}, error) {
	response := map[string]interface{}{
		"@odata.context": odataCtx,
	}
	entityJSON, err := json.Marshal(entity)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal entity: %w", err)
	}
	var entityMap map[string]interface{}
	if err := json.Unmarshal(entityJSON, &entityMap); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal entity: %w", err)
	}
	for k, v := range entityMap {
		response[k] = v
	}
	return response, nil
}

// buildEntityResponseWithType creates a response map with @odata.context and
// @odata.type, then merges in all fields from the entity.
func buildEntityResponseWithType(odataCtx string, odataType string, entity any) (map[string]interface{}, error) {
	response := map[string]interface{}{
		"@odata.context": odataCtx,
		"@odata.type":    odataType,
	}
	entityJSON, err := json.Marshal(entity)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal entity: %w", err)
	}
	var entityMap map[string]interface{}
	if err := json.Unmarshal(entityJSON, &entityMap); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal entity: %w", err)
	}
	for k, v := range entityMap {
		response[k] = v
	}
	return response, nil
}

// applyNestedSelectToDirectoryObjects applies nested $select from an ExpandOption
// to a slice of DirectoryObject. Returns the raw slice if no select fields are specified,
// or a slice of filtered maps if select fields are present.
func applyNestedSelectToDirectoryObjects(objects []model.DirectoryObject, selectFields []string) interface{} {
	if len(selectFields) == 0 {
		return objects
	}
	maps := make([]map[string]interface{}, 0, len(objects))
	for _, obj := range objects {
		objJSON, err := json.Marshal(obj)
		if err != nil {
			maps = append(maps, map[string]interface{}{})
			continue
		}
		var m map[string]interface{}
		json.Unmarshal(objJSON, &m)
		m = applySelect(m, selectFields)
		maps = append(maps, m)
	}
	return maps
}

// applyNestedSelectToUser applies nested $select from an ExpandOption
// to a User object. Serializes the user to a map, applies select filtering,
// and ensures @odata.type is set to #microsoft.graph.user.
func applyNestedSelectToUser(user *model.User, selectFields []string) map[string]interface{} {
	userJSON, _ := json.Marshal(user)
	var m map[string]interface{}
	json.Unmarshal(userJSON, &m)
	if len(selectFields) > 0 {
		m = applySelect(m, selectFields)
	}
	m["@odata.type"] = "#microsoft.graph.user"
	return m
}

// nilToEmptyDirectoryObjects replaces nil slice with empty slice.
func nilToEmptyDirectoryObjects(objects []model.DirectoryObject) []model.DirectoryObject {
	if objects == nil {
		return []model.DirectoryObject{}
	}
	return objects
}

// computeExpandedPropertyNames returns a set of expanded property names for preserving
// them when $select is also applied.
func computeExpandedPropertyNames(expandOptions []model.ExpandOption) map[string]bool {
	expanded := make(map[string]bool)
	for _, eo := range expandOptions {
		expanded[eo.Property] = true
	}
	return expanded
}