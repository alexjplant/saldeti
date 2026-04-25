package ui

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func SPAddOwnerHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := c.PostForm("userId")

		if userID == "" {
			h.handleSPOwnersResponse(c, id, FlashDanger, "No user selected")
			return
		}

		// Add owner via SDK
		refBody := models.NewReferenceCreate()
		odataId := h.baseURL + "/v1.0/users/" + userID
		refBody.SetOdataId(&odataId)

		err := h.client.ServicePrincipals().ByServicePrincipalId(id).Owners().Ref().Post(c.Request.Context(), refBody, nil)
		if err != nil {
			h.handleSPOwnersResponse(c, id, FlashDanger, fmt.Sprintf("Failed to add owner: %v", err))
			return
		}

		h.handleSPOwnersResponse(c, id, FlashSuccess, "Owner added successfully")
	}
}

func SPRemoveOwnerHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ownerID := c.Param("ownerId")

		if err := h.client.ServicePrincipals().ByServicePrincipalId(id).Owners().ByDirectoryObjectId(ownerID).Ref().Delete(c.Request.Context(), nil); err != nil {
			h.handleSPOwnersResponse(c, id, FlashDanger, "Failed to remove owner")
		} else {
			h.handleSPOwnersResponse(c, id, FlashSuccess, "Owner removed successfully")
		}
	}
}
