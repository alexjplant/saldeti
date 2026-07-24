package ui

import (
	"github.com/gin-gonic/gin"
)

func RoleListHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		roles, err := h.client.GetRoles(ctx, "my_customer")
		if err != nil {
			h.render(c, "templates/roles/list.html", gin.H{
				"ActiveNav": "roles",
				"Error":     "Failed to load roles",
			})
			return
		}

		h.render(c, "templates/roles/list.html", gin.H{
			"ActiveNav": "roles",
			"Roles":     roles,
		})
	}
}
