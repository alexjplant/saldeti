package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/model"
)

func SPDetailHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		spID := c.Param("id")

		// Get service principal via SDK
		sdkSP, err := h.client.ServicePrincipals().ByServicePrincipalId(spID).Get(ctx, nil)
		if err != nil {
			SetFlash(c, FlashDanger, "Error loading service principal: "+err.Error())
			c.Redirect(http.StatusSeeOther, "/ui/servicePrincipals")
			return
		}
		sp := sdkServicePrincipalToModel(sdkSP)

		// Get owners via raw HTTP
		owners, err := h.fetchDirectoryObjects(ctx, h.baseURL+"/v1.0/servicePrincipals/"+spID+"/owners")
		if err != nil {
			owners = []model.DirectoryObject{}
		}

		// Get memberOf (groups) via raw HTTP
		memberOf, err := h.fetchDirectoryObjects(ctx, h.baseURL+"/v1.0/servicePrincipals/"+spID+"/memberOf")
		if err != nil {
			memberOf = []model.DirectoryObject{}
		}

		// Get appRoleAssignments via raw HTTP
		appRoleAssignments, err := h.fetchAppRoleAssignments(ctx, h.baseURL+"/v1.0/servicePrincipals/"+spID+"/appRoleAssignments")
		if err != nil {
			appRoleAssignments = []model.AppRoleAssignment{}
		}

		// Get appRoleAssignedTo via raw HTTP
		appRoleAssignedTo, err := h.fetchAppRoleAssignments(ctx, h.baseURL+"/v1.0/servicePrincipals/"+spID+"/appRoleAssignedTo")
		if err != nil {
			appRoleAssignedTo = []model.AppRoleAssignment{}
		}

		// Get oauth2PermissionGrants via raw HTTP
		oauth2Grants, err := h.fetchOAuth2PermissionGrants(ctx, h.baseURL+"/v1.0/servicePrincipals/"+spID+"/oauth2PermissionGrants")
		if err != nil {
			oauth2Grants = []model.OAuth2PermissionGrant{}
		}

		// Get all users for the "Add Owner" dropdown
		allUsers := h.fetchAllUsers(ctx)

		h.render(c, "templates/serviceprincipals/detail.html", gin.H{
			"ActiveNav":              "serviceprincipals",
			"SP":                     sp,
			"Owners":                 owners,
			"MemberOf":               memberOf,
			"AppRoleAssignments":     appRoleAssignments,
			"AppRoleAssignedTo":      appRoleAssignedTo,
			"OAuth2PermissionGrants": oauth2Grants,
			"AllUsers":               allUsers,
		})
	}
}
