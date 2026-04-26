package ui

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
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
			h.handleAppOwnersResponse(c, id, FlashDanger, fmt.Sprintf("Failed to add owner: %v", err))
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
			h.handleAppOwnersResponse(c, id, FlashDanger, "Failed to remove owner")
		} else {
			h.handleAppOwnersResponse(c, id, FlashSuccess, "Owner removed successfully")
		}
	}
}
