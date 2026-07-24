package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/entra/model"
	"github.com/saldeti/saldeti/internal/entra/store"
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
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&req); err != nil {
			writeError(c, http.StatusBadRequest, "InvalidRequest", "Invalid JSON body")
			return
		}

		if req.AddLicenses == nil {
			req.AddLicenses = []model.LicenseAssignment{}
		}
		if req.RemoveLicenses == nil {
			req.RemoveLicenses = []string{}
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

		response, err := buildEntityResponse("https://graph.microsoft.com/v1.0/$metadata#users/$entity", updatedUser)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
			return
		}

		writeJSON(c, http.StatusOK, response)
	}
}
