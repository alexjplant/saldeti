package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func UserDeleteHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		id := c.Param("id")

		if err := h.client.DeleteUser(ctx, id); err != nil {
			SetFlash(c, FlashDanger, "Failed to delete user: "+err.Error())
		} else {
			SetFlash(c, FlashSuccess, "User deleted successfully")
		}

		c.Redirect(http.StatusSeeOther, "/google-ui/users")
	}
}
