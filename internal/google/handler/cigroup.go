package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/google/model"
	"github.com/saldeti/saldeti/internal/google/store"
)

// listCIGroupsHandler handles GET /v1/groups
func listCIGroupsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		opts := parseGoogleListOptions(c)
		groups, nextPageToken, err := st.ListCIGroups(c.Request.Context(), opts)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to list groups")
			}
			return
		}
		if groups == nil {
			groups = []model.CloudIdentityGroup{}
		}
		resp := gin.H{
			"groups": groups,
		}
		if nextPageToken != "" {
			resp["nextPageToken"] = nextPageToken
		}
		writeJSON(c, http.StatusOK, resp)
	}
}

// createCIGroupHandler handles POST /v1/groups
func createCIGroupHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var group model.CloudIdentityGroup
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&group); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		created, err := st.CreateCIGroup(c.Request.Context(), group)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to create group")
			}
			return
		}
		writeJSON(c, http.StatusOK, created)
	}
}

// getCIGroupHandler handles GET /v1/groups/:groupName
func getCIGroupHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupName := c.Param("groupName")
		name := "groups/" + groupName
		group, err := st.GetCIGroup(c.Request.Context(), name)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get group")
			}
			return
		}
		writeJSON(c, http.StatusOK, group)
	}
}

// updateCIGroupHandler handles PATCH /v1/groups/:groupName
func updateCIGroupHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupName := c.Param("groupName")
		name := "groups/" + groupName
		var group model.CloudIdentityGroup
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&group); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.UpdateCIGroup(c.Request.Context(), name, group)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to update group")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// deleteCIGroupHandler handles DELETE /v1/groups/:groupName
func deleteCIGroupHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupName := c.Param("groupName")
		name := "groups/" + groupName
		if err := st.DeleteCIGroup(c.Request.Context(), name); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to delete group")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// lookupCIGroupHandler handles GET /v1/groups:lookup
func lookupCIGroupHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var key model.EntityKey
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&key); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		group, err := st.LookupCIGroup(c.Request.Context(), key)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to lookup group")
			}
			return
		}
		writeJSON(c, http.StatusOK, group)
	}
}

// searchCIGroupsHandler handles GET /v1/groups:search
func searchCIGroupsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("query")
		groups, err := st.SearchCIGroups(c.Request.Context(), query)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to search groups")
			}
			return
		}
		if groups == nil {
			groups = []model.CloudIdentityGroup{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"groups": groups,
		})
	}
}

// getCIGroupSecuritySettingsHandler handles GET /v1/groups/:groupName/securitySettings
func getCIGroupSecuritySettingsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupName := c.Param("groupName")
		name := "groups/" + groupName + "/securitySettings"
		settings, err := st.GetCIGroupSecuritySettings(c.Request.Context(), name)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Security settings not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get security settings")
			}
			return
		}
		writeJSON(c, http.StatusOK, settings)
	}
}

// updateCIGroupSecuritySettingsHandler handles PATCH /v1/groups/:groupName/securitySettings
func updateCIGroupSecuritySettingsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupName := c.Param("groupName")
		name := "groups/" + groupName + "/securitySettings"
		var settings model.SecuritySettings
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&settings); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.UpdateCIGroupSecuritySettings(c.Request.Context(), name, settings)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Security settings not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to update security settings")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}