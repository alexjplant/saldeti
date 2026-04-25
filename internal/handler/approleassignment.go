package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/model"
	"github.com/saldeti/saldeti/internal/store"
)

// listUserAppRoleAssignmentsHandler handles GET /v1.0/users/{id}/appRoleAssignments
func listUserAppRoleAssignmentsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		if userID == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "User ID is required")
			return
		}

		// Parse OData options
		opts := parseListOptions(c.Request.URL.Query())

		// Call st.ListAppRoleAssignments(ctx, userID, opts)
		assignments, totalCount, err := st.ListAppRoleAssignments(c.Request.Context(), userID, opts)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list app role assignments")
			return
		}

		// Ensure assignments not nil
		if assignments == nil {
			assignments = []model.AppRoleAssignment{}
		}

		// Build model.ListResponse with Context
		response := model.ListResponse{
			Context: fmt.Sprintf("https://graph.microsoft.com/v1.0/$metadata#users('%s')/appRoleAssignments", userID),
			Value:   assignments,
		}

		// Add count if opts.Count
		if opts.Count {
			response.Count = &totalCount
		}

		// Add nextLink if pagination applies (same pattern as listSPAppRoleAssignmentsHandler)
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

// createUserAppRoleAssignmentHandler handles POST /v1.0/users/{id}/appRoleAssignments
func createUserAppRoleAssignmentHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		if userID == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "User ID is required")
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

		// Call st.CreateAppRoleAssignment(ctx, req.ResourceID, userID, req.AppRoleID)
		assignment, err := st.CreateAppRoleAssignment(c.Request.Context(), req.ResourceID, userID, req.AppRoleID)
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

		// Build response: marshal assignment to JSON, unmarshal to map, add @odata.context
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

// deleteUserAppRoleAssignmentHandler handles DELETE /v1.0/users/{id}/appRoleAssignments/{assignmentId}
func deleteUserAppRoleAssignmentHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		assignmentId := c.Param("assignmentId")
		if assignmentId == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Assignment ID is required")
			return
		}

		// Call st.DeleteAppRoleAssignment(ctx, assignmentId)
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

// listGroupAppRoleAssignmentsHandler handles GET /v1.0/groups/{id}/appRoleAssignments
func listGroupAppRoleAssignmentsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupID := c.Param("id")
		if groupID == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Group ID is required")
			return
		}

		// Parse OData options
		opts := parseListOptions(c.Request.URL.Query())

		// Call st.ListAppRoleAssignments(ctx, groupID, opts)
		assignments, totalCount, err := st.ListAppRoleAssignments(c.Request.Context(), groupID, opts)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list app role assignments")
			return
		}

		// Ensure assignments not nil
		if assignments == nil {
			assignments = []model.AppRoleAssignment{}
		}

		// Build model.ListResponse with Context
		response := model.ListResponse{
			Context: fmt.Sprintf("https://graph.microsoft.com/v1.0/$metadata#groups('%s')/appRoleAssignments", groupID),
			Value:   assignments,
		}

		// Add count if opts.Count
		if opts.Count {
			response.Count = &totalCount
		}

		// Add nextLink if pagination applies
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

// createGroupAppRoleAssignmentHandler handles POST /v1.0/groups/{id}/appRoleAssignments
func createGroupAppRoleAssignmentHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupID := c.Param("id")
		if groupID == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Group ID is required")
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

		// Use groupID as principalID in CreateAppRoleAssignment call
		assignment, err := st.CreateAppRoleAssignment(c.Request.Context(), req.ResourceID, groupID, req.AppRoleID)
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

		// Build response
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

// deleteGroupAppRoleAssignmentHandler handles DELETE /v1.0/groups/{id}/appRoleAssignments/{assignmentId}
func deleteGroupAppRoleAssignmentHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		assignmentId := c.Param("assignmentId")
		if assignmentId == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Assignment ID is required")
			return
		}

		// Call st.DeleteAppRoleAssignment(ctx, assignmentId) - identical to deleteUserAppRoleAssignmentHandler
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
