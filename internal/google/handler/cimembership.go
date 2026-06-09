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

// listCIMembershipsHandler handles GET /v1/groups/:groupName/memberships
func listCIMembershipsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupName := c.Param("groupName")
		parent := "groups/" + groupName
		memberships, err := st.ListCIMemberships(c.Request.Context(), parent)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to list memberships")
			}
			return
		}
		if memberships == nil {
			memberships = []model.Membership{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"memberships": memberships,
		})
	}
}

// getCIMembershipHandler handles GET /v1/groups/:groupName/memberships/:membershipName
func getCIMembershipHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupName := c.Param("groupName")
		membershipName := c.Param("membershipName")
		name := "groups/" + groupName + "/memberships/" + membershipName
		membership, err := st.GetCIMembership(c.Request.Context(), name)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Membership not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get membership")
			}
			return
		}
		writeJSON(c, http.StatusOK, membership)
	}
}

// createCIMembershipHandler handles POST /v1/groups/:groupName/memberships
func createCIMembershipHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupName := c.Param("groupName")
		parent := "groups/" + groupName
		var membership model.Membership
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&membership); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		created, err := st.CreateCIMembership(c.Request.Context(), parent, membership)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to create membership")
			}
			return
		}
		writeJSON(c, http.StatusOK, created)
	}
}

// deleteCIMembershipHandler handles DELETE /v1/groups/:groupName/memberships/:membershipName
func deleteCIMembershipHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupName := c.Param("groupName")
		membershipName := c.Param("membershipName")
		name := "groups/" + groupName + "/memberships/" + membershipName
		if err := st.DeleteCIMembership(c.Request.Context(), name); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Membership not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to delete membership")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// lookupCIMembershipHandler handles POST /v1/groups/:groupName/memberships:lookup
func lookupCIMembershipHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupName := c.Param("groupName")
		parent := "groups/" + groupName
		var key model.EntityKey
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&key); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		membership, err := st.LookupCIMembership(c.Request.Context(), parent, key)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Membership not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to lookup membership")
			}
			return
		}
		writeJSON(c, http.StatusOK, membership)
	}
}

// modifyMembershipRolesHandler handles POST /v1/groups/:groupName/memberships/:membershipName/modifyMembershipRoles
func modifyMembershipRolesHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupName := c.Param("groupName")
		membershipName := c.Param("membershipName")
		name := "groups/" + groupName + "/memberships/" + membershipName
		var roles model.ModifyMembershipRolesRequest
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&roles); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		membership, err := st.ModifyMembershipRoles(c.Request.Context(), name, roles)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Membership not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to modify membership roles")
			}
			return
		}
		writeJSON(c, http.StatusOK, membership)
	}
}

// checkTransitiveMembershipHandler handles POST /v1/groups/:groupName/memberships:checkTransitiveMembership
func checkTransitiveMembershipHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupName := c.Param("groupName")
		parent := "groups/" + groupName
		var key model.EntityKey
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&key); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		result, err := st.CheckTransitiveMembership(c.Request.Context(), parent, key)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to check transitive membership")
			}
			return
		}
		writeJSON(c, http.StatusOK, gin.H{
			"membershipsExist": result,
		})
	}
}

// getMembershipGraphHandler handles GET /v1/groups/:groupName/memberships:getMembershipGraph
func getMembershipGraphHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupName := c.Param("groupName")
		parent := "groups/" + groupName
		query := c.Query("query")
		graph, err := st.GetMembershipGraph(c.Request.Context(), parent, query)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get membership graph")
			}
			return
		}
		writeJSON(c, http.StatusOK, graph)
	}
}

// searchTransitiveGroupsHandler handles GET /v1/groups/:groupName/memberships:searchTransitiveGroups
func searchTransitiveGroupsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupName := c.Param("groupName")
		parent := "groups/" + groupName
		query := c.Query("query")
		groups, err := st.SearchTransitiveGroups(c.Request.Context(), parent, query)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to search transitive groups")
			}
			return
		}
		if groups == nil {
			groups = []model.CloudIdentityGroup{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"memberships": groups,
		})
	}
}

// searchTransitiveMembershipsHandler handles GET /v1/groups/:groupName/memberships:searchTransitiveMemberships
func searchTransitiveMembershipsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupName := c.Param("groupName")
		parent := "groups/" + groupName
		query := c.Query("query")
		memberships, err := st.SearchTransitiveMemberships(c.Request.Context(), parent, query)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to search transitive memberships")
			}
			return
		}
		if memberships == nil {
			memberships = []model.Membership{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"memberships": memberships,
		})
	}
}

// searchDirectGroupsHandler handles GET /v1/groups/:groupName/memberships:searchDirectGroups
func searchDirectGroupsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupName := c.Param("groupName")
		parent := "groups/" + groupName
		query := c.Query("query")
		memberships, err := st.SearchDirectGroups(c.Request.Context(), parent, query)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Group not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to search direct groups")
			}
			return
		}
		if memberships == nil {
			memberships = []model.Membership{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"memberships": memberships,
		})
	}
}