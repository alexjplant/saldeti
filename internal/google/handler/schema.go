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

// listSchemasHandler handles GET /admin/directory/v1/customer/:customerId/schemas
func listSchemasHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		schemas, err := st.ListSchemas(c.Request.Context(), customerID)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to list schemas")
			return
		}
		if schemas == nil {
			schemas = []model.Schema{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"kind":    "admin#directory#schemas",
			"etag":    "\"placeholder\"",
			"schemas": schemas,
		})
	}
}

// getSchemaHandler handles GET /admin/directory/v1/customer/:customerId/schemas/:schemaKey
func getSchemaHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		schemaKey := c.Param("schemaKey")
		schema, err := st.GetSchema(c.Request.Context(), customerID, schemaKey)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Schema not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get schema")
			}
			return
		}
		writeJSON(c, http.StatusOK, schema)
	}
}

// createSchemaHandler handles POST /admin/directory/v1/customer/:customerId/schemas
func createSchemaHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		var schema model.Schema
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&schema); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		created, err := st.CreateSchema(c.Request.Context(), customerID, schema)
		if err != nil {
			if errors.Is(err, store.ErrAlreadyExists) {
				writeError(c, http.StatusConflict, "duplicate", "Schema already exists")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to create schema")
			}
			return
		}
		writeJSON(c, http.StatusOK, created)
	}
}

// updateSchemaHandler handles PUT /admin/directory/v1/customer/:customerId/schemas/:schemaKey
func updateSchemaHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		schemaKey := c.Param("schemaKey")
		var schema model.Schema
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&schema); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.UpdateSchema(c.Request.Context(), customerID, schemaKey, schema)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Schema not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to update schema")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// patchSchemaHandler handles PATCH /admin/directory/v1/customer/:customerId/schemas/:schemaKey
func patchSchemaHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		schemaKey := c.Param("schemaKey")
		var patch map[string]interface{}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&patch); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.PatchSchema(c.Request.Context(), customerID, schemaKey, patch)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Schema not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to patch schema")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// deleteSchemaHandler handles DELETE /admin/directory/v1/customer/:customerId/schemas/:schemaKey
func deleteSchemaHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		schemaKey := c.Param("schemaKey")
		if err := st.DeleteSchema(c.Request.Context(), customerID, schemaKey); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Schema not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to delete schema")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}
