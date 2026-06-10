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

// listChromeOSDevicesHandler handles GET /admin/directory/v1/customer/:customer/devices/chromeos
func listChromeOSDevicesHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		opts := parseGoogleListOptions(c)
		devices, nextPageToken, err := st.ListChromeOSDevices(c.Request.Context(), customerID, opts)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to list ChromeOS devices")
			return
		}
		if devices == nil {
			devices = []model.ChromeOSDevice{}
		}
		resp := gin.H{
			"kind":  "admin#directory#chromeosdevices",
			"etag":  "\"placeholder\"",
			"chromeosdevices": devices,
		}
		if nextPageToken != "" {
			resp["nextPageToken"] = nextPageToken
		}
		writeJSON(c, http.StatusOK, resp)
	}
}

// getChromeOSDeviceHandler handles GET /admin/directory/v1/customer/:customer/devices/chromeos/:deviceId
func getChromeOSDeviceHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		deviceID := c.Param("deviceId")
		device, err := st.GetChromeOSDevice(c.Request.Context(), customerID, deviceID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "ChromeOS device not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get ChromeOS device")
			}
			return
		}
		writeJSON(c, http.StatusOK, device)
	}
}

// patchChromeOSDeviceHandler handles PATCH /admin/directory/v1/customer/:customer/devices/chromeos/:deviceId
func patchChromeOSDeviceHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		deviceID := c.Param("deviceId")
		var patch map[string]interface{}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&patch); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.PatchChromeOSDevice(c.Request.Context(), customerID, deviceID, patch)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "ChromeOS device not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to patch ChromeOS device")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// updateChromeOSDeviceHandler handles PUT /admin/directory/v1/customer/:customer/devices/chromeos/:deviceId
func updateChromeOSDeviceHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		deviceID := c.Param("deviceId")
		var device model.ChromeOSDevice
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&device); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.UpdateChromeOSDevice(c.Request.Context(), customerID, deviceID, device)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "ChromeOS device not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to update ChromeOS device")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// moveChromeOSDevicesHandler handles POST /admin/directory/v1/customer/:customer/devices/chromeos/moveDevicesToOu
func moveChromeOSDevicesHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		var req struct {
			DeviceIDs    []string `json:"deviceIds"`
			OrgUnitPath  string   `json:"orgUnitPath"`
		}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&req); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		if err := st.MoveChromeOSDevices(c.Request.Context(), customerID, req.DeviceIDs, req.OrgUnitPath); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "ChromeOS device or customer not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to move ChromeOS devices")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// batchChangeChromeOSStatusHandler handles POST /admin/directory/v1/customer/:customer/devices/chromeos:batchChangeStatus
func batchChangeChromeOSStatusHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		var req struct {
			DeviceIDs []string `json:"deviceIds"`
			Action    string   `json:"action"`
		}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&req); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		if err := st.BatchChangeChromeOSStatus(c.Request.Context(), customerID, req.DeviceIDs, req.Action); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "ChromeOS device or customer not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to batch change ChromeOS status")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// countChromeOSDevicesHandler handles GET /admin/directory/v1/customer/:customer/devices/chromeos:countChromeOsDevices
func countChromeOSDevicesHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		count, err := st.CountChromeOSDevices(c.Request.Context(), customerID)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to count ChromeOS devices")
			return
		}
		writeJSON(c, http.StatusOK, gin.H{
			"kind":  "admin#directory#chromeosdevices",
			"count": count,
		})
	}
}

// issueChromeOSCommandHandler handles POST /admin/directory/v1/customer/:customer/devices/chromeos/:deviceId:issueCommand
func issueChromeOSCommandHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		deviceID := c.Param("deviceId")
		var cmd model.ChromeOSCommand
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&cmd); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		result, err := st.IssueChromeOSCommand(c.Request.Context(), customerID, deviceID, cmd)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "ChromeOS device not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to issue ChromeOS command")
			}
			return
		}
		writeJSON(c, http.StatusOK, result)
	}
}

// getChromeOSCommandHandler handles GET /admin/directory/v1/customer/:customer/devices/chromeos/:deviceId/commands/:commandId
func getChromeOSCommandHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		deviceID := c.Param("deviceId")
		commandID := c.Param("commandId")
		result, err := st.GetChromeOSCommand(c.Request.Context(), customerID, deviceID, commandID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "ChromeOS command not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get ChromeOS command")
			}
			return
		}
		writeJSON(c, http.StatusOK, result)
	}
}