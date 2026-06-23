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
		directory.POST("/users%3AcreateGuest", createGuestUserHandler(st))
		directory.POST("/users%3Awatch", watchUsersHandler(st))

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

			// Security (Tier 7) — under users/:userKey
			users.GET("/tokens", listTokensHandler(st))
			users.GET("/tokens/:clientId", getTokenHandler(st))
			users.DELETE("/tokens/:clientId", deleteTokenHandler(st))
			users.GET("/asps", listASPsHandler(st))
			users.GET("/asps/:codeId", getASPHandler(st))
			users.DELETE("/asps/:codeId", deleteASPHandler(st))
			users.GET("/verificationCodes", listVerificationCodesHandler(st))
			users.POST("/verificationCodes/generate", generateVerificationCodesHandler(st))
			users.POST("/verificationCodes/invalidate", invalidateVerificationCodesHandler(st))
			users.POST("/twoStepVerification/turnOff", turnOff2SVHandler(st))
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
		ou := directory.Group("/customer/:customer/orgunits")
		{
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

		// Tier 3 — Customers
		directory.GET("/customers/:customerKey", getCustomerHandler(st))
		directory.PATCH("/customers/:customerKey", patchCustomerHandler(st))
		directory.PUT("/customers/:customerKey", updateCustomerHandler(st))

		// Tier 3 — Domains
		domainsGrp := directory.Group("/customer/:customer/domains")
		{
			domainsGrp.GET("", listDomainsHandler(st))
			domainsGrp.GET("/:domainName", getDomainHandler(st))
			domainsGrp.POST("", addDomainHandler(st))
			domainsGrp.DELETE("/:domainName", deleteDomainHandler(st))
		}

		// Tier 3 — Domain Aliases
		daGrp := directory.Group("/customer/:customer/domainaliases")
		{
			daGrp.GET("", listDomainAliasesHandler(st))
			daGrp.GET("/:aliasName", getDomainAliasHandler(st))
			daGrp.POST("", createDomainAliasHandler(st))
			daGrp.DELETE("/:aliasName", deleteDomainAliasHandler(st))
		}

		// Tier 4 — ChromeOS Devices
		chromeos := directory.Group("/customer/:customer/devices/chromeos")
		{
			chromeos.GET("", listChromeOSDevicesHandler(st))
			chromeos.GET("/:deviceId", getChromeOSDeviceHandler(st))
			chromeos.PATCH("/:deviceId", patchChromeOSDeviceHandler(st))
			chromeos.PUT("/:deviceId", updateChromeOSDeviceHandler(st))
			chromeos.POST("/moveDevicesToOu", moveChromeOSDevicesHandler(st))
			chromeos.POST("%3AbatchChangeStatus", batchChangeChromeOSStatusHandler(st))
			chromeos.GET("%3AcountChromeOsDevices", countChromeOSDevicesHandler(st))
			chromeos.POST("/:deviceId%3AissueCommand", issueChromeOSCommandHandler(st))
			chromeos.GET("/:deviceId/commands/:commandId", getChromeOSCommandHandler(st))
		}

		// Tier 4 — Mobile Devices
		mobile := directory.Group("/customer/:customer/devices/mobile")
		{
			mobile.GET("", listMobileDevicesHandler(st))
			mobile.GET("/:resourceId", getMobileDeviceHandler(st))
			mobile.DELETE("/:resourceId", deleteMobileDeviceHandler(st))
			mobile.POST("/:resourceId/action", mobileDeviceActionHandler(st))
		}

		// Tier 4 — Custom Schemas
		schemasGrp := directory.Group("/customer/:customer/schemas")
		{
			schemasGrp.GET("", listSchemasHandler(st))
			schemasGrp.GET("/:schemaKey", getSchemaHandler(st))
			schemasGrp.POST("", createSchemaHandler(st))
			schemasGrp.PUT("/:schemaKey", updateSchemaHandler(st))
			schemasGrp.PATCH("/:schemaKey", patchSchemaHandler(st))
			schemasGrp.DELETE("/:schemaKey", deleteSchemaHandler(st))
		}

		// Tier 4 — Calendar Resources
		calRes := directory.Group("/customer/:customer/resources/calendars")
		{
			calRes.GET("", listCalendarResourcesHandler(st))
			calRes.GET("/:resourceId", getCalendarResourceHandler(st))
			calRes.POST("", createCalendarResourceHandler(st))
			calRes.PUT("/:resourceId", updateCalendarResourceHandler(st))
			calRes.PATCH("/:resourceId", patchCalendarResourceHandler(st))
			calRes.DELETE("/:resourceId", deleteCalendarResourceHandler(st))
		}

		// Tier 4 — Buildings
		bldg := directory.Group("/customer/:customer/resources/buildings")
		{
			bldg.GET("", listBuildingsHandler(st))
			bldg.GET("/:buildingId", getBuildingHandler(st))
			bldg.POST("", createBuildingHandler(st))
			bldg.PUT("/:buildingId", updateBuildingHandler(st))
			bldg.PATCH("/:buildingId", patchBuildingHandler(st))
			bldg.DELETE("/:buildingId", deleteBuildingHandler(st))
		}

		// Tier 4 — Features
		feat := directory.Group("/customer/:customer/resources/features")
		{
			feat.GET("", listFeaturesHandler(st))
			feat.GET("/:featureKey", getFeatureHandler(st))
			feat.POST("", createFeatureHandler(st))
			feat.PUT("/:featureKey", updateFeatureHandler(st))
			feat.PATCH("/:featureKey", patchFeatureHandler(st))
			feat.DELETE("/:featureKey", deleteFeatureHandler(st))
			feat.POST("/:featureKey/rename", renameFeatureHandler(st))
		}
	}

	// Tier 6 — Admin Reports API
	reports := engine.Group("/admin/reports/v1")
	reports.Use(gauth.RequireAuth())
	{
		reports.GET("/activity/users/:userKey/applications/:applicationName", listActivitiesHandler(st))
		reports.POST("/activity/users/:userKey/applications/:applicationName/watch", watchActivitiesHandler(st))
		reports.GET("/usage/dates/:date", getCustomerUsageReportHandler(st))
		reports.GET("/usage/users/:userKey/dates/:date", getUserUsageReportHandler(st))
		reports.GET("/usage/:entityType/:entityKey/dates/:date", getEntityUsageReportHandler(st))
	}

	// Tier 7 — Data Transfer API
	datatransfer := engine.Group("/admin/datatransfer/v1")
	datatransfer.Use(gauth.RequireAuth())
	{
		datatransfer.GET("/applications", listTransferApplicationsHandler(st))
		datatransfer.GET("/applications/:applicationId", getTransferApplicationHandler(st))
		datatransfer.GET("/transfers", listTransfersHandler(st))
		datatransfer.GET("/transfers/:dataTransferId", getTransferHandler(st))
		datatransfer.POST("/transfers", createTransferHandler(st))
	}

	// Tier 8 — Groups Settings API
	groupsSettings := engine.Group("/groups/v1/groups")
	groupsSettings.Use(gauth.RequireAuth())
	{
		groupsSettings.GET("/:groupUniqueId", getGroupSettingsHandler(st))
		groupsSettings.PUT("/:groupUniqueId", updateGroupSettingsHandler(st))
		groupsSettings.PATCH("/:groupUniqueId", patchGroupSettingsHandler(st))
	}

	// Cloud Identity API + Workspace Events (/v1/)
	v1 := engine.Group("/v1")
	v1.Use(gauth.RequireAuth())
	{
		// Devices — collection
		v1.GET("/devices", listCIDevicesHandler(st))
		v1.POST("/devices", createCIDeviceHandler(st))

		// Devices — named resource
		ciDev := v1.Group("/devices/:deviceName")
		{
			ciDev.GET("", getCIDeviceHandler(st))
			ciDev.DELETE("", deleteCIDeviceHandler(st))
			ciDev.POST("/wipe", wipeCIDeviceHandler(st))
			ciDev.POST("/cancelWipe", cancelWipeCIDeviceHandler(st))

			// Device Users
			ciDev.GET("/deviceUsers", listDeviceUsersHandler(st))
			ciDev.GET("/deviceUsers%3Alookup", lookupDeviceUserHandler(st))

			ciDevUser := ciDev.Group("/deviceUsers/:deviceUserName")
			{
				ciDevUser.GET("", getDeviceUserHandler(st))
				ciDevUser.DELETE("", deleteDeviceUserHandler(st))
				ciDevUser.POST("/approve", approveDeviceUserHandler(st))
				ciDevUser.POST("/block", blockDeviceUserHandler(st))
				ciDevUser.POST("/wipe", wipeDeviceUserHandler(st))
				ciDevUser.POST("/cancelWipe", cancelWipeDeviceUserHandler(st))
			}
		}

		// Groups — collection
		v1.GET("/groups", listCIGroupsHandler(st))
		v1.POST("/groups", createCIGroupHandler(st))
		v1.GET("/groups%3Alookup", lookupCIGroupHandler(st))
		v1.GET("/groups%3Asearch", searchCIGroupsHandler(st))

		// Groups — named resource
		ciGrp := v1.Group("/groups/:groupName")
		{
			ciGrp.GET("", getCIGroupHandler(st))
			ciGrp.PATCH("", updateCIGroupHandler(st))
			ciGrp.DELETE("", deleteCIGroupHandler(st))

			// Memberships — collection
			ciGrp.GET("/memberships", listCIMembershipsHandler(st))
			ciGrp.POST("/memberships", createCIMembershipHandler(st))
			ciGrp.GET("/memberships%3Alookup", lookupCIMembershipHandler(st))
			ciGrp.GET("/memberships%3AcheckTransitiveMembership", checkTransitiveMembershipHandler(st))
			ciGrp.GET("/memberships%3AgetMembershipGraph", getMembershipGraphHandler(st))
			ciGrp.GET("/memberships%3AsearchTransitiveGroups", searchTransitiveGroupsHandler(st))
			ciGrp.GET("/memberships%3AsearchTransitiveMemberships", searchTransitiveMembershipsHandler(st))
			ciGrp.GET("/memberships%3AsearchDirectGroups", searchDirectGroupsHandler(st))

			// Memberships — named resource
			ciMem := ciGrp.Group("/memberships/:membershipName")
			{
				ciMem.GET("", getCIMembershipHandler(st))
				ciMem.DELETE("", deleteCIMembershipHandler(st))
				ciMem.POST("/modifyMembershipRoles", modifyMembershipRolesHandler(st))
			}

			// Security Settings
			ciGrp.GET("/securitySettings", getCIGroupSecuritySettingsHandler(st))
			ciGrp.PATCH("/securitySettings", updateCIGroupSecuritySettingsHandler(st))
		}

		// Subscriptions — collection
		v1.GET("/subscriptions", listSubscriptionsHandler(st))
		v1.POST("/subscriptions", createSubscriptionHandler(st))

		// Subscriptions — named resource
		ciSub := v1.Group("/subscriptions/:subscriptionName")
		{
			ciSub.GET("", getSubscriptionHandler(st))
			ciSub.PATCH("", updateSubscriptionHandler(st))
			ciSub.DELETE("", deleteSubscriptionHandler(st))
			ciSub.POST("/reactivate", reactivateSubscriptionHandler(st))
		}

		// User Invitations
		ciInv := v1.Group("/customers/:customer/userinvitations")
		{
			ciInv.GET("", listUserInvitationsHandler(st))
			ciInv.GET("/:invitationId", getUserInvitationHandler(st))
			ciInv.GET("/:invitationId/isInvitable", isInvitableUserHandler(st))
			ciInv.POST("/:invitationId/send", sendUserInvitationHandler(st))
			ciInv.POST("/:invitationId/cancel", cancelUserInvitationHandler(st))
		}
	}
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
