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

// listTransferApplicationsHandler handles GET /admin/datatransfer/v1/applications
func listTransferApplicationsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		apps, err := st.ListTransferApplications(c.Request.Context())
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to list transfer applications")
			return
		}
		if apps == nil {
			apps = []model.TransferApplication{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"kind":         "admin#datatransfer#applicationsList",
			"etag":         "\"placeholder\"",
			"applications": apps,
		})
	}
}

// getTransferApplicationHandler handles GET /admin/datatransfer/v1/applications/:applicationId
func getTransferApplicationHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		applicationID := c.Param("applicationId")
		app, err := st.GetTransferApplication(c.Request.Context(), applicationID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Transfer application not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get transfer application")
			}
			return
		}
		writeJSON(c, http.StatusOK, app)
	}
}

// listTransfersHandler handles GET /admin/datatransfer/v1/transfers
func listTransfersHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		transfers, err := st.ListTransfers(c.Request.Context())
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to list transfers")
			return
		}
		if transfers == nil {
			transfers = []model.DataTransfer{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"kind":          "admin#datatransfer#dataTransfersList",
			"etag":          "\"placeholder\"",
			"dataTransfers": transfers,
		})
	}
}

// getTransferHandler handles GET /admin/datatransfer/v1/transfers/:dataTransferId
func getTransferHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		transferID := c.Param("dataTransferId")
		transfer, err := st.GetTransfer(c.Request.Context(), transferID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Data transfer not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get data transfer")
			}
			return
		}
		writeJSON(c, http.StatusOK, transfer)
	}
}

// createTransferHandler handles POST /admin/datatransfer/v1/transfers
func createTransferHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var transfer model.DataTransfer
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&transfer); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		created, err := st.CreateTransfer(c.Request.Context(), transfer)
		if err != nil {
			if errors.Is(err, store.ErrAlreadyExists) {
				writeError(c, http.StatusConflict, "duplicate", "Data transfer already exists")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to create data transfer")
			}
			return
		}
		writeJSON(c, http.StatusOK, created)
	}
}
