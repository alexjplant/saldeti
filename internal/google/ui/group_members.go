package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/google/model"
)

func GroupAddMemberHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		groupKey := c.Param("id")

		email := c.PostForm("email")
		role := c.PostForm("role")
		if role == "" {
			role = "MEMBER"
		}

		member := model.Member{
			Email: email,
			Role:  role,
		}

		if _, err := h.client.AddMember(ctx, groupKey, member); err != nil {
			SetFlash(c, FlashDanger, "Failed to add member: "+err.Error())
		} else {
			SetFlash(c, FlashSuccess, "Member added successfully")
		}

		c.Redirect(http.StatusSeeOther, "/google-ui/groups/"+groupKey)
	}
}

func GroupRemoveMemberHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		groupKey := c.Param("id")
		memberKey := c.Param("memberId")

		if err := h.client.RemoveMember(ctx, groupKey, memberKey); err != nil {
			SetFlash(c, FlashDanger, "Failed to remove member: "+err.Error())
		} else {
			SetFlash(c, FlashSuccess, "Member removed successfully")
		}

		c.Redirect(http.StatusSeeOther, "/google-ui/groups/"+groupKey)
	}
}