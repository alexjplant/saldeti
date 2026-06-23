package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/google/model"
	"github.com/saldeti/saldeti/internal/google/store"
)

// listOrgUnitsHandler handles GET /admin/directory/v1/customer/:customer/orgunits (list only)
func listOrgUnitsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		orgUnits, err := st.ListOrgUnits(c.Request.Context(), customerID)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to list org units")
			return
		}
		if orgUnits == nil {
			orgUnits = []model.OrgUnit{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"kind":              "admin#directory#orgUnits",
			"etag":              "\"placeholder\"",
			"organizationUnits": orgUnits,
		})
	}
}

// getOrgUnitHandler handles GET /admin/directory/v1/customer/:customer/orgunits/*orgUnitPath
// Dispatches to list if the wildcard path is empty (just "/" or empty).
func getOrgUnitHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		orgUnitPath := strings.TrimPrefix(c.Param("orgUnitPath"), "/")
		if orgUnitPath == "" {
			// Wildcard matched empty path — delegate to list
			listOrgUnitsHandler(st)(c)
			return
		}
		ou, err := st.GetOrgUnit(c.Request.Context(), customerID, orgUnitPath)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "OrgUnit not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get org unit")
			}
			return
		}
		writeJSON(c, http.StatusOK, ou)
	}
}

// createOrgUnitHandler handles POST /admin/directory/v1/customer/:customer/orgunits
func createOrgUnitHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		var ou model.OrgUnit
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&ou); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		created, err := st.CreateOrgUnit(c.Request.Context(), customerID, ou)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to create org unit")
			return
		}
		writeJSON(c, http.StatusOK, created)
	}
}

// updateOrgUnitHandler handles PUT /admin/directory/v1/customer/:customer/orgunits/*orgUnitPath
func updateOrgUnitHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		orgUnitPath := strings.TrimPrefix(c.Param("orgUnitPath"), "/")
		var ou model.OrgUnit
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&ou); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.UpdateOrgUnit(c.Request.Context(), customerID, orgUnitPath, ou)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "OrgUnit not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to update org unit")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// patchOrgUnitHandler handles PATCH /admin/directory/v1/customer/:customer/orgunits/*orgUnitPath
func patchOrgUnitHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		orgUnitPath := strings.TrimPrefix(c.Param("orgUnitPath"), "/")
		var patch map[string]interface{}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&patch); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.PatchOrgUnit(c.Request.Context(), customerID, orgUnitPath, patch)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "OrgUnit not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to patch org unit")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// deleteOrgUnitHandler handles DELETE /admin/directory/v1/customer/:customer/orgunits/*orgUnitPath
func deleteOrgUnitHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		orgUnitPath := strings.TrimPrefix(c.Param("orgUnitPath"), "/")
		if err := st.DeleteOrgUnit(c.Request.Context(), customerID, orgUnitPath); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "OrgUnit not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to delete org unit")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}
