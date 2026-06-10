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

// getCustomerHandler handles GET /admin/directory/v1/customers/:customerKey
func getCustomerHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerKey := c.Param("customerKey")
		customer, err := st.GetCustomer(c.Request.Context(), customerKey)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Customer not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get customer")
			}
			return
		}
		writeJSON(c, http.StatusOK, customer)
	}
}

// updateCustomerHandler handles PUT /admin/directory/v1/customers/:customerKey
func updateCustomerHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerKey := c.Param("customerKey")
		var customer model.Customer
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&customer); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.UpdateCustomer(c.Request.Context(), customerKey, customer)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Customer not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to update customer")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// patchCustomerHandler handles PATCH /admin/directory/v1/customers/:customerKey
func patchCustomerHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerKey := c.Param("customerKey")
		var patch map[string]interface{}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&patch); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.PatchCustomer(c.Request.Context(), customerKey, patch)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Customer not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to patch customer")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}