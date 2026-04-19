package ui

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func GroupAddOwnerHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := c.PostForm("userId")
		if userID == "" {
			SetFlash(c, FlashDanger, "No user selected")
			c.Redirect(http.StatusFound, "/ui/groups/"+id)
			return
		}

		// Add owner via SDK
		refBody := models.NewReferenceCreate()
		odataId := h.baseURL + "/v1.0/users/" + userID
		refBody.SetOdataId(&odataId)

		err := h.client.Groups().ByGroupId(id).Owners().Ref().Post(c.Request.Context(), refBody, nil)
		if err != nil {
			SetFlash(c, FlashDanger, fmt.Sprintf("Failed to add owner: %v", err))
			c.Redirect(http.StatusFound, "/ui/groups/"+id)
			return
		}

		SetFlash(c, FlashSuccess, "Owner added successfully")
		c.Redirect(http.StatusFound, "/ui/groups/"+id)
	}
}

func GroupRemoveOwnerHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ownerID := c.Param("ownerId")
		if err := h.client.Groups().ByGroupId(id).Owners().ByDirectoryObjectId(ownerID).Ref().Delete(c.Request.Context(), nil); err != nil {
			SetFlash(c, FlashDanger, "Failed to remove owner")
		} else {
			SetFlash(c, FlashSuccess, "Owner removed successfully")
		}
		c.Redirect(http.StatusFound, "/ui/groups/"+id)
	}
}
