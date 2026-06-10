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

// listGroupsHandler handles GET /admin/directory/v1/groups
func listGroupsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		opts := parseGoogleListOptions(c)
		groups, nextPageToken, err := st.ListGroups(c.Request.Context(), opts)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to list groups")
			return
		}
		if groups == nil {
			groups = []model.Group{}
		}
		resp := gin.H{
			"kind":   "admin#directory#groups",
			"etag":   "\"placeholder\"",
			"groups": groups,
		}
		if nextPageToken != "" {
			resp["nextPageToken"] = nextPageToken
		}
		writeJSON(c, http.StatusOK, resp)
	}
}

// getGroupHandler handles GET /admin/directory/v1/groups/:groupKey
func getGroupHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupKey := c.Param("groupKey")
		group, err := st.GetGroup(c.Request.Context(), groupKey)
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

// createGroupHandler handles POST /admin/directory/v1/groups
func createGroupHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var group model.Group
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&group); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		created, err := st.CreateGroup(c.Request.Context(), group)
		if err != nil {
			if errors.Is(err, store.ErrAlreadyExists) {
				writeError(c, http.StatusConflict, "duplicate", "Group already exists")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to create group")
			}
			return
		}
		writeJSON(c, http.StatusOK, created)
	}
}

// updateGroupHandler handles PUT /admin/directory/v1/groups/:groupKey
func updateGroupHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupKey := c.Param("groupKey")
		var group model.Group
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&group); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.UpdateGroup(c.Request.Context(), groupKey, group)
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

// patchGroupHandler handles PATCH /admin/directory/v1/groups/:groupKey
func patchGroupHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupKey := c.Param("groupKey")
		var patch map[string]interface{}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&patch); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.PatchGroup(c.Request.Context(), groupKey, patch)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to patch group")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// deleteGroupHandler handles DELETE /admin/directory/v1/groups/:groupKey
func deleteGroupHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupKey := c.Param("groupKey")
		if err := st.DeleteGroup(c.Request.Context(), groupKey); err != nil {
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

// addGroupAliasHandler handles POST /admin/directory/v1/groups/:groupKey/aliases
func addGroupAliasHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupKey := c.Param("groupKey")
		var req struct {
			Alias string `json:"alias"`
		}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&req); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		if req.Alias == "" {
			writeError(c, http.StatusBadRequest, "invalid", "alias is required")
			return
		}
		if err := st.AddGroupAlias(c.Request.Context(), groupKey, req.Alias); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to add alias")
			}
			return
		}
		writeJSON(c, http.StatusOK, gin.H{
			"kind":  "admin#directory#alias",
			"etag":  "\"placeholder\"",
			"alias": req.Alias,
		})
	}
}

// listGroupAliasesHandler handles GET /admin/directory/v1/groups/:groupKey/aliases
func listGroupAliasesHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupKey := c.Param("groupKey")
		aliases, err := st.ListGroupAliases(c.Request.Context(), groupKey)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to list aliases")
			}
			return
		}
		if aliases == nil {
			aliases = []string{}
		}
		aliasItems := []gin.H{}
		for _, a := range aliases {
			aliasItems = append(aliasItems, gin.H{
				"kind":  "admin#directory#alias",
				"etag":  "\"placeholder\"",
				"alias": a,
			})
		}
		writeJSON(c, http.StatusOK, gin.H{
			"kind":    "admin#directory#aliases",
			"etag":    "\"placeholder\"",
			"aliases": aliasItems,
		})
	}
}

// removeGroupAliasHandler handles DELETE /admin/directory/v1/groups/:groupKey/aliases/:alias
func removeGroupAliasHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupKey := c.Param("groupKey")
		alias := c.Param("alias")
		if err := st.RemoveGroupAlias(c.Request.Context(), groupKey, alias); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Resource not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to remove alias")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}
