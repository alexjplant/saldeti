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

// listCIDevicesHandler handles GET /v1/devices
func listCIDevicesHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		opts := parseGoogleListOptions(c)
		devices, nextPageToken, err := st.ListCIDevices(c.Request.Context(), opts)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Device not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to list devices")
			}
			return
		}
		if devices == nil {
			devices = []model.CloudIdentityDevice{}
		}
		resp := gin.H{
			"devices": devices,
		}
		if nextPageToken != "" {
			resp["nextPageToken"] = nextPageToken
		}
		writeJSON(c, http.StatusOK, resp)
	}
}

// createCIDeviceHandler handles POST /v1/devices
func createCIDeviceHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var device model.CloudIdentityDevice
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&device); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		created, err := st.CreateCIDevice(c.Request.Context(), device)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Device not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to create device")
			}
			return
		}
		writeJSON(c, http.StatusOK, created)
	}
}

// getCIDeviceHandler handles GET /v1/devices/:deviceName
func getCIDeviceHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceName := c.Param("deviceName")
		name := "devices/" + deviceName
		device, err := st.GetCIDevice(c.Request.Context(), name)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Device not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get device")
			}
			return
		}
		writeJSON(c, http.StatusOK, device)
	}
}

// deleteCIDeviceHandler handles DELETE /v1/devices/:deviceName
func deleteCIDeviceHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceName := c.Param("deviceName")
		name := "devices/" + deviceName
		if err := st.DeleteCIDevice(c.Request.Context(), name); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Device not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to delete device")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// wipeCIDeviceHandler handles POST /v1/devices/:deviceName/wipe
func wipeCIDeviceHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceName := c.Param("deviceName")
		name := "devices/" + deviceName
		if err := st.WipeCIDevice(c.Request.Context(), name); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Device not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to wipe device")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// cancelWipeCIDeviceHandler handles POST /v1/devices/:deviceName/cancelWipe
func cancelWipeCIDeviceHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceName := c.Param("deviceName")
		name := "devices/" + deviceName
		if err := st.CancelWipeCIDevice(c.Request.Context(), name); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Device not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to cancel wipe device")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// listDeviceUsersHandler handles GET /v1/devices/:deviceName/deviceUsers
func listDeviceUsersHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceName := c.Param("deviceName")
		parent := "devices/" + deviceName
		users, err := st.ListDeviceUsers(c.Request.Context(), parent)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Device not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to list device users")
			}
			return
		}
		if users == nil {
			users = []model.DeviceUser{}
		}
		writeJSON(c, http.StatusOK, gin.H{
			"deviceUsers": users,
		})
	}
}

// getDeviceUserHandler handles GET /v1/devices/:deviceName/deviceUsers/:deviceUserName
func getDeviceUserHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceName := c.Param("deviceName")
		deviceUserName := c.Param("deviceUserName")
		name := "devices/" + deviceName + "/deviceUsers/" + deviceUserName
		user, err := st.GetDeviceUser(c.Request.Context(), name)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Device user not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get device user")
			}
			return
		}
		writeJSON(c, http.StatusOK, user)
	}
}

// deleteDeviceUserHandler handles DELETE /v1/devices/:deviceName/deviceUsers/:deviceUserName
func deleteDeviceUserHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceName := c.Param("deviceName")
		deviceUserName := c.Param("deviceUserName")
		name := "devices/" + deviceName + "/deviceUsers/" + deviceUserName
		if err := st.DeleteDeviceUser(c.Request.Context(), name); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Device user not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to delete device user")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// approveDeviceUserHandler handles POST /v1/devices/:deviceName/deviceUsers/:deviceUserName/approve
func approveDeviceUserHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceName := c.Param("deviceName")
		deviceUserName := c.Param("deviceUserName")
		name := "devices/" + deviceName + "/deviceUsers/" + deviceUserName
		if err := st.ApproveDeviceUser(c.Request.Context(), name); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Device user not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to approve device user")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// blockDeviceUserHandler handles POST /v1/devices/:deviceName/deviceUsers/:deviceUserName/block
func blockDeviceUserHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceName := c.Param("deviceName")
		deviceUserName := c.Param("deviceUserName")
		name := "devices/" + deviceName + "/deviceUsers/" + deviceUserName
		if err := st.BlockDeviceUser(c.Request.Context(), name); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Device user not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to block device user")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// wipeDeviceUserHandler handles POST /v1/devices/:deviceName/deviceUsers/:deviceUserName/wipe
func wipeDeviceUserHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceName := c.Param("deviceName")
		deviceUserName := c.Param("deviceUserName")
		name := "devices/" + deviceName + "/deviceUsers/" + deviceUserName
		if err := st.WipeDeviceUser(c.Request.Context(), name); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Device user not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to wipe device user")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// cancelWipeDeviceUserHandler handles POST /v1/devices/:deviceName/deviceUsers/:deviceUserName/cancelWipe
func cancelWipeDeviceUserHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceName := c.Param("deviceName")
		deviceUserName := c.Param("deviceUserName")
		name := "devices/" + deviceName + "/deviceUsers/" + deviceUserName
		if err := st.CancelWipeDeviceUser(c.Request.Context(), name); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Device user not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to cancel wipe device user")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// lookupDeviceUserHandler handles GET /v1/devices/:deviceName/deviceUsers:lookup
func lookupDeviceUserHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceName := c.Param("deviceName")
		parent := "devices/" + deviceName
		users, err := st.LookupDeviceUser(c.Request.Context(), parent)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Device not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to lookup device user")
			}
			return
		}
		writeJSON(c, http.StatusOK, gin.H{
			"deviceUsers": users,
		})
	}
}
