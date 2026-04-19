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
			members = []model.DirectoryObject{} // Empty slice on error
		}

		// Get owners
		owners, err := h.fetchDirectoryObjects(ctx, h.baseURL+"/v1.0/groups/"+groupID+"/owners")
		if err != nil {
			owners = []model.DirectoryObject{} // Empty slice on error
		}

		// Get memberOf (groups this group is a member of)
		memberOf, err := h.fetchDirectoryObjects(ctx, h.baseURL+"/v1.0/groups/"+groupID+"/memberOf")
		if err != nil {
			memberOf = []model.DirectoryObject{} // Empty slice on error
		}

		// Get all users for the "Add Member" / "Add Owner" dropdowns
		var allUsers []model.User
		sdkUsers, _ := h.client.Users().Get(ctx, nil)
		if sdkUsers != nil {
			for _, u := range sdkUsers.GetValue() {
				allUsers = append(allUsers, sdkUserToModel(u))
			}
		}

		h.render(c, "templates/groups/detail.html", gin.H{
			"ActiveNav": "groups",
			"Group":     group,
			"Members":   members,
			"Owners":    owners,
			"MemberOf":  memberOf,
			"AllUsers":  allUsers,
		})
	}
}
