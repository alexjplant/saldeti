package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func UserDetailHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		id := c.Param("id")

		user, err := h.client.GetUser(ctx, id)
		if err != nil {
			SetFlash(c, FlashDanger, "User not found")
			c.Redirect(http.StatusSeeOther, "/google-ui/users")
			return
		}

		h.render(c, "templates/users/detail.html", gin.H{
			"ActiveNav": "users",
			"User":      user,
		})
	}
}
