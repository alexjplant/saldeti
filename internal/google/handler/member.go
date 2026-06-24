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

// listMembersHandler handles GET /admin/directory/v1/groups/:groupKey/members
func listMembersHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupKey := c.Param("groupKey")
		opts := parseGoogleListOptions(c)
		members, nextPageToken, err := st.ListMembers(c.Request.Context(), groupKey, opts)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to list members")
			}
			return
		}
		if members == nil {
			members = []model.Member{}
		}
		resp := gin.H{
			"kind":    "admin#directory#members",
			"etag":    "\"placeholder\"",
			"members": members,
		}
		if nextPageToken != "" {
			resp["nextPageToken"] = nextPageToken
		}
		writeJSON(c, http.StatusOK, resp)
	}
}

// getMemberHandler handles GET /admin/directory/v1/groups/:groupKey/members/:memberKey
func getMemberHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupKey := c.Param("groupKey")
		memberKey := c.Param("memberKey")
		member, err := st.GetMember(c.Request.Context(), groupKey, memberKey)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Member not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get member")
			}
			return
		}
		writeJSON(c, http.StatusOK, member)
	}
}

// addMemberHandler handles POST /admin/directory/v1/groups/:groupKey/members
func addMemberHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupKey := c.Param("groupKey")
		var member model.Member
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&member); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		created, err := st.AddMember(c.Request.Context(), groupKey, member)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to add member")
			}
			return
		}
		writeJSON(c, http.StatusOK, created)
	}
}

// updateMemberHandler handles PUT /admin/directory/v1/groups/:groupKey/members/:memberKey
func updateMemberHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupKey := c.Param("groupKey")
		memberKey := c.Param("memberKey")
		var member model.Member
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&member); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.UpdateMember(c.Request.Context(), groupKey, memberKey, member)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Member not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to update member")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// patchMemberHandler handles PATCH /admin/directory/v1/groups/:groupKey/members/:memberKey
// The store has no PatchMember, so we get-merge-update via JSON round-trip.
func patchMemberHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupKey := c.Param("groupKey")
		memberKey := c.Param("memberKey")
		var patch map[string]any
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&patch); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		existing, err := st.GetMember(c.Request.Context(), groupKey, memberKey)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Member not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get member")
			}
			return
		}
		// JSON round-trip merge
		existingJSON, err := json.Marshal(existing)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to patch member")
			return
		}
		var existingMap map[string]any
		if err := json.Unmarshal(existingJSON, &existingMap); err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to patch member")
			return
		}
		for k, v := range patch {
			existingMap[k] = v
		}
		mergedJSON, err := json.Marshal(existingMap)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to patch member")
			return
		}
		var merged model.Member
		if err := json.Unmarshal(mergedJSON, &merged); err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to patch member")
			return
		}
		updated, err := st.UpdateMember(c.Request.Context(), groupKey, memberKey, merged)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Member not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to update member")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// removeMemberHandler handles DELETE /admin/directory/v1/groups/:groupKey/members/:memberKey
func removeMemberHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupKey := c.Param("groupKey")
		memberKey := c.Param("memberKey")
		if err := st.RemoveMember(c.Request.Context(), groupKey, memberKey); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Member not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to remove member")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// hasMemberHandler handles GET /admin/directory/v1/groups/:groupKey/hasMember/:memberKey
func hasMemberHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupKey := c.Param("groupKey")
		memberKey := c.Param("memberKey")
		isMember, err := st.HasMember(c.Request.Context(), groupKey, memberKey)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to check membership")
			}
			return
		}
		writeJSON(c, http.StatusOK, gin.H{
			"isMember": isMember,
		})
	}
}
