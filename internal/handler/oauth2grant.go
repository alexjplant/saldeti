package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/saldeti/saldeti/internal/model"
	"github.com/saldeti/saldeti/internal/store"
)

// listGrantsHandler handles GET /v1.0/oauth2PermissionGrants
func listGrantsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse OData options
		opts := parseListOptions(c.Request.URL.Query())

		// Validate $top <= 999
		if topStr := c.Request.URL.Query().Get("$top"); topStr != "" {
			if top, err := strconv.Atoi(topStr); err == nil && top > 0 {
				if top > 999 {
					writeError(c, http.StatusBadRequest, "Request_BadRequest",
						fmt.Sprintf("$top value %d exceeds maximum of 999.", top))
					return
				}
			}
		}

		// Call store to list grants
		grants, totalCount, err := st.ListOAuth2PermissionGrants(c.Request.Context(), opts)
		if err != nil {
			// Handle filter parse errors (same pattern as application handler)
			errStr := err.Error()
			if strings.Contains(errStr, "unable to parse filter expression") ||
			   strings.Contains(errStr, "cannot compare values") ||
			   strings.Contains(errStr, "operator not supported") ||
			   strings.Contains(errStr, "function value must be string") ||
			   strings.Contains(errStr, "unknown function") ||
			   strings.Contains(errStr, "invalid filter node") {
				writeError(c, http.StatusBadRequest, "InvalidRequest", errStr)
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list oauth2 permission grants")
			}
			return
		}

		// Ensure grants is not nil
		if grants == nil {
			grants = []model.OAuth2PermissionGrant{}
		}

		// Build model.ListResponse
		response := model.ListResponse{
			Context: "https://graph.microsoft.com/v1.0/$metadata#oauth2PermissionGrants",
			Value:   grants,
		}

		// Add count if opts.Count
		if opts.Count {
			response.Count = &totalCount
		}

		// Add nextLink if pagination applies
		if opts.Top > 0 && len(grants) == opts.Top && opts.Skip+opts.Top < totalCount {
			nextSkip := opts.Skip + opts.Top
			nextURL := url.URL{
				Path:     c.Request.URL.Path,
				RawQuery: buildNextLinkQuery(c.Request.URL.Query(), nextSkip),
			}

			response.NextLink = getBaseURL(c) + nextURL.String()
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// getGrantHandler handles GET /v1.0/oauth2PermissionGrants/{id}
func getGrantHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "OAuth2 permission grant ID is required")
			return
		}

		grant, err := st.GetOAuth2PermissionGrant(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, store.ErrGrantNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "OAuth2 permission grant not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to get oauth2 permission grant")
			}
			return
		}

		// Build response map with @odata.context
		response := map[string]interface{}{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#oauth2PermissionGrants/$entity",
		}

		// Marshal/unmarshal grant to map, merge into response
		grantJSON, err := json.Marshal(grant)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		var grantMap map[string]interface{}
		if err := json.Unmarshal(grantJSON, &grantMap); err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		for k, v := range grantMap {
			response[k] = v
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// createGrantHandler handles POST /v1.0/oauth2PermissionGrants
func createGrantHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var grant model.OAuth2PermissionGrant
		if err := json.NewDecoder(c.Request.Body).Decode(&grant); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		// Validate required fields
		if grant.ClientID == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "clientId is required")
			return
		}
		if grant.ConsentType == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "consentType is required")
			return
		}
		if grant.ConsentType != "AllPrincipals" && grant.ConsentType != "Principal" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "consentType must be 'AllPrincipals' or 'Principal'")
			return
		}
		if grant.ResourceID == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "resourceId is required")
			return
		}
		if grant.Scope == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "scope is required")
			return
		}
		if grant.ConsentType == "Principal" && grant.PrincipalID == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "principalId is required when consentType is 'Principal'")
			return
		}

		// If grant.ID == "" -> grant.ID = uuid.New().String()
		if grant.ID == "" {
			grant.ID = uuid.New().String()
		}

		// If grant.ODataType == "" -> grant.ODataType = "#microsoft.graph.oAuth2PermissionGrant"
		if grant.ODataType == "" {
			grant.ODataType = "#microsoft.graph.oAuth2PermissionGrant"
		}

		// Call st.CreateOAuth2PermissionGrant(ctx, grant)
		createdGrant, err := st.CreateOAuth2PermissionGrant(c.Request.Context(), grant)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "InternalError", "Failed to create oauth2 permission grant")
			return
		}

		// Build response with @odata.context
		response := map[string]interface{}{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#oauth2PermissionGrants/$entity",
		}

		// Marshal/unmarshal, merge
		grantJSON, err := json.Marshal(createdGrant)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		var grantMap map[string]interface{}
		if err := json.Unmarshal(grantJSON, &grantMap); err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		for k, v := range grantMap {
			response[k] = v
		}

		// Set Location header
		c.Header("Location", "/v1.0/oauth2PermissionGrants/"+createdGrant.ID)

		writeJSON(c, http.StatusCreated, response)
	}
}

// updateGrantHandler handles PATCH /v1.0/oauth2PermissionGrants/{id}
func updateGrantHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "OAuth2 permission grant ID is required")
			return
		}

		// Decode body as map[string]interface{}
		var patch map[string]interface{}
		if err := json.NewDecoder(c.Request.Body).Decode(&patch); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		// Call st.UpdateOAuth2PermissionGrant(ctx, id, patch)
		grant, err := st.UpdateOAuth2PermissionGrant(c.Request.Context(), id, patch)
		if err != nil {
			if errors.Is(err, store.ErrGrantNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "OAuth2 permission grant not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to update oauth2 permission grant")
			}
			return
		}

		// Build response with @odata.context
		response := map[string]interface{}{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#oauth2PermissionGrants/$entity",
		}

		// Marshal/unmarshal, merge
		grantJSON, err := json.Marshal(grant)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		var grantMap map[string]interface{}
		if err := json.Unmarshal(grantJSON, &grantMap); err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		for k, v := range grantMap {
			response[k] = v
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// deleteGrantHandler handles DELETE /v1.0/oauth2PermissionGrants/{id}
func deleteGrantHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "OAuth2 permission grant ID is required")
			return
		}

		// Call st.DeleteOAuth2PermissionGrant(ctx, id)
		err := st.DeleteOAuth2PermissionGrant(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, store.ErrGrantNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "OAuth2 permission grant not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to delete oauth2 permission grant")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}
