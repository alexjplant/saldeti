package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/model"
	"github.com/saldeti/saldeti/internal/store"
)

// listUserMemberOfHandler handles GET /v1.0/users/{id}/memberOf
func listUserMemberOfHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param( "id")
		opts := parseListOptions(c.Request.URL.Query())

		memberOf, totalCount, err := st.ListUserMemberOf(c.Request.Context(), userID, opts)
		if err != nil {
			if errors.Is(err, store.ErrUserNotFound) {
				writeError(c, http.StatusNotFound, "Request_ResourceNotFound", "User not found")
				return
			}
			writeError(c, http.StatusInternalServerError, "InternalError", err.Error())
			return
		}

		response := model.ListResponse{
			Context: "https://graph.microsoft.com/v1.0/$metadata#directoryObjects",
			Value:   memberOf,
		}
		if opts.Count {
			response.Count = &totalCount
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// listUserTransitiveMemberOfHandler handles GET /v1.0/users/{id}/transitiveMemberOf
func listUserTransitiveMemberOfHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param( "id")
		opts := parseListOptions(c.Request.URL.Query())

		memberOf, totalCount, err := st.ListUserTransitiveMemberOf(c.Request.Context(), userID, opts)
		if err != nil {
			if errors.Is(err, store.ErrUserNotFound) {
				writeError(c, http.StatusNotFound, "Request_ResourceNotFound", "User not found")
				return
			}
			writeError(c, http.StatusInternalServerError, "InternalError", err.Error())
			return
		}

		response := model.ListResponse{
			Context: "https://graph.microsoft.com/v1.0/$metadata#directoryObjects",
			Value:   memberOf,
		}
		if opts.Count {
			response.Count = &totalCount
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// getManagerHandler handles GET /v1.0/users/{id}/manager
func getManagerHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param( "id")

		manager, err := st.GetManager(c.Request.Context(), userID)
		if err != nil {
			if errors.Is(err, store.ErrUserNotFound) {
				writeError(c, http.StatusNotFound, "Request_ResourceNotFound", "User not found")
				return
			}
			if errors.Is(err, store.ErrManagerNotFound) {
				writeError(c, http.StatusNotFound, "Request_ResourceNotFound", "Manager not set for user")
				return
			}
			writeError(c, http.StatusInternalServerError, "InternalError", err.Error())
			return
		}

		writeJSON(c, http.StatusOK, manager)
	}
}

// setManagerHandler handles PUT /v1.0/users/{id}/manager/$ref
func setManagerHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param( "id")

		var requestBody struct {
			ODataID string `json:"@odata.id"`
		}
		if err := json.NewDecoder(c.Request.Body).Decode(&requestBody); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON")
			return
		}

		// Extract manager ID from @odata.id URL
		// Format: "https://graph.microsoft.com/v1.0/users/{managerId}"
		parts := strings.Split(requestBody.ODataID, "/")
		if len(parts) == 0 {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid @odata.id format")
			return
		}
		managerID := parts[len(parts)-1]

		if err := st.SetManager(c.Request.Context(), userID, managerID); err != nil {
			if errors.Is(err, store.ErrUserNotFound) {
				writeError(c, http.StatusNotFound, "Request_ResourceNotFound", "User or manager not found")
				return
			}
			writeError(c, http.StatusInternalServerError, "InternalError", err.Error())
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// removeManagerHandler handles DELETE /v1.0/users/{id}/manager/$ref
func removeManagerHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param( "id")

		if err := st.RemoveManager(c.Request.Context(), userID); err != nil {
			if errors.Is(err, store.ErrUserNotFound) {
				writeError(c, http.StatusNotFound, "Request_ResourceNotFound", "User not found")
				return
			}
			writeError(c, http.StatusInternalServerError, "InternalError", err.Error())
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// listDirectReportsHandler handles GET /v1.0/users/{id}/directReports
func listDirectReportsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param( "id")
		opts := parseListOptions(c.Request.URL.Query())

		directReports, totalCount, err := st.ListDirectReports(c.Request.Context(), userID, opts)
		if err != nil {
			if errors.Is(err, store.ErrUserNotFound) {
				writeError(c, http.StatusNotFound, "Request_ResourceNotFound", "User not found")
				return
			}
			writeError(c, http.StatusInternalServerError, "InternalError", err.Error())
			return
		}

		response := model.ListResponse{
			Context: "https://graph.microsoft.com/v1.0/$metadata#directoryObjects",
			Value:   directReports,
		}
		if opts.Count {
			response.Count = &totalCount
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// getByIdsHandler handles POST /v1.0/directoryObjects/getByIds
func getByIdsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestBody struct {
			IDs   []string `json:"ids"`
			Types []string `json:"types,omitempty"`
		}
		if err := json.NewDecoder(c.Request.Body).Decode(&requestBody); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON")
			return
		}

		if len(requestBody.IDs) == 0 {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "ids field is required")
			return
		}

		objects, err := st.GetDirectoryObjects(c.Request.Context(), requestBody.IDs, requestBody.Types)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "InternalError", err.Error())
			return
		}

		writeJSON(c, http.StatusOK, map[string]interface{}{
			"value": objects,
		})
	}
}

// checkUserMemberGroupsHandler handles POST /v1.0/users/{id}/checkMemberGroups
func checkUserMemberGroupsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param( "id")

		var requestBody struct {
			GroupIDs []string `json:"groupIds"`
		}
		if err := json.NewDecoder(c.Request.Body).Decode(&requestBody); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON")
			return
		}

		if len(requestBody.GroupIDs) == 0 {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "groupIds field is required")
			return
		}

		memberGroups, err := st.CheckMemberGroups(c.Request.Context(), userID, requestBody.GroupIDs)
		if err != nil {
			if errors.Is(err, store.ErrUserNotFound) {
				writeError(c, http.StatusNotFound, "Request_ResourceNotFound", "User not found")
				return
			}
			writeError(c, http.StatusInternalServerError, "InternalError", err.Error())
			return
		}

		writeJSON(c, http.StatusOK, map[string]interface{}{
			"value": memberGroups,
		})
	}
}

// getUserMemberGroupsHandler handles POST /v1.0/users/{id}/getMemberGroups
func getUserMemberGroupsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param( "id")

		var requestBody struct {
			SecurityEnabledOnly bool `json:"securityEnabledOnly"`
		}
		if err := json.NewDecoder(c.Request.Body).Decode(&requestBody); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON")
			return
		}

		memberGroups, err := st.GetMemberGroups(c.Request.Context(), userID, requestBody.SecurityEnabledOnly)
		if err != nil {
			if errors.Is(err, store.ErrUserNotFound) {
				writeError(c, http.StatusNotFound, "Request_ResourceNotFound", "User not found")
				return
			}
			writeError(c, http.StatusInternalServerError, "InternalError", err.Error())
			return
		}

		writeJSON(c, http.StatusOK, map[string]interface{}{
			"value": memberGroups,
		})
	}
}

// usersDeltaHandler handles GET /v1.0/users/delta
func usersDeltaHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		deltaToken := c.Request.URL.Query().Get("$deltatoken")

		items, newDeltaToken, _, err := st.GetUsersDelta(c.Request.Context(), deltaToken)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "InternalError", err.Error())
			return
		}

		response := map[string]interface{}{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#users",
			"value":          items,
		}

		// On initial sync (no token), we return deltaLink directly (all users in one response)
		// On incremental sync (token provided), we also return deltaLink (only changed users)
		// We don't support pagination for delta queries in this implementation
		response["@odata.deltaLink"] = "https://graph.microsoft.com/v1.0/users/delta?$deltatoken=" + newDeltaToken

		writeJSON(c, http.StatusOK, response)
	}
}

// groupsDeltaHandler handles GET /v1.0/groups/delta
func groupsDeltaHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		deltaToken := c.Request.URL.Query().Get("$deltatoken")

		items, newDeltaToken, _, err := st.GetGroupsDelta(c.Request.Context(), deltaToken)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "InternalError", err.Error())
			return
		}

		response := map[string]interface{}{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#groups",
			"value":          items,
		}

		// On initial sync (no token), we return deltaLink directly (all groups in one response)
		// On incremental sync (token provided), we also return deltaLink (only changed groups)
		// We don't support pagination for delta queries in this implementation
		response["@odata.deltaLink"] = "https://graph.microsoft.com/v1.0/groups/delta?$deltatoken=" + newDeltaToken

		writeJSON(c, http.StatusOK, response)
	}
}