package ui

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func GroupAddMemberHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := c.PostForm("userId")
		if userID == "" {
			SetFlash(c, FlashDanger, "No user selected")
			c.Redirect(http.StatusFound, "/ui/groups/"+id)
			return
		}

		// Add member via SDK
		refBody := models.NewReferenceCreate()
		odataId := h.baseURL + "/v1.0/users/" + userID
		refBody.SetOdataId(&odataId)

		err := h.client.Groups().ByGroupId(id).Members().Ref().Post(c.Request.Context(), refBody, nil)
		if err != nil {
			SetFlash(c, FlashDanger, fmt.Sprintf("Failed to add member: %v", err))
			c.Redirect(http.StatusFound, "/ui/groups/"+id)
			return
		}

		SetFlash(c, FlashSuccess, "Member added successfully")
		c.Redirect(http.StatusFound, "/ui/groups/"+id)
	}
}

func GroupRemoveMemberHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		memberID := c.Param("memberId")
		if err := h.client.Groups().ByGroupId(id).Members().ByDirectoryObjectId(memberID).Ref().Delete(c.Request.Context(), nil); err != nil {
			SetFlash(c, FlashDanger, "Failed to remove member")
		} else {
			SetFlash(c, FlashSuccess, "Member removed successfully")
		}
		c.Redirect(http.StatusFound, "/ui/groups/"+id)
	}
}
