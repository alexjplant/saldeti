package ui

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func GroupAddMemberHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := c.PostForm("userId")

		if userID == "" {
			h.handleMembersResponse(c, id, FlashDanger, "No user selected")
			return
		}

		// Add member via SDK
		refBody := models.NewReferenceCreate()
		odataId := h.baseURL + "/v1.0/users/" + userID
		refBody.SetOdataId(&odataId)

		err := h.client.Groups().ByGroupId(id).Members().Ref().Post(c.Request.Context(), refBody, nil)
		if err != nil {
			h.handleMembersResponse(c, id, FlashDanger, fmt.Sprintf("Failed to add member: %v", err))
			return
		}

		h.handleMembersResponse(c, id, FlashSuccess, "Member added successfully")
	}
}

func GroupRemoveMemberHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		memberID := c.Param("memberId")

		if err := h.client.Groups().ByGroupId(id).Members().ByDirectoryObjectId(memberID).Ref().Delete(c.Request.Context(), nil); err != nil {
			h.handleMembersResponse(c, id, FlashDanger, "Failed to remove member")
		} else {
			h.handleMembersResponse(c, id, FlashSuccess, "Member removed successfully")
		}
	}
}
