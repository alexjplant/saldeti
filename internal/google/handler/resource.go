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

// Calendar Resources Handlers

// listCalendarResourcesHandler handles GET /admin/directory/v1/resources/calendars
func listCalendarResourcesHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		resources, err := st.ListCalendarResources(c.Request.Context(), customerID)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to list calendar resources")
			return
		}
		if resources == nil {
			resources = []model.CalendarResource{}
		}
		resp := gin.H{
			"kind":  "admin#directory#resources#calendars#CalendarResources",
			"etag":  "\"placeholder\"",
			"items": resources,
		}
		writeJSON(c, http.StatusOK, resp)
	}
}

// getCalendarResourceHandler handles GET /admin/directory/v1/resources/calendars/:resourceId
func getCalendarResourceHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		resourceID := c.Param("resourceId")
		resource, err := st.GetCalendarResource(c.Request.Context(), customerID, resourceID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Calendar resource not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get calendar resource")
			}
			return
		}
		writeJSON(c, http.StatusOK, resource)
	}
}

// createCalendarResourceHandler handles POST /admin/directory/v1/resources/calendars
func createCalendarResourceHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		var resource model.CalendarResource
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&resource); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		created, err := st.CreateCalendarResource(c.Request.Context(), customerID, resource)
		if err != nil {
			if errors.Is(err, store.ErrAlreadyExists) {
				writeError(c, http.StatusConflict, "duplicate", "Calendar resource already exists")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to create calendar resource")
			}
			return
		}
		writeJSON(c, http.StatusOK, created)
	}
}

// updateCalendarResourceHandler handles PUT /admin/directory/v1/resources/calendars/:resourceId
func updateCalendarResourceHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		resourceID := c.Param("resourceId")
		var resource model.CalendarResource
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&resource); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.UpdateCalendarResource(c.Request.Context(), customerID, resourceID, resource)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Calendar resource not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to update calendar resource")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// patchCalendarResourceHandler handles PATCH /admin/directory/v1/resources/calendars/:resourceId
func patchCalendarResourceHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		resourceID := c.Param("resourceId")
		var patch map[string]interface{}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&patch); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.PatchCalendarResource(c.Request.Context(), customerID, resourceID, patch)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Calendar resource not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to patch calendar resource")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// deleteCalendarResourceHandler handles DELETE /admin/directory/v1/resources/calendars/:resourceId
func deleteCalendarResourceHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		resourceID := c.Param("resourceId")
		if err := st.DeleteCalendarResource(c.Request.Context(), customerID, resourceID); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Calendar resource not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to delete calendar resource")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// Buildings Handlers

// listBuildingsHandler handles GET /admin/directory/v1/resources/buildings
func listBuildingsHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		buildings, err := st.ListBuildings(c.Request.Context(), customerID)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to list buildings")
			return
		}
		if buildings == nil {
			buildings = []model.Building{}
		}
		resp := gin.H{
			"kind":      "admin#directory#resources#buildings#Buildings",
			"etag":      "\"placeholder\"",
			"buildings": buildings,
		}
		writeJSON(c, http.StatusOK, resp)
	}
}

// getBuildingHandler handles GET /admin/directory/v1/resources/buildings/:buildingId
func getBuildingHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		buildingID := c.Param("buildingId")
		building, err := st.GetBuilding(c.Request.Context(), customerID, buildingID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Building not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get building")
			}
			return
		}
		writeJSON(c, http.StatusOK, building)
	}
}

// createBuildingHandler handles POST /admin/directory/v1/resources/buildings
func createBuildingHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		var building model.Building
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&building); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		created, err := st.CreateBuilding(c.Request.Context(), customerID, building)
		if err != nil {
			if errors.Is(err, store.ErrAlreadyExists) {
				writeError(c, http.StatusConflict, "duplicate", "Building already exists")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to create building")
			}
			return
		}
		writeJSON(c, http.StatusOK, created)
	}
}

// updateBuildingHandler handles PUT /admin/directory/v1/resources/buildings/:buildingId
func updateBuildingHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		buildingID := c.Param("buildingId")
		var building model.Building
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&building); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.UpdateBuilding(c.Request.Context(), customerID, buildingID, building)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Building not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to update building")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// patchBuildingHandler handles PATCH /admin/directory/v1/resources/buildings/:buildingId
func patchBuildingHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		buildingID := c.Param("buildingId")
		var patch map[string]interface{}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&patch); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.PatchBuilding(c.Request.Context(), customerID, buildingID, patch)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Building not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to patch building")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// deleteBuildingHandler handles DELETE /admin/directory/v1/resources/buildings/:buildingId
func deleteBuildingHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		buildingID := c.Param("buildingId")
		if err := st.DeleteBuilding(c.Request.Context(), customerID, buildingID); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Building not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to delete building")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// Features Handlers

// listFeaturesHandler handles GET /admin/directory/v1/resources/features
func listFeaturesHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		features, err := st.ListFeatures(c.Request.Context(), customerID)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "backendError", "Failed to list features")
			return
		}
		if features == nil {
			features = []model.Feature{}
		}
		resp := gin.H{
			"kind":     "admin#directory#resources#features#Features",
			"etag":     "\"placeholder\"",
			"features": features,
		}
		writeJSON(c, http.StatusOK, resp)
	}
}

// getFeatureHandler handles GET /admin/directory/v1/resources/features/:featureKey
func getFeatureHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		featureKey := c.Param("featureKey")
		feature, err := st.GetFeature(c.Request.Context(), customerID, featureKey)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Feature not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to get feature")
			}
			return
		}
		writeJSON(c, http.StatusOK, feature)
	}
}

// createFeatureHandler handles POST /admin/directory/v1/resources/features
func createFeatureHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		var feature model.Feature
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&feature); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		created, err := st.CreateFeature(c.Request.Context(), customerID, feature)
		if err != nil {
			if errors.Is(err, store.ErrAlreadyExists) {
				writeError(c, http.StatusConflict, "duplicate", "Feature already exists")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to create feature")
			}
			return
		}
		writeJSON(c, http.StatusOK, created)
	}
}

// updateFeatureHandler handles PUT /admin/directory/v1/resources/features/:featureKey
func updateFeatureHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		featureKey := c.Param("featureKey")
		var feature model.Feature
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&feature); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.UpdateFeature(c.Request.Context(), customerID, featureKey, feature)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Feature not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to update feature")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// patchFeatureHandler handles PATCH /admin/directory/v1/resources/features/:featureKey
func patchFeatureHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		featureKey := c.Param("featureKey")
		var patch map[string]interface{}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&patch); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		updated, err := st.PatchFeature(c.Request.Context(), customerID, featureKey, patch)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Feature not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to patch feature")
			}
			return
		}
		writeJSON(c, http.StatusOK, updated)
	}
}

// deleteFeatureHandler handles DELETE /admin/directory/v1/resources/features/:featureKey
func deleteFeatureHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		featureKey := c.Param("featureKey")
		if err := st.DeleteFeature(c.Request.Context(), customerID, featureKey); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Feature not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to delete feature")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}

// renameFeatureHandler handles POST /admin/directory/v1/resources/features/:featureKey/rename
func renameFeatureHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("customer")
		featureKey := c.Param("featureKey")
		var req struct {
			NewName string `json:"newName"`
		}
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, maxBodyBytes)).Decode(&req); err != nil {
			writeError(c, http.StatusBadRequest, "invalid", "Invalid JSON body")
			return
		}
		if err := st.RenameFeature(c.Request.Context(), customerID, featureKey, req.NewName); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(c, http.StatusNotFound, "notFound", "Feature not found")
			} else {
				writeError(c, http.StatusInternalServerError, "backendError", "Failed to rename feature")
			}
			return
		}
		c.Status(http.StatusOK)
	}
}