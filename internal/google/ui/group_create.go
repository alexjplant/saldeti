package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/google/model"
)

func GroupCreateHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			h.render(c, "templates/groups/form.html", gin.H{
				"ActiveNav": "groups",
				"Form": gin.H{
					"Email":       "",
					"Name":        "",
					"Description": "",
				},
				"FormAction": "/google-ui/groups/new",
				"CancelURL":  "/google-ui/groups",
			})
			return
		}

		// POST request
		ctx := c.Request.Context()

		email := c.PostForm("email")
		name := c.PostForm("name")
		description := c.PostForm("description")

		if email == "" {
			h.render(c, "templates/groups/form.html", gin.H{
				"ActiveNav": "groups",
				"Error":     "Email is required",
				"Form": gin.H{
					"Email":       email,
					"Name":        name,
					"Description": description,
				},
				"FormAction": "/google-ui/groups/new",
				"CancelURL":  "/google-ui/groups",
			})
			return
		}

		group := model.Group{
			Email:       email,
			Name:        name,
			Description: description,
		}

		created, err := h.client.CreateGroup(ctx, group)
		if err != nil {
			h.render(c, "templates/groups/form.html", gin.H{
				"ActiveNav": "groups",
				"Error":     "Failed to create group: " + err.Error(),
				"Form": gin.H{
					"Email":       email,
					"Name":        name,
					"Description": description,
				},
				"FormAction": "/google-ui/groups/new",
				"CancelURL":  "/google-ui/groups",
			})
			return
		}

		SetFlash(c, FlashSuccess, "Group created successfully")
		c.Redirect(http.StatusSeeOther, "/google-ui/groups/"+created.ID)
	}
}
