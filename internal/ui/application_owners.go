package ui

import (
	"github.com/gin-gonic/gin"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/rs/zerolog/log"
)

func ApplicationAddOwnerHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := c.PostForm("userId")

		if userID == "" {
			h.handleAppOwnersResponse(c, id, FlashDanger, "No user selected")
			return
		}

		// Add owner via SDK - use raw HTTP since the SDK endpoint structure may differ
		refBody := models.NewReferenceCreate()
		odataId := h.baseURL + "/v1.0/users/" + userID
		refBody.SetOdataId(&odataId)

		// Use the applications owners ref endpoint
		err := h.client.Applications().ByApplicationId(id).Owners().Ref().Post(c.Request.Context(), refBody, nil)
		if err != nil {
			log.Error().Err(err).Msg("Failed to add application owner")
			h.handleAppOwnersResponse(c, id, FlashDanger, "Failed to add owner. Please try again.")
			return
		}

		h.handleAppOwnersResponse(c, id, FlashSuccess, "Owner added successfully")
	}
}

func ApplicationRemoveOwnerHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ownerID := c.Param("ownerId")

		if err := h.client.Applications().ByApplicationId(id).Owners().ByDirectoryObjectId(ownerID).Ref().Delete(c.Request.Context(), nil); err != nil {
			log.Error().Err(err).Msg("Failed to remove application owner")
			h.handleAppOwnersResponse(c, id, FlashDanger, "Failed to remove owner. Please try again.")
		} else {
			h.handleAppOwnersResponse(c, id, FlashSuccess, "Owner removed successfully")
		}
	}
}
