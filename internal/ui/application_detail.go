package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/model"
)

func ApplicationDetailHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		appID := c.Param("id")

		// Get application
		sdkApp, err := h.client.Applications().ByApplicationId(appID).Get(ctx, nil)
		if err != nil {
			SetFlash(c, FlashDanger, "Error loading application: "+err.Error())
			c.Redirect(http.StatusSeeOther, "/ui/applications")
			return
		}
		app := sdkApplicationToModel(sdkApp)

		// Get owners via raw HTTP
		owners, err := h.fetchDirectoryObjects(ctx, h.baseURL+"/v1.0/applications/"+appID+"/owners")
		if err != nil {
			owners = []model.DirectoryObject{}
		}

		// Get extension properties via raw HTTP
		extProps, err := h.fetchExtensionProperties(ctx, h.baseURL+"/v1.0/applications/"+appID+"/extensionProperties")
		if err != nil {
			extProps = []model.ExtensionProperty{}
		}

		// Get all users for the "Add Owner" dropdown
		var allUsers []model.User
		sdkUsers, _ := h.client.Users().Get(ctx, nil)
		if sdkUsers != nil {
			for _, u := range sdkUsers.GetValue() {
				allUsers = append(allUsers, sdkUserToModel(u))
			}
		}

		h.render(c, "templates/applications/detail.html", gin.H{
			"ActiveNav":           "applications",
			"App":                 app,
			"Owners":              owners,
			"AllUsers":            allUsers,
			"ExtensionProperties": extProps,
		})
	}
}
