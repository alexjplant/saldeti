package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/google/model"
)

func GroupEditHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ctx := c.Request.Context()

		if c.Request.Method == http.MethodGet {
			group, err := h.client.GetGroup(ctx, id)
			if err != nil {
				SetFlash(c, FlashDanger, "Group not found")
				c.Redirect(http.StatusFound, "/google-ui/groups")
				return
			}
			h.render(c, "templates/groups/form.html", gin.H{
				"ActiveNav":  "groups",
				"IsEdit":     true,
				"FormAction": "/google-ui/groups/" + id + "/edit",
				"CancelURL":  "/google-ui/groups/" + id,
				"Form": map[string]interface{}{
					"Email":       group.Email,
					"Name":        group.Name,
					"Description": group.Description,
				},
			})
			return
		}

		// POST - update
		email := c.PostForm("email")
		name := c.PostForm("name")
		if email == "" {
			h.render(c, "templates/groups/form.html", gin.H{
				"ActiveNav":  "groups",
				"IsEdit":     true,
				"FormAction": "/google-ui/groups/" + id + "/edit",
				"CancelURL":  "/google-ui/groups/" + id,
				"Error":      "Email is required",
				"Form": map[string]interface{}{
					"Email":       c.PostForm("email"),
					"Name":        c.PostForm("name"),
					"Description": c.PostForm("description"),
				},
			})
			return
		}

		group := model.Group{
			Email:       email,
			Name:        name,
			Description: c.PostForm("description"),
		}

		_, err := h.client.UpdateGroup(ctx, id, group)
		if err != nil {
			SetFlash(c, FlashDanger, "Failed to update group")
			c.Redirect(http.StatusFound, "/google-ui/groups/"+id+"/edit")
			return
		}

		SetFlash(c, FlashSuccess, "Group updated successfully")
		c.Redirect(http.StatusFound, "/google-ui/groups/"+id)
	}
}
