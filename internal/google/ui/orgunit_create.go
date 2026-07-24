package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/google/model"
)

func OrgUnitCreateHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			h.render(c, "templates/orgunits/form.html", gin.H{
				"ActiveNav": "orgunits",
				"Form": gin.H{
					"Name":              "",
					"OrgUnitPath":       "",
					"ParentOrgUnitPath": "",
					"Description":       "",
				},
				"FormAction": "/google-ui/orgunits/new",
				"CancelURL":  "/google-ui/orgunits",
			})
			return
		}

		// POST request
		ctx := c.Request.Context()

		name := c.PostForm("name")
		orgUnitPath := c.PostForm("orgUnitPath")
		parentOrgUnitPath := c.PostForm("parentOrgUnitPath")
		description := c.PostForm("description")

		if name == "" || orgUnitPath == "" {
			h.render(c, "templates/orgunits/form.html", gin.H{
				"ActiveNav": "orgunits",
				"Error":     "Name and Org Unit Path are required",
				"Form": gin.H{
					"Name":              name,
					"OrgUnitPath":       orgUnitPath,
					"ParentOrgUnitPath": parentOrgUnitPath,
					"Description":       description,
				},
				"FormAction": "/google-ui/orgunits/new",
				"CancelURL":  "/google-ui/orgunits",
			})
			return
		}

		ou := model.OrgUnit{
			Name:              name,
			OrgUnitPath:       orgUnitPath,
			ParentOrgUnitPath: parentOrgUnitPath,
			Description:       description,
		}

		if _, err := h.client.CreateOrgUnit(ctx, "my_customer", ou); err != nil {
			h.render(c, "templates/orgunits/form.html", gin.H{
				"ActiveNav": "orgunits",
				"Error":     "Failed to create org unit: " + err.Error(),
				"Form": gin.H{
					"Name":              name,
					"OrgUnitPath":       orgUnitPath,
					"ParentOrgUnitPath": parentOrgUnitPath,
					"Description":       description,
				},
				"FormAction": "/google-ui/orgunits/new",
				"CancelURL":  "/google-ui/orgunits",
			})
			return
		}

		SetFlash(c, FlashSuccess, "Org unit created successfully")
		c.Redirect(http.StatusSeeOther, "/google-ui/orgunits")
	}
}
