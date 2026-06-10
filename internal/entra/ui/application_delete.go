package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ApplicationDeleteHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := h.client.Applications().ByApplicationId(id).Delete(c.Request.Context(), nil); err != nil {
			SetFlash(c, FlashDanger, "Failed to delete application")
		} else {
			SetFlash(c, FlashSuccess, "Application deleted successfully")
		}
		c.Redirect(http.StatusFound, "/ui/applications")
	}
}
