package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	gauth "github.com/saldeti/saldeti/internal/google/auth"
	"github.com/saldeti/saldeti/internal/google/store"
)

// RegisterRoutes registers all Google Workspace API routes on the given gin engine.
func RegisterRoutes(engine *gin.Engine, st store.Store) {
	// OAuth2 token endpoint (no auth required)
	engine.POST("/oauth2/v1/token", gauth.TokenHandler(st))

	// Admin Directory API
	directory := engine.Group("/admin/directory/v1")
	directory.Use(gauth.RequireAuth())
	{
		directory.GET("", directoryInfoHandler())
		directory.GET("/", directoryInfoHandler())

		// Users (Tier 1B, endpoints 4-21)
		directory.GET("/users", listUsersHandler(st))
		directory.POST("/users", createUserHandler(st))
		directory.POST("/users:createGuest", createGuestUserHandler(st))
		directory.POST("/users:watch", watchUsersHandler(st))

		users := directory.Group("/users/:userKey")
		{
			users.GET("", getUserHandler(st))
			users.PUT("", updateUserHandler(st))
			users.PATCH("", patchUserHandler(st))
			users.DELETE("", deleteUserHandler(st))
			users.POST("/makeAdmin", makeAdminHandler(st))
			users.POST("/undelete", undeleteUserHandler(st))
			users.POST("/signOut", signOutUserHandler(st))
			users.POST("/aliases", addUserAliasHandler(st))
			users.GET("/aliases", listUserAliasesHandler(st))
			users.DELETE("/aliases/:alias", removeUserAliasHandler(st))
			users.GET("/photos/thumbnail", getUserPhotoHandler(st))
			users.PUT("/photos/thumbnail", updateUserPhotoHandler(st))
			users.PATCH("/photos/thumbnail", patchUserPhotoHandler(st))
			users.DELETE("/photos/thumbnail", deleteUserPhotoHandler(st))
		}

		// Groups (Tier 1C, endpoints 22-30)
		directory.GET("/groups", listGroupsHandler(st))
		directory.POST("/groups", createGroupHandler(st))

		groups := directory.Group("/groups/:groupKey")
		{
			groups.GET("", getGroupHandler(st))
			groups.PUT("", updateGroupHandler(st))
			groups.PATCH("", patchGroupHandler(st))
			groups.DELETE("", deleteGroupHandler(st))
			groups.POST("/aliases", addGroupAliasHandler(st))
			groups.GET("/aliases", listGroupAliasesHandler(st))
			groups.DELETE("/aliases/:alias", removeGroupAliasHandler(st))

			// Members (Tier 1D, endpoints 31-37)
			groups.GET("/members", listMembersHandler(st))
			groups.POST("/members", addMemberHandler(st))
			groups.GET("/hasMember/:memberKey", hasMemberHandler(st))

			members := groups.Group("/members/:memberKey")
			{
				members.GET("", getMemberHandler(st))
				members.PUT("", updateMemberHandler(st))
				members.PATCH("", patchMemberHandler(st))
				members.DELETE("", removeMemberHandler(st))
			}
		}

		// OrgUnits (Tier 2A, endpoints 38-43)
		// NOTE: Using :customer (not :customerId) to avoid Gin param name conflict
		// with the Roles group below. Both share /customer/:customer at the same tree level.
		ou := directory.Group("/customer/:customer/orgunits")
		{
			// GET with empty wildcard path dispatches to list; non-empty goes to get.
			ou.GET("/*orgUnitPath", getOrgUnitHandler(st))
			ou.POST("", createOrgUnitHandler(st))
			ou.PUT("/*orgUnitPath", updateOrgUnitHandler(st))
			ou.PATCH("/*orgUnitPath", patchOrgUnitHandler(st))
			ou.DELETE("/*orgUnitPath", deleteOrgUnitHandler(st))
		}

		// Roles, Role Assignments, Privileges (Tier 2B-2D, endpoints 44-54)
		roles := directory.Group("/customer/:customer")
		{
			roles.GET("/roles", listRolesHandler(st))
			roles.GET("/roles/:roleId", getRoleHandler(st))
			roles.POST("/roles", createRoleHandler(st))
			roles.PUT("/roles/:roleId", updateRoleHandler(st))
			roles.PATCH("/roles/:roleId", patchRoleHandler(st))
			roles.DELETE("/roles/:roleId", deleteRoleHandler(st))
			roles.GET("/roles/ALL/privileges", listPrivilegesHandler(st))

			roles.GET("/roleassignments", listRoleAssignmentsHandler(st))
			roles.GET("/roleassignments/:roleAssignmentId", getRoleAssignmentHandler(st))
			roles.POST("/roleassignments", createRoleAssignmentHandler(st))
			roles.DELETE("/roleassignments/:roleAssignmentId", deleteRoleAssignmentHandler(st))
		}
	}

	// Cloud Identity API: /v1/devices, /v1/groups, etc.
	// Routes will be registered in a future milestone.

	// Admin Reports API: /admin/reports/v1/activity, /admin/reports/v1/usage, etc.
	// Routes will be registered in a future milestone.

	// Admin Data Transfer API: /admin/datatransfer/v1/*
	// Routes will be registered in a future milestone.

	// Workspace Events API: /v1/subscriptions
	// Routes will be registered in a future milestone.
}

func directoryInfoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"kind":      "admin#directory",
			"etag":      "\"placeholder\"",
			"simulator": true,
		})
	}
}
