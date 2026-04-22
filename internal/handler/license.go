package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/model"
	"github.com/saldeti/saldeti/internal/store"
)

// listSubscribedSkusHandler handles GET /v1.0/subscribedSkus
func listSubscribedSkusHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		skus, err := st.ListSubscribedSkus(c.Request.Context())
		if err != nil {
			writeError(c, http.StatusInternalServerError, "InternalError", "Failed to list subscribed SKUs")
			return
		}

		if skus == nil {
			skus = []model.SubscribedSku{}
		}

		response := model.ListResponse{
			Context: "https://graph.microsoft.com/v1.0/$metadata#subscribedSkus",
			Value:   skus,
		}

		writeJSON(c, http.StatusOK, response)
	}
}

// assignLicenseHandler handles POST /v1.0/users/{id}/assignLicense
func assignLicenseHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "User ID is required")
			return
		}

		var req model.LicenseAssignmentRequest
		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		if req.AddLicenses == nil {
			req.AddLicenses = []model.LicenseAssignment{}
		}
		if req.RemoveLicenses == nil {
			req.RemoveLicenses = []model.LicenseRemoval{}
		}

		updatedUser, err := st.AssignLicense(c.Request.Context(), id, req.AddLicenses, req.RemoveLicenses)
		if err != nil {
			if errors.Is(err, store.ErrUserNotFound) {
				writeError(c, http.StatusNotFound, "ResourceNotFound", "User not found")
			} else {
				writeError(c, http.StatusInternalServerError, "InternalError", "Failed to assign license")
			}
			return
		}

		// Build response with OData context
		response := map[string]interface{}{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#users/$entity",
		}

		userJSON, err := json.Marshal(updatedUser)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}
		var userMap map[string]interface{}
		if err := json.Unmarshal(userJSON, &userMap); err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		for k, v := range userMap {
			response[k] = v
		}

		writeJSON(c, http.StatusOK, response)
	}
}
