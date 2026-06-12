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

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/entra/model"
	"github.com/saldeti/saldeti/internal/entra/store"
)

// listApplicationsHandler handles GET /v1.0/applications
func listApplicationsHandler(st store.Store) gin.HandlerFunc {
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

		// Call store to list applications
		applications, totalCount, err := st.ListApplications(c.Request.Context(), opts)
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
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list applications")
			}
			return
		}

		// Ensure applications is not nil
		if applications == nil {
			applications = []model.Application{}
		}

		var responseValue interface{} = applications
		if len(opts.ExpandOptions) > 0 {
			expandedApps := make([]map[string]interface{}, 0, len(applications))
			for _, app := range applications {
				appMap := make(map[string]interface{})
				appJSON, err := json.Marshal(app)
				if err != nil {
					writeError(c, http.StatusInternalServerError, "InternalError", "Failed to process application expansion")
					return
				}
				if err := json.Unmarshal(appJSON, &appMap); err != nil {
					writeError(c, http.StatusInternalServerError, "InternalError", "Failed to process application expansion")
					return
				}

				for _, expand := range opts.ExpandOptions {
					switch expand.Property {
			case "owners":
					owners, _, err := st.ListApplicationOwners(c.Request.Context(), app.ID, model.ListOptions{Top: 999})
					if err == nil {
						appMap["owners"] = applyNestedSelectToDirectoryObjects(nilToEmptyDirectoryObjects(owners), expand.Select)
					}
					}
				}
				expandedApps = append(expandedApps, appMap)
			}
			responseValue = expandedApps
		}

		// Apply $select if specified
		if len(opts.Select) > 0 {
			if len(opts.ExpandOptions) > 0 {
				expandedProps := computeExpandedPropertyNames(opts.ExpandOptions)
				maps := responseValue.([]map[string]interface{})
				for i, m := range maps {
					maps[i] = applySelect(m, opts.Select, expandedProps)
				}
			} else {
				filteredItems := make([]map[string]interface{}, 0, len(applications))
				for i := range applications {
					itemJSON, err := json.Marshal(applications[i])
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
		}

		// Build response
		response := model.ListResponse{
			Context: "https://graph.microsoft.com/v1.0/$metadata#applications",
			Value:   responseValue,
		}

		// Add count if requested
		if opts.Count {
			response.Count = &totalCount
		}

		// Add nextLink if there are more results
		if opts.Top > 0 && len(applications) == opts.Top && opts.Skip+opts.Top < totalCount {
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

// getApplicationHandler handles GET /v1.0/applications/{id}
func getApplicationHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Application ID is required")
			return
		}

		application, err := st.GetApplication(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, store.ErrApplicationNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Application not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to get application")
			}
			return
		}

		response, err := buildEntityResponse("https://graph.microsoft.com/v1.0/$metadata#applications/$entity", application)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		opts := parseListOptions(c.Request.URL.Query())

		// Handle $expand
		for _, expand := range opts.ExpandOptions {
			switch expand.Property {
			case "owners":
				owners, _, err := st.ListApplicationOwners(c.Request.Context(), id, model.ListOptions{Top: 999})
				if err == nil {
					response["owners"] = applyNestedSelectToDirectoryObjects(nilToEmptyDirectoryObjects(owners), expand.Select)
				}
			}
		}

		// Apply $select if specified
		if len(opts.Select) > 0 {
			// Preserve expanded properties so $select doesn't strip them
			expandedProps := computeExpandedPropertyNames(opts.ExpandOptions)
			response = applySelect(response, opts.Select, expandedProps)
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// getApplicationByAppIDHandler handles GET /v1.0/applications/(appId={appId})
func getApplicationByAppIDHandler(st store.Store) gin.HandlerFunc {
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

		application, err := st.GetApplicationByAppID(c.Request.Context(), appId)
		if err != nil {
			if errors.Is(err, store.ErrApplicationNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Application not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to get application")
			}
			return
		}

		response, err := buildEntityResponse("https://graph.microsoft.com/v1.0/$metadata#applications/$entity", application)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		opts := parseListOptions(c.Request.URL.Query())

		// Handle $expand
		for _, expand := range opts.ExpandOptions {
			switch expand.Property {
			case "owners":
				owners, _, err := st.ListApplicationOwners(c.Request.Context(), application.ID, model.ListOptions{Top: 999})
				if err == nil {
					response["owners"] = applyNestedSelectToDirectoryObjects(nilToEmptyDirectoryObjects(owners), expand.Select)
				}
			}
		}

		// Apply $select if specified
		if len(opts.Select) > 0 {
			expandedProps := computeExpandedPropertyNames(opts.ExpandOptions)
			response = applySelect(response, opts.Select, expandedProps)
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// createApplicationHandler handles POST /v1.0/applications
func createApplicationHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, read the raw request body to handle owners@odata.bind
		var requestBody map[string]interface{}
		bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBodyBytes))
		if err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Failed to read request body")
			return
		}
		if err := json.Unmarshal(bodyBytes, &requestBody); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		var app model.Application
		if err := json.Unmarshal(bodyBytes, &app); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		// Handle owners@odata.bind
		if ownersBind, ok := requestBody["owners@odata.bind"]; ok {
			if ownersArray, ok := ownersBind.([]interface{}); ok {
				for _, ownerRef := range ownersArray {
					if ownerStr, ok := ownerRef.(string); ok {
						app.Owners = append(app.Owners, model.DirectoryObjectRef{
							ODataID: ownerStr,
						})
					}
				}
			} else if ownerStr, ok := ownersBind.(string); ok {
				app.Owners = append(app.Owners, model.DirectoryObjectRef{
					ODataID: ownerStr,
				})
			}
		}

		// Validate required fields
		if app.DisplayName == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "displayName is required")
			return
		}

		// Set default OData type if not provided
		if app.ODataType == "" {
			app.ODataType = "#microsoft.graph.application"
		}

		createdApp, err := st.CreateApplication(c.Request.Context(), app)
		if err != nil {
			if errors.Is(err, store.ErrDuplicateAppID) {
				writeError(c, http.StatusConflict, "Conflict", "Application already exists")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to create application")
			}
			return
		}

		response, err := buildEntityResponse("https://graph.microsoft.com/v1.0/$metadata#applications/$entity", createdApp)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		c.Header("Location", "/v1.0/applications/"+createdApp.ID)
		writeJSON(c, http.StatusCreated, response)
	}
}

// updateApplicationHandler handles PATCH /v1.0/applications/{id}
func updateApplicationHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Application ID is required")
			return
		}

		// Decode patch as map
		var patch map[string]interface{}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&patch); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		application, err := st.UpdateApplication(c.Request.Context(), id, patch)
		if err != nil {
			if errors.Is(err, store.ErrApplicationNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Application not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to update application")
			}
			return
		}

		response, err := buildEntityResponse("https://graph.microsoft.com/v1.0/$metadata#applications/$entity", application)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// deleteApplicationHandler handles DELETE /v1.0/applications/{id}
func deleteApplicationHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Application ID is required")
			return
		}

		err := st.DeleteApplication(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, store.ErrApplicationNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Application not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to delete application")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// addPasswordHandler handles POST /v1.0/applications/{id}/addPassword
func addPasswordHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Application ID is required")
			return
		}

		var req struct {
			PasswordCredential model.PasswordCredential `json:"passwordCredential"`
		}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&req); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		addedCred, err := st.AddApplicationPassword(c.Request.Context(), id, req.PasswordCredential)
		if err != nil {
			if errors.Is(err, store.ErrApplicationNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Application not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to add password")
			}
			return
		}

		credMap, err := buildEntityResponse("https://graph.microsoft.com/v1.0/$metadata#microsoft.graph.passwordCredential", addedCred)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		// Explicitly include secretText in addPassword response since the json tag is "-"
		credMap["secretText"] = addedCred.SecretText

		writeJSON(c, http.StatusOK, credMap)
	}
}

// removePasswordHandler handles POST /v1.0/applications/{id}/removePassword
func removePasswordHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Application ID is required")
			return
		}

		var req struct {
			KeyID string `json:"keyId"`
		}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&req); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		if req.KeyID == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "keyId is required")
			return
		}

		err := st.RemoveApplicationPassword(c.Request.Context(), id, req.KeyID)
		if err != nil {
			if errors.Is(err, store.ErrApplicationNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Application not found")
			} else if errors.Is(err, store.ErrCredentialNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Credential not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to remove password")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// addKeyHandler handles POST /v1.0/applications/{id}/addKey
func addKeyHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Application ID is required")
			return
		}

		var req struct {
			KeyCredential model.KeyCredential `json:"keyCredential"`
		}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&req); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		addedCred, err := st.AddApplicationKey(c.Request.Context(), id, req.KeyCredential)
		if err != nil {
			if errors.Is(err, store.ErrApplicationNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Application not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to add key")
			}
			return
		}

		credMap, err := buildEntityResponse("https://graph.microsoft.com/v1.0/$metadata#microsoft.graph.keyCredential", addedCred)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		writeJSON(c, http.StatusOK, credMap)
	}
}

// removeKeyHandler handles POST /v1.0/applications/{id}/removeKey
func removeKeyHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Application ID is required")
			return
		}

		var req struct {
			KeyID string `json:"keyId"`
		}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&req); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		if req.KeyID == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "keyId is required")
			return
		}

		err := st.RemoveApplicationKey(c.Request.Context(), id, req.KeyID)
		if err != nil {
			if errors.Is(err, store.ErrApplicationNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Application not found")
			} else if errors.Is(err, store.ErrCredentialNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Credential not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to remove key")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// listApplicationOwnersHandler handles GET /v1.0/applications/{id}/owners
func listApplicationOwnersHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Application ID is required")
			return
		}

		// Parse OData query parameters
		opts := parseListOptions(c.Request.URL.Query())

		owners, totalCount, err := st.ListApplicationOwners(c.Request.Context(), id, opts)
		if err != nil {
			if errors.Is(err, store.ErrApplicationNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Application not found")
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

// addApplicationOwnerHandler handles POST /v1.0/applications/{id}/owners/$ref
func addApplicationOwnerHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Application ID is required")
			return
		}

		var refBody map[string]string
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&refBody); err != nil {
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

		err = st.AddApplicationOwner(c.Request.Context(), id, objectID, objectType)
		if err != nil {
			if errors.Is(err, store.ErrApplicationNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Application not found")
			} else if errors.Is(err, store.ErrAlreadyAppOwner) {
				writeError(c, http.StatusBadRequest, "InvalidRequest", "Object is already an owner of the application")
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

// removeApplicationOwnerHandler handles DELETE /v1.0/applications/{id}/owners/{ownerId}/$ref
func removeApplicationOwnerHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Application ID is required")
			return
		}

		ownerID := c.Param("ownerId")
		if ownerID == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Owner ID is required")
			return
		}

		err := st.RemoveApplicationOwner(c.Request.Context(), id, ownerID)
		if err != nil {
			if errors.Is(err, store.ErrApplicationNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Application not found")
			} else if errors.Is(err, store.ErrNotAppOwner) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Object is not an owner of the application")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to remove owner")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// listExtensionPropertiesHandler handles GET /v1.0/applications/{id}/extensionProperties
func listExtensionPropertiesHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Application ID is required")
			return
		}

		extProps, err := st.ListExtensionProperties(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, store.ErrApplicationNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Application not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list extension properties")
			}
			return
		}

		// Ensure extProps is not nil
		if extProps == nil {
			extProps = []model.ExtensionProperty{}
		}

		response := map[string]interface{}{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#extensionProperties",
			"value":          extProps,
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// createExtensionPropertyHandler handles POST /v1.0/applications/{id}/extensionProperties
func createExtensionPropertyHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Application ID is required")
			return
		}

		var ep model.ExtensionProperty
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&ep); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		// Validate required fields
		if ep.Name == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "name is required")
			return
		}

		if ep.DataType == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "dataType is required")
			return
		}

		// Validate dataType is one of the allowed values
		if !model.ValidExtensionDataTypes[ep.DataType] {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "dataType must be one of: Binary, Boolean, DateTime, Integer, LargeString, String")
			return
		}

		createdEP, err := st.CreateExtensionProperty(c.Request.Context(), id, ep)
		if err != nil {
			if errors.Is(err, store.ErrApplicationNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Application not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to create extension property")
			}
			return
		}

		epMap, err := buildEntityResponse("https://graph.microsoft.com/v1.0/$metadata#extensionProperties/$entity", createdEP)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		c.Header("Location", "/v1.0/applications/"+id+"/extensionProperties/"+createdEP.ID)
		writeJSON(c, http.StatusCreated, epMap)
	}
}

// deleteExtensionPropertyHandler handles DELETE /v1.0/applications/{id}/extensionProperties/{extId}
func deleteExtensionPropertyHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Application ID is required")
			return
		}

		extID := c.Param("extId")
		if extID == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Extension property ID is required")
			return
		}

		err := st.DeleteExtensionProperty(c.Request.Context(), id, extID)
		if err != nil {
			if errors.Is(err, store.ErrApplicationNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Application not found")
			} else if errors.Is(err, store.ErrExtensionNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Extension property not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to delete extension property")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// setVerifiedPublisherHandler handles POST /v1.0/applications/{id}/setVerifiedPublisher
func setVerifiedPublisherHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Application ID is required")
			return
		}

		var body model.VerifiedPublisher
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&body); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		_, err := st.UpdateApplication(c.Request.Context(), id, map[string]interface{}{"verifiedPublisher": body})
		if err != nil {
			if errors.Is(err, store.ErrApplicationNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "Application not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to set verified publisher")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// applicationsDeltaHandler handles GET /v1.0/applications/delta
func applicationsDeltaHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		deltaToken := c.Request.URL.Query().Get("$deltatoken")

		items, newDeltaToken, _, err := st.GetApplicationsDelta(c.Request.Context(), deltaToken)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "InternalError", "Failed to get delta")
			return
		}

		response := map[string]interface{}{
			"@odata.context":    "https://graph.microsoft.com/v1.0/$metadata#applications",
			"value":             items,
			"@odata.deltaLink":  getBaseURL(c) + "/v1.0/applications/delta?$deltatoken=" + newDeltaToken,
		}

		writeJSON(c, http.StatusOK, response)
	}
}
