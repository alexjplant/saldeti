package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/saldeti/saldeti/internal/model"
	"github.com/saldeti/saldeti/internal/store"
)

// listServicePrincipalsHandler handles GET /v1.0/servicePrincipals
func listServicePrincipalsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse OData query parameters
		opts := parseListOptions(c.Request.URL.Query())

		// Validate $top parameter
		if topStr := c.Request.URL.Query().Get("$top"); topStr != "" {
			if top, err := strconv.Atoi(topStr); err == nil && top > 0 {
				if top > 999 {
					writeError(c, http.StatusBadRequest, "Request_BadRequest",
						fmt.Sprintf("$top value %d exceeds maximum of 999.", top))
					return
				}
			}
		}

		// Call store to list service principals
		sps, totalCount, err := st.ListServicePrincipals(c.Request.Context(), opts)
		if err != nil {
			// Check if error is a filter parsing error
			errStr := err.Error()
			if strings.Contains(errStr, "unable to parse filter expression") ||
				strings.Contains(errStr, "cannot compare values") ||
				strings.Contains(errStr, "operator not supported") ||
				strings.Contains(errStr, "function value must be string") ||
				strings.Contains(errStr, "unknown function") ||
				strings.Contains(errStr, "invalid filter node") {
				writeError(c, http.StatusBadRequest, "InvalidRequest", errStr)
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list service principals")
			}
			return
		}

		// Ensure service principals is not nil
		if sps == nil {
			sps = []model.ServicePrincipal{}
		}

		// Apply $select if specified
		var responseValue interface{} = sps
		if len(opts.Select) > 0 {
			filteredItems := make([]map[string]interface{}, 0, len(sps))
			for i := range sps {
				itemJSON, err := json.Marshal(sps[i])
				if err != nil {
					writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
					return
				}
				var itemMap map[string]interface{}
				if err := json.Unmarshal(itemJSON, &itemMap); err != nil {
					writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
					return
				}
				filteredItems = append(filteredItems, applySelect(itemMap, opts.Select))
			}
			responseValue = filteredItems
		}

		// Build response
		response := model.ListResponse{
			Context: "https://graph.microsoft.com/v1.0/$metadata#servicePrincipals",
			Value:   responseValue,
		}

		// Add count if requested
		if opts.Count {
			response.Count = &totalCount
		}

		// Add nextLink if there are more results
		if opts.Top > 0 && len(sps) == opts.Top && opts.Skip+opts.Top < totalCount {
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

// getServicePrincipalHandler handles GET /v1.0/servicePrincipals/{id}
func getServicePrincipalHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Service Principal ID is required")
			return
		}

		sp, err := st.GetServicePrincipal(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, store.ErrServicePrincipalNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Service principal not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to get service principal")
			}
			return
		}

		// Build response with proper context
		response := map[string]interface{}{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#servicePrincipals/$entity",
		}

		// Merge service principal fields into response
		spJSON, err := json.Marshal(sp)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		var spMap map[string]interface{}
		if err := json.Unmarshal(spJSON, &spMap); err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		for k, v := range spMap {
			response[k] = v
		}

		// Apply $select if specified
		opts := parseListOptions(c.Request.URL.Query())
		if len(opts.Select) > 0 {
			response = applySelect(response, opts.Select)
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// getSPByAppIDHandler handles GET /v1.0/servicePrincipals/(appId={appId})
func getSPByAppIDHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get appId from context (set by middleware) or param
		appId := ""
		if val, exists := c.Get("appId"); exists {
			appId = val.(string)
		} else {
			appId = c.Param("appId")
		}
		if appId == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "App ID is required")
			return
		}

		sp, err := st.GetServicePrincipalByAppID(c.Request.Context(), appId)
		if err != nil {
			if errors.Is(err, store.ErrServicePrincipalNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Service principal not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to get service principal")
			}
			return
		}

		// Build response with proper context
		response := map[string]interface{}{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#servicePrincipals/$entity",
		}

		// Merge service principal fields into response
		spJSON, err := json.Marshal(sp)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		var spMap map[string]interface{}
		if err := json.Unmarshal(spJSON, &spMap); err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		for k, v := range spMap {
			response[k] = v
		}

		// Apply $select if specified
		opts := parseListOptions(c.Request.URL.Query())
		if len(opts.Select) > 0 {
			response = applySelect(response, opts.Select)
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// createServicePrincipalHandler handles POST /v1.0/servicePrincipals
func createServicePrincipalHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, read the raw request body
		var requestBody map[string]interface{}
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Failed to read request body")
			return
		}
		if err := json.Unmarshal(bodyBytes, &requestBody); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		var sp model.ServicePrincipal
		if err := json.Unmarshal(bodyBytes, &sp); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		// Validate that appId is present
		if sp.AppID == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "appId is required")
			return
		}

		createdSP, err := st.CreateServicePrincipal(c.Request.Context(), sp)
		if err != nil {
			if errors.Is(err, store.ErrApplicationNotFound) {
				writeError(c, http.StatusBadRequest, "InvalidRequest", "Application not found for the given appId")
			} else if errors.Is(err, store.ErrDuplicateSPAppID) {
				writeError(c, http.StatusConflict, "Conflict", "A service principal already exists for the given appId")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to create service principal")
			}
			return
		}

		// Build response with OData context
		response := map[string]interface{}{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#servicePrincipals/$entity",
		}

		// Merge service principal fields into response
		spJSON, err := json.Marshal(createdSP)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		var spMap map[string]interface{}
		if err := json.Unmarshal(spJSON, &spMap); err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		for k, v := range spMap {
			response[k] = v
		}

		c.Header("Location", "/v1.0/servicePrincipals/"+createdSP.ID)
		writeJSON(c, http.StatusCreated, response)
	}
}

// updateServicePrincipalHandler handles PATCH /v1.0/servicePrincipals/{id}
func updateServicePrincipalHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Service Principal ID is required")
			return
		}

		// Decode patch as map
		var patch map[string]interface{}
		if err := json.NewDecoder(c.Request.Body).Decode(&patch); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		sp, err := st.UpdateServicePrincipal(c.Request.Context(), id, patch)
		if err != nil {
			if errors.Is(err, store.ErrServicePrincipalNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Service principal not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to update service principal")
			}
			return
		}

		// Build response with proper context
		response := map[string]interface{}{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#servicePrincipals/$entity",
		}

		// Merge service principal fields into response
		spJSON, err := json.Marshal(sp)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		var spMap map[string]interface{}
		if err := json.Unmarshal(spJSON, &spMap); err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		for k, v := range spMap {
			response[k] = v
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// deleteServicePrincipalHandler handles DELETE /v1.0/servicePrincipals/{id}
func deleteServicePrincipalHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Service Principal ID is required")
			return
		}

		err := st.DeleteServicePrincipal(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, store.ErrServicePrincipalNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Service principal not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to delete service principal")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// listSPOwnersHandler handles GET /v1.0/servicePrincipals/{id}/owners
func listSPOwnersHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Service Principal ID is required")
			return
		}

		// Parse OData query parameters
		opts := parseListOptions(c.Request.URL.Query())

		owners, totalCount, err := st.ListSPOwners(c.Request.Context(), id, opts)
		if err != nil {
			if errors.Is(err, store.ErrServicePrincipalNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Service principal not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list owners")
			}
			return
		}

		// Ensure owners is not nil
		if owners == nil {
			owners = []model.DirectoryObject{}
		}

		// Build response
		response := model.ListResponse{
			Context: "https://graph.microsoft.com/v1.0/$metadata#directoryObjects",
			Value:   owners,
		}

		// Add count if requested
		if opts.Count {
			response.Count = &totalCount
		}

		// Add nextLink if there are more results
		if opts.Top > 0 && len(owners) == opts.Top && opts.Skip+opts.Top < totalCount {
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

// addSPOwnerHandler handles POST /v1.0/servicePrincipals/{id}/owners/$ref
func addSPOwnerHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Service Principal ID is required")
			return
		}

		var refBody map[string]string
		if err := json.NewDecoder(c.Request.Body).Decode(&refBody); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		odataID, ok := refBody["@odata.id"]
		if !ok {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "@odata.id is required")
			return
		}

		// Extract object ID from @odata.id URL
		parts := strings.Split(odataID, "/")
		if len(parts) == 0 {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid @odata.id format")
			return
		}
		objectID := parts[len(parts)-1]

		// Determine object type by looking it up in the store
		objectType, err := st.ResolveObjectType(c.Request.Context(), objectID)
		if err != nil {
			writeError(c, http.StatusNotFound, "ResourceNotFound", "Object not found")
			return
		}

		err = st.AddSPOwner(c.Request.Context(), id, objectID, objectType)
		if err != nil {
			if errors.Is(err, store.ErrServicePrincipalNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Service principal not found")
			} else if errors.Is(err, store.ErrAlreadySPOwner) {
				writeError(c, http.StatusBadRequest, "InvalidRequest", "Object is already an owner of the service principal")
			} else if errors.Is(err, store.ErrObjectNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Object not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to add owner")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// removeSPOwnerHandler handles DELETE /v1.0/servicePrincipals/{id}/owners/{ownerId}/$ref
func removeSPOwnerHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Service Principal ID is required")
			return
		}

		ownerID := c.Param("ownerId")
		if ownerID == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Owner ID is required")
			return
		}

		err := st.RemoveSPOwner(c.Request.Context(), id, ownerID)
		if err != nil {
			if errors.Is(err, store.ErrServicePrincipalNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Service principal not found")
			} else if errors.Is(err, store.ErrNotSPOwner) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Object is not an owner of the service principal")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to remove owner")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// listSPMemberOfHandler handles GET /v1.0/servicePrincipals/{id}/memberOf
func listSPMemberOfHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Service Principal ID is required")
			return
		}

		// Parse OData query parameters
		opts := parseListOptions(c.Request.URL.Query())

		groups, totalCount, err := st.ListSPMemberOf(c.Request.Context(), id, opts)
		if err != nil {
			if errors.Is(err, store.ErrServicePrincipalNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Service principal not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list member of")
			}
			return
		}

		// Ensure groups is not nil
		if groups == nil {
			groups = []model.DirectoryObject{}
		}

		// Build response
		response := model.ListResponse{
			Context: "https://graph.microsoft.com/v1.0/$metadata#directoryObjects",
			Value:   groups,
		}

		// Add count if requested
		if opts.Count {
			response.Count = &totalCount
		}

		// Add nextLink if there are more results
		if opts.Top > 0 && len(groups) == opts.Top && opts.Skip+opts.Top < totalCount {
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

// listSPTransitiveMemberOfHandler handles GET /v1.0/servicePrincipals/{id}/transitiveMemberOf
func listSPTransitiveMemberOfHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Service Principal ID is required")
			return
		}

		// Parse OData query parameters
		opts := parseListOptions(c.Request.URL.Query())

		groups, totalCount, err := st.ListSPTransitiveMemberOf(c.Request.Context(), id, opts)
		if err != nil {
			if errors.Is(err, store.ErrServicePrincipalNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Service principal not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list transitive member of")
			}
			return
		}

		// Ensure groups is not nil
		if groups == nil {
			groups = []model.DirectoryObject{}
		}

		// Build response
		response := model.ListResponse{
			Context: "https://graph.microsoft.com/v1.0/$metadata#directoryObjects",
			Value:   groups,
		}

		// Add count if requested
		if opts.Count {
			response.Count = &totalCount
		}

		// Add nextLink if there are more results
		if opts.Top > 0 && len(groups) == opts.Top && opts.Skip+opts.Top < totalCount {
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

// listSPAppRoleAssignmentsHandler handles GET /v1.0/servicePrincipals/{id}/appRoleAssignments
func listSPAppRoleAssignmentsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Service Principal ID is required")
			return
		}

		// Parse OData query parameters
		opts := parseListOptions(c.Request.URL.Query())

		assignments, totalCount, err := st.ListAppRoleAssignments(c.Request.Context(), id, opts)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list app role assignments")
			return
		}

		// Ensure assignments is not nil
		if assignments == nil {
			assignments = []model.AppRoleAssignment{}
		}

		// Build response
		response := model.ListResponse{
			Context: fmt.Sprintf("https://graph.microsoft.com/v1.0/$metadata#servicePrincipals('%s')/appRoleAssignments", id),
			Value:   assignments,
		}

		// Add count if requested
		if opts.Count {
			response.Count = &totalCount
		}

		// Add nextLink if there are more results
		if opts.Top > 0 && len(assignments) == opts.Top && opts.Skip+opts.Top < totalCount {
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

// createSPAppRoleAssignmentHandler handles POST /v1.0/servicePrincipals/{id}/appRoleAssignments
func createSPAppRoleAssignmentHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Service Principal ID is required")
			return
		}

		var req struct {
			PrincipalID string `json:"principalId"`
			ResourceID  string `json:"resourceId"`
			AppRoleID   string `json:"appRoleId"`
		}
		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		// Override principalId from URL — URL takes precedence per Graph API behavior
		req.PrincipalID = id

		// Create app role assignment using resourceID from body, principalID from URL
		assignment, err := st.CreateAppRoleAssignment(c.Request.Context(), req.ResourceID, req.PrincipalID, req.AppRoleID)
		if err != nil {
			if errors.Is(err, store.ErrServicePrincipalNotFound) {
				writeError(c, http.StatusBadRequest, "InvalidRequest", "Resource service principal not found")
			} else if errors.Is(err, store.ErrObjectNotFound) {
				writeError(c, http.StatusBadRequest, "InvalidRequest", "Principal not found")
			} else if errors.Is(err, store.ErrAppRoleNotFound) {
				writeError(c, http.StatusBadRequest, "InvalidRequest", "appRole not found on resource service principal")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to create app role assignment")
			}
			return
		}

		// Build response with OData context
		assignmentJSON, err := json.Marshal(assignment)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		var assignmentMap map[string]interface{}
		if err := json.Unmarshal(assignmentJSON, &assignmentMap); err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		assignmentMap["@odata.context"] = "https://graph.microsoft.com/v1.0/$metadata#microsoft.graph.appRoleAssignment"

		writeJSON(c, http.StatusCreated, assignmentMap)
	}
}

// deleteSPAppRoleAssignmentHandler handles DELETE /v1.0/servicePrincipals/{id}/appRoleAssignments/{assignmentId}
func deleteSPAppRoleAssignmentHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		assignmentId := c.Param("assignmentId")
		if assignmentId == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Assignment ID is required")
			return
		}

		err := st.DeleteAppRoleAssignment(c.Request.Context(), assignmentId)
		if err != nil {
			if errors.Is(err, store.ErrAssignmentNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "App role assignment not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to delete app role assignment")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// listSPAppRoleAssignedToHandler handles GET /v1.0/servicePrincipals/{id}/appRoleAssignedTo
func listSPAppRoleAssignedToHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Service Principal ID is required")
			return
		}

		// Parse OData query parameters
		opts := parseListOptions(c.Request.URL.Query())

		assignments, totalCount, err := st.ListAppRoleAssignedTo(c.Request.Context(), id, opts)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list app role assigned to")
			return
		}

		// Ensure assignments is not nil
		if assignments == nil {
			assignments = []model.AppRoleAssignment{}
		}

		// Build response
		response := model.ListResponse{
			Context: fmt.Sprintf("https://graph.microsoft.com/v1.0/$metadata#servicePrincipals('%s')/appRoleAssignedTo", id),
			Value:   assignments,
		}

		// Add count if requested
		if opts.Count {
			response.Count = &totalCount
		}

		// Add nextLink if there are more results
		if opts.Top > 0 && len(assignments) == opts.Top && opts.Skip+opts.Top < totalCount {
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

// createSPAppRoleAssignedToHandler handles POST /v1.0/servicePrincipals/{id}/appRoleAssignedTo
func createSPAppRoleAssignedToHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Service Principal ID is required")
			return
		}

		var req struct {
			PrincipalID string `json:"principalId"`
			ResourceID  string `json:"resourceId"`
			AppRoleID   string `json:"appRoleId"`
		}
		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		// Override resourceId from URL
		req.ResourceID = id

		// Create app role assignment
		assignment, err := st.CreateAppRoleAssignment(c.Request.Context(), req.ResourceID, req.PrincipalID, req.AppRoleID)
		if err != nil {
			if errors.Is(err, store.ErrServicePrincipalNotFound) {
				writeError(c, http.StatusBadRequest, "InvalidRequest", "Resource service principal not found")
			} else if errors.Is(err, store.ErrObjectNotFound) {
				writeError(c, http.StatusBadRequest, "InvalidRequest", "Principal not found")
			} else if errors.Is(err, store.ErrAppRoleNotFound) {
				writeError(c, http.StatusBadRequest, "InvalidRequest", "appRole not found on resource service principal")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to create app role assignment")
			}
			return
		}

		// Build response with OData context
		assignmentJSON, err := json.Marshal(assignment)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		var assignmentMap map[string]interface{}
		if err := json.Unmarshal(assignmentJSON, &assignmentMap); err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		assignmentMap["@odata.context"] = "https://graph.microsoft.com/v1.0/$metadata#microsoft.graph.appRoleAssignment"

		writeJSON(c, http.StatusCreated, assignmentMap)
	}
}

// deleteSPAppRoleAssignedToHandler handles DELETE /v1.0/servicePrincipals/{id}/appRoleAssignedTo/{assignmentId}
func deleteSPAppRoleAssignedToHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		assignmentId := c.Param("assignmentId")
		if assignmentId == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Assignment ID is required")
			return
		}

		err := st.DeleteAppRoleAssignment(c.Request.Context(), assignmentId)
		if err != nil {
			if errors.Is(err, store.ErrAssignmentNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "App role assignment not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to delete app role assignment")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// listSPOAuth2GrantsHandler handles GET /v1.0/servicePrincipals/{id}/oauth2PermissionGrants
func listSPOAuth2GrantsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Service Principal ID is required")
			return
		}

		// First, get the SP to retrieve its appId
		sp, err := st.GetServicePrincipal(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, store.ErrServicePrincipalNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Service principal not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to get service principal")
			}
			return
		}

		// Parse OData query parameters
		opts := parseListOptions(c.Request.URL.Query())

		// Fetch ALL grants without pagination so we can filter first, then paginate
		grants, _, err := st.ListOAuth2PermissionGrants(c.Request.Context(), model.ListOptions{})
		if err != nil {
			writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list oauth2 permission grants")
			return
		}

		// Filter client-side to only include grants where clientID matches SP's object ID
		filteredGrants := make([]model.OAuth2PermissionGrant, 0)
		for _, grant := range grants {
			if grant.ClientID == sp.ID {
				filteredGrants = append(filteredGrants, grant)
			}
		}

		// Apply pagination manually
		filteredCount := len(filteredGrants)
		if opts.Skip > 0 {
			if opts.Skip >= filteredCount {
				filteredGrants = []model.OAuth2PermissionGrant{}
			} else {
				filteredGrants = filteredGrants[opts.Skip:]
			}
		}
		if opts.Top > 0 && opts.Top < len(filteredGrants) {
			filteredGrants = filteredGrants[:opts.Top]
		}

		// Ensure filteredGrants is not nil
		if filteredGrants == nil {
			filteredGrants = []model.OAuth2PermissionGrant{}
		}

		// Build response
		response := model.ListResponse{
			Context: fmt.Sprintf("https://graph.microsoft.com/v1.0/$metadata#servicePrincipals('%s')/oauth2PermissionGrants", id),
			Value:   filteredGrants,
		}

		// Add count if requested
		if opts.Count {
			count := filteredCount
			response.Count = &count
		}

		// Add nextLink if there are more results
		if opts.Top > 0 && len(filteredGrants) == opts.Top && opts.Skip+opts.Top < filteredCount {
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

// spAddPasswordHandler handles POST /v1.0/servicePrincipals/{id}/addPassword
func spAddPasswordHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Service Principal ID is required")
			return
		}

		var req model.PasswordCredential
		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		// Capture request fields for use inside the callback
		displayName := req.DisplayName
		var providedEndDateTime *time.Time
		if req.EndDateTime != nil {
			providedEndDateTime = req.EndDateTime
		}

		var resultCred model.PasswordCredential
		err := st.UpdateSPCredentials(c.Request.Context(), id, func(sp *model.ServicePrincipal) error {
			keyID := uuid.New().String()
			secretText := strings.ReplaceAll(uuid.New().String()+uuid.New().String(), "-", "")[:32]
			now := time.Now()
			startDateTime := now

			var endDateTime time.Time
			if providedEndDateTime != nil {
				endDateTime = *providedEndDateTime
			} else {
				endDateTime = now.Add(2 * 365 * 24 * time.Hour)
			}

			hint := secretText
			if len(hint) > 3 {
				hint = hint[:3]
			}
			hint += "***"

			cred := model.PasswordCredential{
				DisplayName:   displayName,
				KeyID:         keyID,
				SecretText:    secretText,
				StartDateTime: &startDateTime,
				EndDateTime:   &endDateTime,
				Hint:          hint,
			}

			sp.PasswordCredentials = append(sp.PasswordCredentials, cred)
			resultCred = cred
			return nil
		})
		if err != nil {
			if errors.Is(err, store.ErrServicePrincipalNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Service principal not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to add password")
			}
			return
		}

		// Build response with OData context
		credJSON, err := json.Marshal(resultCred)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		var credMap map[string]interface{}
		if err := json.Unmarshal(credJSON, &credMap); err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		credMap["@odata.context"] = "https://graph.microsoft.com/v1.0/$metadata#microsoft.graph.passwordCredential"

		writeJSON(c, http.StatusOK, credMap)
	}
}

// spRemovePasswordHandler handles POST /v1.0/servicePrincipals/{id}/removePassword
func spRemovePasswordHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Service Principal ID is required")
			return
		}

		var req struct {
			KeyID string `json:"keyId"`
		}
		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		if req.KeyID == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "keyId is required")
			return
		}

		// Update SP credentials using the callback method
		err := st.UpdateSPCredentials(c.Request.Context(), id, func(sp *model.ServicePrincipal) error {
			found := false
			updatedCreds := make([]model.PasswordCredential, 0, len(sp.PasswordCredentials))
			for _, pc := range sp.PasswordCredentials {
				if pc.KeyID == req.KeyID {
					found = true
					continue
				}
				updatedCreds = append(updatedCreds, pc)
			}
			if !found {
				return store.ErrCredentialNotFound
			}
			sp.PasswordCredentials = updatedCreds
			return nil
		})
		if err != nil {
			if errors.Is(err, store.ErrServicePrincipalNotFound) || errors.Is(err, store.ErrCredentialNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Credential not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to remove password")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// spAddKeyHandler handles POST /v1.0/servicePrincipals/{id}/addKey
func spAddKeyHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Service Principal ID is required")
			return
		}

		var cred model.KeyCredential
		if err := json.NewDecoder(c.Request.Body).Decode(&cred); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		var resultCred model.KeyCredential
		err := st.UpdateSPCredentials(c.Request.Context(), id, func(sp *model.ServicePrincipal) error {
			if cred.KeyID == "" {
				cred.KeyID = uuid.New().String()
			}
			if cred.StartDateTime == nil {
				now := time.Now()
				cred.StartDateTime = &now
			}
			sp.KeyCredentials = append(sp.KeyCredentials, cred)
			resultCred = cred
			return nil
		})
		if err != nil {
			if errors.Is(err, store.ErrServicePrincipalNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Service principal not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to add key")
			}
			return
		}

		// Build response with OData context
		credJSON, err := json.Marshal(resultCred)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		var credMap map[string]interface{}
		if err := json.Unmarshal(credJSON, &credMap); err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		credMap["@odata.context"] = "https://graph.microsoft.com/v1.0/$metadata#microsoft.graph.keyCredential"

		writeJSON(c, http.StatusOK, credMap)
	}
}

// spRemoveKeyHandler handles POST /v1.0/servicePrincipals/{id}/removeKey
func spRemoveKeyHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Service Principal ID is required")
			return
		}

		var req struct {
			KeyID string `json:"keyId"`
		}
		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		if req.KeyID == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "keyId is required")
			return
		}

		// Update SP credentials using the callback method
		err := st.UpdateSPCredentials(c.Request.Context(), id, func(sp *model.ServicePrincipal) error {
			found := false
			updatedCreds := make([]model.KeyCredential, 0, len(sp.KeyCredentials))
			for _, kc := range sp.KeyCredentials {
				if kc.KeyID == req.KeyID {
					found = true
					continue
				}
				updatedCreds = append(updatedCreds, kc)
			}
			if !found {
				return store.ErrCredentialNotFound
			}
			sp.KeyCredentials = updatedCreds
			return nil
		})
		if err != nil {
			if errors.Is(err, store.ErrServicePrincipalNotFound) || errors.Is(err, store.ErrCredentialNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Credential not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to remove key")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// listEmptyPoliciesHandler returns a handler that returns empty policy lists
func listEmptyPoliciesHandler(st store.Store, collectionName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Service Principal ID is required")
			return
		}

		// Verify SP exists
		_, err := st.GetServicePrincipal(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, store.ErrServicePrincipalNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Service principal not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to get service principal")
			}
			return
		}

		writeJSON(c, http.StatusOK, gin.H{
			"@odata.context": fmt.Sprintf("https://graph.microsoft.com/v1.0/$metadata#servicePrincipals('%s')/%s", id, collectionName),
			"value":          []interface{}{},
		})
	}
}
