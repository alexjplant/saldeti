package ui

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func ApplicationEditHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		appID := c.Param("id")

		if c.Request.Method == http.MethodGet {
			// Get existing application
			sdkApp, err := h.client.Applications().ByApplicationId(appID).Get(ctx, nil)
			if err != nil {
				c.Redirect(http.StatusFound, "/ui/applications")
				return
			}

			signInAudience := strVal(sdkApp.GetSignInAudience())
			if signInAudience == "" {
				signInAudience = "AzureADandPersonalMicrosoftAccount"
			}

			h.render(c, "templates/applications/form.html", gin.H{
				"ActiveNav":  "applications",
				"IsEdit":     true,
				"FormAction": "/ui/applications/" + appID + "/edit",
				"CancelURL":  "/ui/applications/" + appID,
				"Form": map[string]interface{}{
					"DisplayName":    strVal(sdkApp.GetDisplayName()),
					"Description":    strVal(sdkApp.GetDescription()),
					"SignInAudience": signInAudience,
				},
			})
			return
		}

		// POST - handle form submission
		patch := models.NewApplication()

		displayName := c.PostForm("displayName")
		if displayName == "" {
			h.render(c, "templates/applications/form.html", gin.H{
				"ActiveNav":  "applications",
				"IsEdit":     true,
				"FormAction": "/ui/applications/" + appID + "/edit",
				"CancelURL":  "/ui/applications/" + appID,
				"Error":      "Display name is required",
				"Form": map[string]interface{}{
					"DisplayName":    c.PostForm("displayName"),
					"Description":    c.PostForm("description"),
					"SignInAudience": c.PostForm("signInAudience"),
				},
			})
			return
		}
		patch.SetDisplayName(&displayName)

		// Always set description (even if empty) so users can clear it
		description := c.PostForm("description")
		patch.SetDescription(&description)

		if signInAudience := c.PostForm("signInAudience"); signInAudience != "" {
			patch.SetSignInAudience(&signInAudience)
		}

		_, err := h.client.Applications().ByApplicationId(appID).Patch(ctx, patch, nil)
		if err != nil {
			h.render(c, "templates/applications/form.html", gin.H{
				"ActiveNav":  "applications",
				"IsEdit":     true,
				"FormAction": "/ui/applications/" + appID + "/edit",
				"CancelURL":  "/ui/applications/" + appID,
				"Error":      fmt.Sprintf("Failed to update application: %v", err),
				"Form": map[string]interface{}{
					"DisplayName":    c.PostForm("displayName"),
					"Description":    c.PostForm("description"),
					"SignInAudience": c.PostForm("signInAudience"),
				},
			})
			return
		}

		// Success - redirect to application detail page
		SetFlash(c, FlashSuccess, "Application updated successfully")
		c.Redirect(http.StatusFound, "/ui/applications/"+appID)
	}
}
