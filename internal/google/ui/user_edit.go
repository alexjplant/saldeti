package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/google/model"
)

func UserEditHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		id := c.Param("id")

		if c.Request.Method == http.MethodGet {
			user, err := h.client.GetUser(ctx, id)
			if err != nil {
				SetFlash(c, FlashDanger, "User not found")
				c.Redirect(http.StatusSeeOther, "/google-ui/users")
				return
			}

			suspended := user.Suspended
			orgUnitPath := user.OrgUnitPath
			if orgUnitPath == "" {
				orgUnitPath = "/"
			}

			givenName := user.GivenName
			familyName := user.FamilyName
			if user.Name != nil {
				givenName = user.Name.GivenName
				familyName = user.Name.FamilyName
			}

			h.render(c, "templates/users/form.html", gin.H{
				"ActiveNav": "users",
				"Form": gin.H{
					"PrimaryEmail": user.PrimaryEmail,
					"GivenName":    givenName,
					"FamilyName":   familyName,
					"OrgUnitPath":  orgUnitPath,
					"Suspended":    suspended,
				},
				"FormAction": "/google-ui/users/" + id + "/edit",
				"CancelURL":  "/google-ui/users/" + id,
				"IsCreate":   false,
			})
			return
		}

		// POST request
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
				"FormAction": "/google-ui/users/" + id + "/edit",
				"CancelURL":  "/google-ui/users/" + id,
				"IsCreate":   false,
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

		_, err := h.client.UpdateUser(ctx, id, user)
		if err != nil {
			h.render(c, "templates/users/form.html", gin.H{
				"ActiveNav": "users",
				"Error":     "Failed to update user: " + err.Error(),
				"Form": gin.H{
					"PrimaryEmail": primaryEmail,
					"GivenName":    givenName,
					"FamilyName":   familyName,
					"OrgUnitPath":  orgUnitPath,
					"Suspended":    suspended,
				},
				"FormAction": "/google-ui/users/" + id + "/edit",
				"CancelURL":  "/google-ui/users/" + id,
				"IsCreate":   false,
			})
			return
		}

		SetFlash(c, FlashSuccess, "User updated successfully")
		c.Redirect(http.StatusSeeOther, "/google-ui/users/"+id)
	}
}
