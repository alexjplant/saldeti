package ui

import (
	"github.com/gin-gonic/gin"
)

func DomainListHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		domains, err := h.client.GetDomains(ctx, "my_customer")
		if err != nil {
			h.render(c, "templates/domains/list.html", gin.H{
				"ActiveNav": "domains",
				"Error":     "Failed to load domains",
			})
			return
		}

		h.render(c, "templates/domains/list.html", gin.H{
			"ActiveNav": "domains",
			"Domains":   domains,
		})
	}
}
