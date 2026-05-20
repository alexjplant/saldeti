package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/model"
)

func UserDetailHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ctx := c.Request.Context()

		sdkUser, err := h.client.Users().ByUserId(id).Get(ctx, nil)
		if err != nil {
			SetFlash(c, FlashDanger, "User not found")
			c.Redirect(http.StatusFound, "/ui/users")
			return
		}
		user := sdkUserToModel(sdkUser)

		// Manager
		manager, err := h.fetchDirectoryObject(ctx, h.baseURL+"/v1.0/users/"+id+"/manager")
		if err != nil {
			manager = nil // No manager on error
		}

		// Direct Reports
		directReports, err := h.fetchDirectoryObjects(ctx, h.baseURL+"/v1.0/users/"+id+"/directReports")
		if err != nil {
			directReports = []model.DirectoryObject{} // Empty slice on error
		}

		// Group Memberships
		groupMemberships, err := h.fetchDirectoryObjects(ctx, h.baseURL+"/v1.0/users/"+id+"/memberOf")
		if err != nil {
			groupMemberships = []model.DirectoryObject{} // Empty slice on error
		}

		// App Role Assignments
		appRoleAssignments, err := h.fetchAppRoleAssignments(ctx, h.baseURL+"/v1.0/users/"+id+"/appRoleAssignments")
		if err != nil {
			appRoleAssignments = []model.AppRoleAssignment{}
		}

		// Subscribed SKUs (for license dropdown)
		subscribedSkus, err := h.fetchSubscribedSkus(ctx)
		if err != nil {
			subscribedSkus = []model.SubscribedSku{}
		}

		// Build available SKUs (not yet assigned to user)
		assignedSkuIDs := make(map[string]bool)
		for _, lic := range user.AssignedLicenses {
			assignedSkuIDs[lic.SkuID] = true
		}
		availableSkus := make([]model.SubscribedSku, 0)
		for _, sku := range subscribedSkus {
			if !assignedSkuIDs[sku.SkuID] {
				availableSkus = append(availableSkus, sku)
			}
		}

		// All users for manager dropdown
		allUsers := h.fetchAllUsers(ctx)

		h.render(c, "templates/users/detail.html", gin.H{
			"ActiveNav":          "users",
			"User":               user,
			"Manager":            manager,
			"DirectReports":      directReports,
			"GroupMemberships":   groupMemberships,
			"AppRoleAssignments": appRoleAssignments,
			"AssignedLicenses":   user.AssignedLicenses,
			"AvailableSkus":      availableSkus,
			"AllUsers":           allUsers,
		})
	}
}
