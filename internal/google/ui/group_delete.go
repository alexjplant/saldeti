package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GroupDeleteHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		id := c.Param("id")

		if err := h.client.DeleteGroup(ctx, id); err != nil {
			SetFlash(c, FlashDanger, "Failed to delete group: "+err.Error())
		} else {
			SetFlash(c, FlashSuccess, "Group deleted successfully")
		}

		c.Redirect(http.StatusSeeOther, "/google-ui/groups")
	}
}
