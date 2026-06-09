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

// listRolesHandler handles GET /admin/directory/v1/customer/:customer/roles
func listRolesHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		roles, err := st.ListRoles(c.Request.Context(), customerID)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to list roles")
			return
		}
		if roles == nil {
			roles = []model.Role{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"kind":  "admin#directory#roles",
			"etag":  "\"placeholder\"",
			"items": roles,
		})
	}
}

// getRoleHandler handles GET /admin/directory/v1/customer/:customer/roles/:roleId
func getRoleHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		roleID := c.Param("roleId")
		role, err := st.GetRole(c.Request.Context(), customerID, roleID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Role not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get role")
			}
			return
		}
		writeJSON(c, http.StatusOK, role)
	}
}

// createRoleHandler handles POST /admin/directory/v1/customer/:customer/roles
func createRoleHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		var role model.Role
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&role); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		created, err := st.CreateRole(c.Request.Context(), customerID, role)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to create role")
			return
		}
		writeJSON(c, http.StatusOK, created)
	}
}

// updateRoleHandler handles PUT /admin/directory/v1/customer/:customer/roles/:roleId
func updateRoleHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		roleID := c.Param("roleId")
		var role model.Role
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&role); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.UpdateRole(c.Request.Context(), customerID, roleID, role)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Role not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to update role")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// patchRoleHandler handles PATCH /admin/directory/v1/customer/:customer/roles/:roleId
func patchRoleHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		roleID := c.Param("roleId")
		var patch map[string]interface{}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&patch); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.PatchRole(c.Request.Context(), customerID, roleID, patch)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Role not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to patch role")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// deleteRoleHandler handles DELETE /admin/directory/v1/customer/:customer/roles/:roleId
func deleteRoleHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		roleID := c.Param("roleId")
		if err := st.DeleteRole(c.Request.Context(), customerID, roleID); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Role not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to delete role")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// listRoleAssignmentsHandler handles GET /admin/directory/v1/customer/:customer/roleassignments
func listRoleAssignmentsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		assignments, err := st.ListRoleAssignments(c.Request.Context(), customerID)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to list role assignments")
			return
		}
		if assignments == nil {
			assignments = []model.RoleAssignment{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"kind":  "admin#directory#roleAssignments",
			"etag":  "\"placeholder\"",
			"items": assignments,
		})
	}
}

// getRoleAssignmentHandler handles GET /admin/directory/v1/customer/:customer/roleassignments/:roleAssignmentId
func getRoleAssignmentHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		assignmentID := c.Param("roleAssignmentId")
		ra, err := st.GetRoleAssignment(c.Request.Context(), customerID, assignmentID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Role assignment not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get role assignment")
			}
			return
		}
		writeJSON(c, http.StatusOK, ra)
	}
}

// createRoleAssignmentHandler handles POST /admin/directory/v1/customer/:customer/roleassignments
func createRoleAssignmentHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		var ra model.RoleAssignment
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&ra); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		created, err := st.CreateRoleAssignment(c.Request.Context(), customerID, ra)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to create role assignment")
			return
		}
		writeJSON(c, http.StatusOK, created)
	}
}

// deleteRoleAssignmentHandler handles DELETE /admin/directory/v1/customer/:customer/roleassignments/:roleAssignmentId
func deleteRoleAssignmentHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		assignmentID := c.Param("roleAssignmentId")
		if err := st.DeleteRoleAssignment(c.Request.Context(), customerID, assignmentID); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Role assignment not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to delete role assignment")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// listPrivilegesHandler handles GET /admin/directory/v1/customer/:customer/roles/ALL/privileges
func listPrivilegesHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		privileges, err := st.ListPrivileges(c.Request.Context(), customerID)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to list privileges")
			return
		}
		if privileges == nil {
			privileges = []model.Privilege{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"kind":  "admin#directory#privileges",
			"etag":  "\"placeholder\"",
			"items": privileges,
		})
	}
}
