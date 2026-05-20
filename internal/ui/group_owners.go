package ui

import (
	"github.com/gin-gonic/gin"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/rs/zerolog/log"
)

func GroupAddOwnerHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := c.PostForm("userId")

		if userID == "" {
			h.handleOwnersResponse(c, id, FlashDanger, "No user selected")
			return
		}

		// Add owner via SDK
		refBody := models.NewReferenceCreate()
		odataId := h.baseURL + "/v1.0/users/" + userID
		refBody.SetOdataId(&odataId)

		err := h.client.Groups().ByGroupId(id).Owners().Ref().Post(c.Request.Context(), refBody, nil)
		if err != nil {
			log.Error().Err(err).Msg("Failed to add group owner")
			h.handleOwnersResponse(c, id, FlashDanger, "Failed to add owner. Please try again.")
			return
		}

		h.handleOwnersResponse(c, id, FlashSuccess, "Owner added successfully")
	}
}

func GroupRemoveOwnerHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ownerID := c.Param("ownerId")

		if err := h.client.Groups().ByGroupId(id).Owners().ByDirectoryObjectId(ownerID).Ref().Delete(c.Request.Context(), nil); err != nil {
			log.Error().Err(err).Msg("Failed to remove group owner")
			h.handleOwnersResponse(c, id, FlashDanger, "Failed to remove owner. Please try again.")
		} else {
			h.handleOwnersResponse(c, id, FlashSuccess, "Owner removed successfully")
		}
	}
}
