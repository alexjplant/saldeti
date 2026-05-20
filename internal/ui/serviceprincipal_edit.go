package ui

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func SPEditHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		spID := c.Param("id")

		if c.Request.Method == http.MethodGet {
			// Get existing service principal
			sdkSP, err := h.client.ServicePrincipals().ByServicePrincipalId(spID).Get(ctx, nil)
			if err != nil {
				c.Redirect(http.StatusFound, "/ui/servicePrincipals")
				return
			}

			sp := sdkServicePrincipalToModel(sdkSP)

			h.render(c, "templates/serviceprincipals/form.html", gin.H{
				"ActiveNav":  "serviceprincipals",
				"IsEdit":     true,
				"FormAction": "/ui/servicePrincipals/" + spID + "/edit",
				"CancelURL":  "/ui/servicePrincipals/" + spID,
				"Form": map[string]interface{}{
					"ID":             sp.ID,
					"DisplayName":    sp.DisplayName,
					"Notes":          sp.Description,
					"AccountEnabled": sp.AccountEnabled != nil && *sp.AccountEnabled,
				},
			})
			return
		}

		// POST - handle form submission
		patch := models.NewServicePrincipal()

		displayName := c.PostForm("displayName")
		if displayName == "" {
			h.render(c, "templates/serviceprincipals/form.html", gin.H{
				"ActiveNav":  "serviceprincipals",
				"IsEdit":     true,
				"FormAction": "/ui/servicePrincipals/" + spID + "/edit",
				"CancelURL":  "/ui/servicePrincipals/" + spID,
				"Error":      "Display name is required",
				"Form": map[string]interface{}{
					"ID":             c.PostForm("id"),
					"DisplayName":    c.PostForm("displayName"),
					"Notes":          c.PostForm("notes"),
					"AccountEnabled": c.PostForm("accountEnabled") == "on",
				},
			})
			return
		}
		patch.SetDisplayName(&displayName)

		// Always set description (even if empty) so users can clear it
		notes := c.PostForm("notes")
		patch.SetDescription(&notes)

		accountEnabled := c.PostForm("accountEnabled") == "on"
		patch.SetAccountEnabled(&accountEnabled)

		_, err := h.client.ServicePrincipals().ByServicePrincipalId(spID).Patch(ctx, patch, nil)
		if err != nil {
			h.render(c, "templates/serviceprincipals/form.html", gin.H{
				"ActiveNav":  "serviceprincipals",
				"IsEdit":     true,
				"FormAction": "/ui/servicePrincipals/" + spID + "/edit",
				"CancelURL":  "/ui/servicePrincipals/" + spID,
				"Error":      fmt.Sprintf("Failed to update service principal: %v", err),
				"Form": map[string]interface{}{
					"ID":             c.PostForm("id"),
					"DisplayName":    c.PostForm("displayName"),
					"Notes":          c.PostForm("notes"),
					"AccountEnabled": c.PostForm("accountEnabled") == "on",
				},
			})
			return
		}

		// Success - redirect to service principal detail page
		SetFlash(c, FlashSuccess, "Service principal updated successfully")
		c.Redirect(http.StatusFound, "/ui/servicePrincipals/"+spID)
	}
}
