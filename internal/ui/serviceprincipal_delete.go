package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SPDeleteHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := h.client.ServicePrincipals().ByServicePrincipalId(id).Delete(c.Request.Context(), nil); err != nil {
			SetFlash(c, FlashDanger, "Failed to delete service principal")
		} else {
			SetFlash(c, FlashSuccess, "Service principal deleted successfully")
		}
		c.Redirect(http.StatusFound, "/ui/servicePrincipals")
	}
}
