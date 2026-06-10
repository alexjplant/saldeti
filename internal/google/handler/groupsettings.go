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

// getGroupSettingsHandler handles GET /groups/v1/groups/:groupUniqueId
func getGroupSettingsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupUniqueId := c.Param("groupUniqueId")
		settings, err := st.GetGroupSettings(c.Request.Context(), groupUniqueId)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group settings not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get group settings")
			}
			return
		}
		writeJSON(c, http.StatusOK, settings)
	}
}

// updateGroupSettingsHandler handles PUT /groups/v1/groups/:groupUniqueId
func updateGroupSettingsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupUniqueId := c.Param("groupUniqueId")
		var settings model.GroupSettings
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&settings); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.UpdateGroupSettings(c.Request.Context(), groupUniqueId, settings)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group settings not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to update group settings")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// patchGroupSettingsHandler handles PATCH /groups/v1/groups/:groupUniqueId
func patchGroupSettingsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupUniqueId := c.Param("groupUniqueId")
		var patch map[string]interface{}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&patch); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.PatchGroupSettings(c.Request.Context(), groupUniqueId, patch)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group settings not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to patch group settings")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}