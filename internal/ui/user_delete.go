package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func UserDeleteHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := h.client.Users().ByUserId(id).Delete(c.Request.Context(), nil); err != nil {
			SetFlash(c, FlashDanger, "Failed to delete user")
		} else {
			SetFlash(c, FlashSuccess, "User deleted successfully")
		}
		c.Redirect(http.StatusFound, "/ui/users")
	}
}
