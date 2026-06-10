package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/google/model"
)

func UserCreateHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			h.render(c, "templates/users/form.html", gin.H{
				"ActiveNav": "users",
				"Form": gin.H{
					"PrimaryEmail": "",
					"GivenName":    "",
					"FamilyName":   "",
					"OrgUnitPath":  "/",
					"Suspended":    false,
				},
				"FormAction": "/google-ui/users/new",
				"CancelURL":  "/google-ui/users",
				"IsCreate":   true,
			})
			return
		}

		// POST request
		ctx := c.Request.Context()

		primaryEmail := c.PostForm("primaryEmail")
		givenName := c.PostForm("givenName")
		familyName := c.PostForm("familyName")
		orgUnitPath := c.PostForm("orgUnitPath")
		suspended := c.PostForm("suspended") == "true"

		if primaryEmail == "" {
			h.render(c, "templates/users/form.html", gin.H{
				"ActiveNav": "users",
				"Error":     "Primary email is required",
				"Form": gin.H{
					"PrimaryEmail": primaryEmail,
					"GivenName":    givenName,
					"FamilyName":   familyName,
					"OrgUnitPath":  orgUnitPath,
					"Suspended":    suspended,
				},
				"FormAction": "/google-ui/users/new",
				"CancelURL":  "/google-ui/users",
				"IsCreate":   true,
			})
			return
		}

		user := model.User{
			Name: &model.UserName{
				GivenName:  givenName,
				FamilyName: familyName,
			},
			PrimaryEmail: primaryEmail,
			OrgUnitPath:  orgUnitPath,
			Suspended:    suspended,
		}

		created, err := h.client.CreateUser(ctx, user)
		if err != nil {
			h.render(c, "templates/users/form.html", gin.H{
				"ActiveNav": "users",
				"Error":     "Failed to create user: " + err.Error(),
				"Form": gin.H{
					"PrimaryEmail": primaryEmail,
					"GivenName":    givenName,
					"FamilyName":   familyName,
					"OrgUnitPath":  orgUnitPath,
					"Suspended":    suspended,
				},
				"FormAction": "/google-ui/users/new",
				"CancelURL":  "/google-ui/users",
				"IsCreate":   true,
			})
			return
		}

		SetFlash(c, FlashSuccess, "User created successfully")
		c.Redirect(http.StatusSeeOther, "/google-ui/users/"+created.ID)
	}
}