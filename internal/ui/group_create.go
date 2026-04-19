package ui

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func GroupCreateHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			// Render empty form with defaults
			h.render(c, "templates/groups/form.html", gin.H{
				"ActiveNav":  "groups",
				"IsEdit":     false,
				"FormAction": "/ui/groups/new",
				"CancelURL":  "/ui/groups",
				"Form": map[string]interface{}{
					"DisplayName":      "",
					"Description":      "",
					"MailNickname":     "",
					"SecurityEnabled":  "true",
					"MailEnabled":      "false",
					"Unified":          "false",
					"Visibility":       "Public",
				},
			})
			return
		}

		// POST - handle form submission
		displayName := c.PostForm("displayName")

		// Validation
		if displayName == "" {
			h.render(c, "templates/groups/form.html", gin.H{
				"ActiveNav":  "groups",
				"IsEdit":     false,
				"FormAction": "/ui/groups/new",
				"CancelURL":  "/ui/groups",
				"Error":      "Display Name is required",
				"Form": map[string]interface{}{
					"DisplayName":     c.PostForm("displayName"),
					"Description":     c.PostForm("description"),
					"MailNickname":    c.PostForm("mailNickname"),
					"SecurityEnabled": c.PostForm("securityEnabled"),
					"MailEnabled":     c.PostForm("mailEnabled"),
					"Unified":         c.PostForm("unified"),
					"Visibility":      c.PostForm("visibility"),
				},
			})
			return
		}

		// Handle securityEnabled checkbox
		securityEnabled := c.PostForm("securityEnabled") == "true"

		// Handle mailEnabled checkbox
		mailEnabled := c.PostForm("mailEnabled") == "true"

		// Handle unified checkbox (GroupTypes)
		var groupTypes []string
		if c.PostForm("unified") == "true" {
			groupTypes = []string{"Unified"}
		}

		// Create group via SDK
		newGroup := models.NewGroup()
		newGroup.SetDisplayName(&displayName)
		newGroup.SetSecurityEnabled(&securityEnabled)
		newGroup.SetMailEnabled(&mailEnabled)

		if mailNickname := c.PostForm("mailNickname"); mailNickname != "" {
			newGroup.SetMailNickname(&mailNickname)
		}
		if description := c.PostForm("description"); description != "" {
			newGroup.SetDescription(&description)
		}
		if len(groupTypes) > 0 {
			newGroup.SetGroupTypes(groupTypes)
		}
		if visibility := c.PostForm("visibility"); visibility != "" {
			newGroup.SetVisibility(&visibility)
		}

		created, err := h.client.Groups().Post(c.Request.Context(), newGroup, nil)
		if err != nil {
			h.render(c, "templates/groups/form.html", gin.H{
				"ActiveNav":  "groups",
				"IsEdit":     false,
				"FormAction": "/ui/groups/new",
				"CancelURL":  "/ui/groups",
				"Error":      fmt.Sprintf("Failed to create group: %v", err),
				"Form": map[string]interface{}{
					"DisplayName":     c.PostForm("displayName"),
					"Description":     c.PostForm("description"),
					"MailNickname":    c.PostForm("mailNickname"),
					"SecurityEnabled": c.PostForm("securityEnabled"),
					"MailEnabled":     c.PostForm("mailEnabled"),
					"Unified":         c.PostForm("unified"),
					"Visibility":      c.PostForm("visibility"),
				},
			})
			return
		}

		groupID := *created.GetId()

		// Success - redirect to group detail page
		SetFlash(c, FlashSuccess, "Group created successfully")
		c.Redirect(http.StatusFound, "/ui/groups/"+groupID)
	}
}
