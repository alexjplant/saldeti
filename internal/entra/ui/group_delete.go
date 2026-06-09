package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GroupDeleteHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := h.client.Groups().ByGroupId(id).Delete(c.Request.Context(), nil); err != nil {
			SetFlash(c, FlashDanger, "Failed to delete group")
		} else {
			SetFlash(c, FlashSuccess, "Group deleted successfully")
		}
		c.Redirect(http.StatusFound, "/ui/groups")
	}
}
