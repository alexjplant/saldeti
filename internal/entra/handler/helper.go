package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/saldeti/saldeti/internal/entra/model"
	"github.com/saldeti/saldeti/internal/entra/store"
)

const maxBodyBytes int64 = 1 << 20 // 1MB
const maxTopValue = 999            // Maximum $top value accepted by the MS Graph API; also used for $expand to fetch all related objects

var (
	configuredBaseURL     string
	trustForwardedHeaders bool
	configMu              sync.RWMutex
)

func SetBaseURL(url string) {
	configMu.Lock()
	defer configMu.Unlock()
	configuredBaseURL = url
}

func SetTrustForwardedHeaders(trust bool) {
	configMu.Lock()
	defer configMu.Unlock()
	trustForwardedHeaders = trust
}

func writeJSON(c *gin.Context, status int, data any) {
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
				"date":              time.Now().Format(time.RFC3339),
				"request-id":        requestID,
				"client-request-id": clientRequestID,
			},
		},
	})
}

func getBaseURL(c *gin.Context) string {
	configMu.RLock()
	baseURL := configuredBaseURL
	trust := trustForwardedHeaders
	configMu.RUnlock()

	if baseURL != "" {
		return baseURL
	}
	host := c.Request.Host
	if trust {
		if forwarded := c.GetHeader("X-Forwarded-Host"); forwarded != "" {
			host = forwarded
		}
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if trust {
		if c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
	}
	return scheme + "://" + host
}

func applySelect(itemMap map[string]any, selects []string, extraPreserve ...map[string]bool) map[string]any {
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
	result := make(map[string]any, 0)
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
	return errors.Is(err, store.ErrUnableToParseFilter) ||
		errors.Is(err, store.ErrCannotCompareValues) ||
		errors.Is(err, store.ErrOperatorNotSupported) ||
		errors.Is(err, store.ErrFunctionValueMustString) ||
		errors.Is(err, store.ErrUnknownFunction) ||
		errors.Is(err, store.ErrInvalidFilterNode)
}

// buildEntityResponse creates a response map with @odata.context and merges in
// all fields from the entity by marshaling it to JSON and back.
func buildEntityResponse(odataCtx string, entity any) (map[string]any, error) {
	response := map[string]any{
		"@odata.context": odataCtx,
	}
	entityJSON, err := json.Marshal(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entity: %w", err)
	}
	var entityMap map[string]any
	if err := json.Unmarshal(entityJSON, &entityMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entity: %w", err)
	}
	for k, v := range entityMap {
		response[k] = v
	}
	return response, nil
}

// buildEntityResponseWithType creates a response map with @odata.context and
// @odata.type, then merges in all fields from the entity.
func buildEntityResponseWithType(odataCtx string, odataType string, entity any) (map[string]any, error) {
	response := map[string]any{
		"@odata.context": odataCtx,
		"@odata.type":    odataType,
	}
	entityJSON, err := json.Marshal(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entity: %w", err)
	}
	var entityMap map[string]any
	if err := json.Unmarshal(entityJSON, &entityMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entity: %w", err)
	}
	for k, v := range entityMap {
		response[k] = v
	}
	return response, nil
}

// applyNestedSelectToDirectoryObjects applies nested $select from an ExpandOption
// to a slice of DirectoryObject. Returns the raw slice if no select fields are specified,
// or a slice of filtered maps if select fields are present.
func applyNestedSelectToDirectoryObjects(objects []model.DirectoryObject, selectFields []string) any {
	if len(selectFields) == 0 {
		return objects
	}
	maps := make([]map[string]any, 0, len(objects))
	for _, obj := range objects {
		objJSON, err := json.Marshal(obj)
		if err != nil {
			maps = append(maps, map[string]any{})
			continue
		}
		var m map[string]any
		if err := json.Unmarshal(objJSON, &m); err != nil {
			maps = append(maps, map[string]any{})
			continue
		}
		m = applySelect(m, selectFields)
		maps = append(maps, m)
	}
	return maps
}

// applyNestedSelectToUser applies nested $select from an ExpandOption
// to a User object. Serializes the user to a map, applies select filtering,
// and ensures @odata.type is set to #microsoft.graph.user.
func applyNestedSelectToUser(user *model.User, selectFields []string) map[string]any {
	userJSON, _ := json.Marshal(user)
	var m map[string]any
	if err := json.Unmarshal(userJSON, &m); err != nil {
		return map[string]any{"@odata.type": "#microsoft.graph.user"}
	}
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
