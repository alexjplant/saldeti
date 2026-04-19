package ui

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func GroupEditHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		groupID := c.Param("id")

		if c.Request.Method == http.MethodGet {
			// Get existing group
			sdkGroup, err := h.client.Groups().ByGroupId(groupID).Get(ctx, nil)
			if err != nil {
				c.Redirect(http.StatusFound, "/ui/groups")
				return
			}

			// Convert boolean pointers to strings for form
			securityEnabled := "false"
			if sdkGroup.GetSecurityEnabled() != nil && *sdkGroup.GetSecurityEnabled() {
				securityEnabled = "true"
			}

			mailEnabled := "false"
			if sdkGroup.GetMailEnabled() != nil && *sdkGroup.GetMailEnabled() {
				mailEnabled = "true"
			}

			unified := "false"
			for _, gt := range sdkGroup.GetGroupTypes() {
				if gt == "Unified" {
					unified = "true"
					break
				}
			}

			// Default visibility to Public if empty
			visibility := strVal(sdkGroup.GetVisibility())
			if visibility == "" {
				visibility = "Public"
			}

			// Render form pre-populated with current values
			h.render(c, "templates/groups/form.html", gin.H{
				"ActiveNav":  "groups",
				"IsEdit":     true,
				"FormAction": "/ui/groups/" + groupID + "/edit",
				"CancelURL":  "/ui/groups/" + groupID,
				"Form": map[string]interface{}{
					"DisplayName":     strVal(sdkGroup.GetDisplayName()),
					"Description":     strVal(sdkGroup.GetDescription()),
					"MailNickname":    strVal(sdkGroup.GetMailNickname()),
					"SecurityEnabled": securityEnabled,
					"MailEnabled":     mailEnabled,
					"Unified":         unified,
					"Visibility":      visibility,
				},
			})
			return
		}

		// POST - handle form submission
		// Handle securityEnabled checkbox
		securityEnabled := c.PostForm("securityEnabled") == "true"

		// Handle mailEnabled checkbox
		mailEnabled := c.PostForm("mailEnabled") == "true"

		// Handle unified checkbox (GroupTypes)
		var groupTypes []string
		if c.PostForm("unified") == "true" {
			groupTypes = []string{"Unified"}
		}

		// Update via SDK
		patch := models.NewGroup()

		if displayName := c.PostForm("displayName"); displayName != "" {
			patch.SetDisplayName(&displayName)
		}
		if description := c.PostForm("description"); description != "" {
			patch.SetDescription(&description)
		}
		if mailNickname := c.PostForm("mailNickname"); mailNickname != "" {
			patch.SetMailNickname(&mailNickname)
		}
		patch.SetSecurityEnabled(&securityEnabled)
		patch.SetMailEnabled(&mailEnabled)
		if len(groupTypes) > 0 {
			patch.SetGroupTypes(groupTypes)
		}
		if visibility := c.PostForm("visibility"); visibility != "" {
			patch.SetVisibility(&visibility)
		}

		_, err := h.client.Groups().ByGroupId(groupID).Patch(ctx, patch, nil)
		if err != nil {
			h.render(c, "templates/groups/form.html", gin.H{
				"ActiveNav":  "groups",
				"IsEdit":     true,
				"FormAction": "/ui/groups/" + groupID + "/edit",
				"CancelURL":  "/ui/groups/" + groupID,
				"Error":      fmt.Sprintf("Failed to update group: %v", err),
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

		// Success - redirect to group detail page
		SetFlash(c, FlashSuccess, "Group updated successfully")
		c.Redirect(http.StatusFound, "/ui/groups/"+groupID)
	}
}
