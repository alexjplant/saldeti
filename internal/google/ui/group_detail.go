package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/google/model"
)

func GroupDetailHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		id := c.Param("id")

		group, err := h.client.GetGroup(ctx, id)
		if err != nil {
			SetFlash(c, FlashDanger, "Group not found")
			c.Redirect(http.StatusSeeOther, "/google-ui/groups")
			return
		}

		members, err := h.client.GetMembers(ctx, id)
		if err != nil {
			members = []model.Member{}
		}

		// Get all users for the add member dropdown
		allUsers, err := h.client.GetUsers(ctx)
		if err != nil {
			allUsers = []model.User{}
		}

		h.render(c, "templates/groups/detail.html", gin.H{
			"ActiveNav": "groups",
			"Group":     group,
			"Members":   members,
			"AllUsers":  allUsers,
		})
	}
}