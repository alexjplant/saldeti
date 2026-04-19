package ui

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func UserCreateHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			h.render(c, "templates/users/form.html", gin.H{
				"ActiveNav":  "users",
				"IsEdit":     false,
				"FormAction": "/ui/users/new",
				"CancelURL":  "/ui/users",
				"Form": map[string]interface{}{
					"DisplayName":        "",
					"UserPrincipalName":  "",
					"GivenName":          "",
					"Surname":            "",
					"Mail":               "",
					"MailNickname":       "",
					"JobTitle":           "",
					"Department":         "",
					"OfficeLocation":     "",
					"MobilePhone":        "",
					"AccountEnabled":     true,
				},
			})
			return
		}

		displayName := c.PostForm("displayName")
		upn := c.PostForm("userPrincipalName")
		if displayName == "" || upn == "" {
			h.render(c, "templates/users/form.html", gin.H{
				"ActiveNav":  "users",
				"IsEdit":     false,
				"FormAction": "/ui/users/new",
				"CancelURL":  "/ui/users",
				"Error":      "Display Name and User Principal Name are required",
				"Form": map[string]interface{}{
					"DisplayName":       c.PostForm("displayName"),
					"UserPrincipalName": c.PostForm("userPrincipalName"),
					"GivenName":         c.PostForm("givenName"),
					"Surname":           c.PostForm("surname"),
					"Mail":              c.PostForm("mail"),
					"MailNickname":      c.PostForm("mailNickname"),
					"JobTitle":          c.PostForm("jobTitle"),
					"Department":        c.PostForm("department"),
					"OfficeLocation":    c.PostForm("officeLocation"),
					"MobilePhone":       c.PostForm("mobilePhone"),
					"AccountEnabled":    c.PostForm("accountEnabled") == "true",
				},
			})
			return
		}

		accountEnabled := c.PostForm("accountEnabled") == "true"

		// Create user via SDK
		newUser := models.NewUser()
		newUser.SetDisplayName(&displayName)
		newUser.SetUserPrincipalName(&upn)
		newUser.SetAccountEnabled(&accountEnabled)

		if givenName := c.PostForm("givenName"); givenName != "" {
			newUser.SetGivenName(&givenName)
		}
		if surname := c.PostForm("surname"); surname != "" {
			newUser.SetSurname(&surname)
		}
		if mail := c.PostForm("mail"); mail != "" {
			newUser.SetMail(&mail)
		}
		if mailNickname := c.PostForm("mailNickname"); mailNickname != "" {
			newUser.SetMailNickname(&mailNickname)
		}
		if jobTitle := c.PostForm("jobTitle"); jobTitle != "" {
			newUser.SetJobTitle(&jobTitle)
		}
		if department := c.PostForm("department"); department != "" {
			newUser.SetDepartment(&department)
		}
		if officeLocation := c.PostForm("officeLocation"); officeLocation != "" {
			newUser.SetOfficeLocation(&officeLocation)
		}
		if mobilePhone := c.PostForm("mobilePhone"); mobilePhone != "" {
			newUser.SetMobilePhone(&mobilePhone)
		}

		created, err := h.client.Users().Post(c.Request.Context(), newUser, nil)
		if err != nil {
			h.render(c, "templates/users/form.html", gin.H{
				"ActiveNav":  "users",
				"IsEdit":     false,
				"FormAction": "/ui/users/new",
				"CancelURL":  "/ui/users",
				"Error":      fmt.Sprintf("Failed to create user: %v", err),
				"Form": map[string]interface{}{
					"DisplayName":       c.PostForm("displayName"),
					"UserPrincipalName": c.PostForm("userPrincipalName"),
					"GivenName":         c.PostForm("givenName"),
					"Surname":           c.PostForm("surname"),
					"Mail":              c.PostForm("mail"),
					"MailNickname":      c.PostForm("mailNickname"),
					"JobTitle":          c.PostForm("jobTitle"),
					"Department":        c.PostForm("department"),
					"OfficeLocation":    c.PostForm("officeLocation"),
					"MobilePhone":       c.PostForm("mobilePhone"),
					"AccountEnabled":    c.PostForm("accountEnabled") == "true",
				},
			})
			return
		}

		userID := *created.GetId()

		SetFlash(c, FlashSuccess, "User created successfully")
		c.Redirect(http.StatusFound, "/ui/users/"+userID)
	}
}
