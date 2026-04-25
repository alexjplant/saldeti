package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/model"
)

func GroupDetailHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		groupID := c.Param("id")

		// Get group
		sdkGroup, err := h.client.Groups().ByGroupId(groupID).Get(ctx, nil)
		if err != nil {
			c.Redirect(http.StatusFound, "/ui/groups")
			return
		}
		group := sdkGroupToModel(sdkGroup)

		// Get members
		members, err := h.fetchDirectoryObjects(ctx, h.baseURL+"/v1.0/groups/"+groupID+"/members")
		if err != nil {
			members = []model.DirectoryObject{}
		}

		// Get owners
		owners, err := h.fetchDirectoryObjects(ctx, h.baseURL+"/v1.0/groups/"+groupID+"/owners")
		if err != nil {
			owners = []model.DirectoryObject{}
		}

		// Get memberOf (groups this group is a member of)
		memberOf, err := h.fetchDirectoryObjects(ctx, h.baseURL+"/v1.0/groups/"+groupID+"/memberOf")
		if err != nil {
			memberOf = []model.DirectoryObject{}
		}

		// Get transitive members (all nested members)
		transitiveMembers, err := h.fetchDirectoryObjects(ctx, h.baseURL+"/v1.0/groups/"+groupID+"/transitiveMembers")
		if err != nil {
			transitiveMembers = []model.DirectoryObject{}
		}

		// Get transitive memberOf (groups this group is transitively a member of)
		transitiveMemberOf, err := h.fetchDirectoryObjects(ctx, h.baseURL+"/v1.0/groups/"+groupID+"/transitiveMemberOf")
		if err != nil {
			transitiveMemberOf = []model.DirectoryObject{}
		}

		// Get app role assignments
		groupAppRoleAssignments, err := h.fetchAppRoleAssignments(ctx, h.baseURL+"/v1.0/groups/"+groupID+"/appRoleAssignments")
		if err != nil {
			groupAppRoleAssignments = []model.AppRoleAssignment{}
		}

		// Get all users for the "Add Member" / "Add Owner" dropdowns
		allUsers := h.fetchAllUsers(ctx)

		h.render(c, "templates/groups/detail.html", gin.H{
			"ActiveNav":               "groups",
			"Group":                   group,
			"Members":                 members,
			"Owners":                  owners,
			"MemberOf":                memberOf,
			"TransitiveMembers":       transitiveMembers,
			"TransitiveMemberOf":      transitiveMemberOf,
			"GroupAppRoleAssignments": groupAppRoleAssignments,
			"AllUsers":                allUsers,
		})
	}
}
