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

// listMobileDevicesHandler handles GET /admin/directory/v1/customer/:customer/devices/mobile
func listMobileDevicesHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		opts := parseGoogleListOptions(c)
		devices, nextPageToken, err := st.ListMobileDevices(c.Request.Context(), customerID, opts)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to list mobile devices")
			return
		}
		if devices == nil {
			devices = []model.MobileDevice{}
		}
		resp := gin.H{
			"kind":  "admin#directory#mobiledevices",
			"etag":  "\"placeholder\"",
			"mobiledevices": devices,
		}
		if nextPageToken != "" {
			resp["nextPageToken"] = nextPageToken
		}
		writeJSON(c, http.StatusOK, resp)
	}
}

// getMobileDeviceHandler handles GET /admin/directory/v1/customer/:customer/devices/mobile/:resourceId
func getMobileDeviceHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		resourceID := c.Param("resourceId")
		device, err := st.GetMobileDevice(c.Request.Context(), customerID, resourceID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Mobile device not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get mobile device")
			}
			return
		}
		writeJSON(c, http.StatusOK, device)
	}
}

// deleteMobileDeviceHandler handles DELETE /admin/directory/v1/customer/:customer/devices/mobile/:resourceId
func deleteMobileDeviceHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		resourceID := c.Param("resourceId")
		if err := st.DeleteMobileDevice(c.Request.Context(), customerID, resourceID); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Mobile device not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to delete mobile device")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// mobileDeviceActionHandler handles POST /admin/directory/v1/customer/:customer/devices/mobile/:resourceId/action
func mobileDeviceActionHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		resourceID := c.Param("resourceId")
		var action model.MobileDeviceAction
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&action); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		if err := st.MobileDeviceAction(c.Request.Context(), customerID, resourceID, action); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Mobile device not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to perform mobile device action")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}