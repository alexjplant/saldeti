package ui

import (
	"github.com/gin-gonic/gin"
)

func OrgUnitListHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		orgUnits, err := h.client.GetOrgUnits(ctx, "my_customer")
		if err != nil {
			h.render(c, "templates/orgunits/list.html", gin.H{
				"ActiveNav": "orgunits",
				"Error":     "Failed to load org units",
			})
			return
		}

		h.render(c, "templates/orgunits/list.html", gin.H{
			"ActiveNav": "orgunits",
			"OrgUnits":  orgUnits,
		})
	}
}