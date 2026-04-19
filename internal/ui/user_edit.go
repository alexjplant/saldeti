package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func UserEditHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ctx := c.Request.Context()

		if c.Request.Method == http.MethodGet {
			sdkUser, err := h.client.Users().ByUserId(id).Get(ctx, nil)
			if err != nil {
				SetFlash(c, FlashDanger, "User not found")
				c.Redirect(http.StatusFound, "/ui/users")
				return
			}
			accountEnabled := false
			if sdkUser.GetAccountEnabled() != nil && *sdkUser.GetAccountEnabled() {
				accountEnabled = true
			}
			h.render(c, "templates/users/form.html", gin.H{
				"ActiveNav":  "users",
				"IsEdit":     true,
				"FormAction": "/ui/users/" + id + "/edit",
				"CancelURL":  "/ui/users/" + id,
				"Form": map[string]interface{}{
					"DisplayName":       strVal(sdkUser.GetDisplayName()),
					"GivenName":         strVal(sdkUser.GetGivenName()),
					"Surname":           strVal(sdkUser.GetSurname()),
					"UserPrincipalName": strVal(sdkUser.GetUserPrincipalName()),
					"Mail":              strVal(sdkUser.GetMail()),
					"MailNickname":      strVal(sdkUser.GetMailNickname()),
					"JobTitle":          strVal(sdkUser.GetJobTitle()),
					"Department":        strVal(sdkUser.GetDepartment()),
					"OfficeLocation":    strVal(sdkUser.GetOfficeLocation()),
					"MobilePhone":       strVal(sdkUser.GetMobilePhone()),
					"AccountEnabled":    accountEnabled,
				},
			})
			return
		}

		// POST - update via SDK
		patch := models.NewUser()

		if displayName := c.PostForm("displayName"); displayName != "" {
			patch.SetDisplayName(&displayName)
		}
		if givenName := c.PostForm("givenName"); givenName != "" {
			patch.SetGivenName(&givenName)
		}
		if surname := c.PostForm("surname"); surname != "" {
			patch.SetSurname(&surname)
		}
		if userPrincipalName := c.PostForm("userPrincipalName"); userPrincipalName != "" {
			patch.SetUserPrincipalName(&userPrincipalName)
		}
		if mail := c.PostForm("mail"); mail != "" {
			patch.SetMail(&mail)
		}
		if mailNickname := c.PostForm("mailNickname"); mailNickname != "" {
			patch.SetMailNickname(&mailNickname)
		}
		if jobTitle := c.PostForm("jobTitle"); jobTitle != "" {
			patch.SetJobTitle(&jobTitle)
		}
		if department := c.PostForm("department"); department != "" {
			patch.SetDepartment(&department)
		}
		if officeLocation := c.PostForm("officeLocation"); officeLocation != "" {
			patch.SetOfficeLocation(&officeLocation)
		}
		if mobilePhone := c.PostForm("mobilePhone"); mobilePhone != "" {
			patch.SetMobilePhone(&mobilePhone)
		}
		accountEnabled := c.PostForm("accountEnabled") == "true"
		patch.SetAccountEnabled(&accountEnabled)

		_, err := h.client.Users().ByUserId(id).Patch(ctx, patch, nil)
		if err != nil {
			SetFlash(c, FlashDanger, "Failed to update user: "+err.Error())
			c.Redirect(http.StatusFound, "/ui/users/"+id+"/edit")
			return
		}

		SetFlash(c, FlashSuccess, "User updated successfully")
		c.Redirect(http.StatusFound, "/ui/users/"+id)
	}
}
