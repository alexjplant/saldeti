package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/saldeti/saldeti/internal/auth"
	"github.com/saldeti/saldeti/internal/store"
)

func zerologMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("latency", duration).
			Str("client_ip", c.ClientIP()).
			Msg("request")
	}
}

func NewRouter(st store.Store) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.RedirectTrailingSlash = false
	r.Use(zerologMiddleware())
	r.Use(gin.Recovery())

	// Token endpoint (no auth)
	r.POST("/:tenant/oauth2/v2.0/token", auth.TokenHandler(st))
	// OpenID configuration endpoint (no auth) - required by azidentity
	r.GET("/:tenant/v2.0/.well-known/openid-configuration", openIDConfigurationHandler)

	// v1.0 API group (requires auth)
	v1 := r.Group("/v1.0")
	v1.Use(auth.RequireAuth())
	// Handle /applications/(appId={appId}) format using middleware
	v1.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		if len(path) >= len("/v1.0/applications/(appId=")+1 &&
			path[:len("/v1.0/applications")] == "/v1.0/applications" &&
			path[len("/v1.0/applications"):len("/v1.0/applications/(appId=")] == "/(appId=" &&
			path[len(path)-1:] == ")" {
			// Extract appId
			appId := path[len("/v1.0/applications/(appId=") : len(path)-1]
			// Call handler directly
			c.Set("appId", appId)
			c.Abort()
			getApplicationByAppIDHandler(st)(c)
		}
		c.Next()
	})

	// Handle /servicePrincipals(appId='{appId}') format using middleware
	v1.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		spPrefix := "/v1.0/servicePrincipals"
		if len(path) >= len(spPrefix)+len("/(appId=")+1 &&
			path[:len(spPrefix)] == spPrefix &&
			path[len(spPrefix):len(spPrefix)+len("/(appId=")] == "/(appId=" &&
			path[len(path)-1:] == ")" {
			appId := path[len(spPrefix)+len("/(appId=") : len(path)-1]
			c.Set("appId", appId)
			c.Abort()
			getSPByAppIDHandler(st)(c)
		}
		c.Next()
	})
	{
		v1.POST("/$batch", batchHandler(r))

		v1.GET("/me", meHandler(st))

		// Subscribed SKUs
		v1.GET("/subscribedSkus", listSubscribedSkusHandler(st))

		// OAuth2 Permission Grants
		v1.GET("/oauth2PermissionGrants", listGrantsHandler(st))
		v1.POST("/oauth2PermissionGrants", createGrantHandler(st))
		grants := v1.Group("/oauth2PermissionGrants")
		{
			grants.GET("/:id", getGrantHandler(st))
			grants.PATCH("/:id", updateGrantHandler(st))
			grants.DELETE("/:id", deleteGrantHandler(st))
		}

		// Users
		// Add routes without trailing slash for SDK compatibility
		v1.POST("/users", createUserHandler(st))
		v1.GET("/users", listUsersHandler(st)) // List users without trailing slash
		users := v1.Group("/users")
		{
			users.GET("/", listUsersHandler(st))
			// Register delta routes BEFORE the :id group to ensure they're matched first
			users.GET("/delta", usersDeltaHandler(st))
			users.GET("/delta/", usersDeltaHandler(st))
			users.GET("/delta()", usersDeltaHandler(st)) // SDK sometimes calls with parentheses
			usersUID := users.Group("/:id")
			{
				usersUID.GET("/", getUserHandler(st))
				usersUID.GET("", getUserHandler(st)) // Handle /users/:id without trailing slash
				usersUID.PATCH("/", updateUserHandler(st))
				usersUID.PATCH("", updateUserHandler(st)) // Handle /users/:id without trailing slash
				usersUID.DELETE("/", deleteUserHandler(st))
				usersUID.DELETE("", deleteUserHandler(st)) // Handle /users/:id without trailing slash
				usersUID.GET("/memberOf", listUserMemberOfHandler(st))
				usersUID.GET("/transitiveMemberOf", listUserTransitiveMemberOfHandler(st))
				usersUID.GET("/manager", getManagerHandler(st))
				usersUID.PUT("/manager/$ref", setManagerHandler(st))
				usersUID.DELETE("/manager/$ref", removeManagerHandler(st))
				usersUID.GET("/directReports", listDirectReportsHandler(st))
				usersUID.POST("/checkMemberGroups", checkUserMemberGroupsHandler(st))
				usersUID.POST("/getMemberGroups", getUserMemberGroupsHandler(st))
			usersUID.POST("/assignLicense", assignLicenseHandler(st))
			usersUID.GET("/photo", getUserPhotoHandler(st))
			usersUID.GET("/photo/$value", getUserPhotoValueHandler(st))
			usersUID.PATCH("/photo/$value", updateUserPhotoValueHandler(st))
			usersUID.POST("/changePassword", changePasswordHandler(st))
			usersUID.POST("/reprocessLicenseAssignment", reprocessLicenseHandler(st))
			usersUID.GET("/licenseDetails", listLicenseDetailsHandler(st))
			usersUID.GET("/appRoleAssignments", listUserAppRoleAssignmentsHandler(st))
			usersUID.POST("/appRoleAssignments", createUserAppRoleAssignmentHandler(st))
			usersUID.DELETE("/appRoleAssignments/:assignmentId", deleteUserAppRoleAssignmentHandler(st))
		}
	}

		// Groups
		// Add routes without trailing slash for SDK compatibility
		v1.POST("/groups", createGroupHandler(st))
		v1.GET("/groups", listGroupsHandler(st)) // List groups without trailing slash
		groups := v1.Group("/groups")
		{
			groups.GET("/", listGroupsHandler(st))
			groups.GET("/delta", groupsDeltaHandler(st))
			groupsGID := groups.Group("/:id")
			{
				groupsGID.GET("/", getGroupHandler(st))
				groupsGID.GET("", getGroupHandler(st)) // Handle /groups/:id without trailing slash
				groupsGID.PATCH("/", updateGroupHandler(st))
				groupsGID.PATCH("", updateGroupHandler(st)) // Handle /groups/:id without trailing slash
				groupsGID.DELETE("/", deleteGroupHandler(st))
				groupsGID.DELETE("", deleteGroupHandler(st)) // Handle /groups/:id without trailing slash
				groupsGID.GET("/members", listMembersHandler(st))
				groupsGID.POST("/members/$ref", addMemberHandler(st))
				groupsGID.DELETE("/members/:memberId/$ref", removeMemberHandler(st))
				groupsGID.GET("/transitiveMembers", listTransitiveMembersHandler(st))
				groupsGID.GET("/owners", listOwnersHandler(st))
				groupsGID.POST("/owners/$ref", addOwnerHandler(st))
				groupsGID.DELETE("/owners/:ownerId/$ref", removeOwnerHandler(st))
				groupsGID.GET("/memberOf", listGroupMemberOfHandler(st))
				groupsGID.GET("/transitiveMemberOf", listGroupTransitiveMemberOfHandler(st))
				groupsGID.POST("/checkMemberGroups", checkMemberGroupsHandler(st))
			groupsGID.POST("/getMemberGroups", getMemberGroupsHandler(st))
			groupsGID.GET("/appRoleAssignments", listGroupAppRoleAssignmentsHandler(st))
			groupsGID.POST("/appRoleAssignments", createGroupAppRoleAssignmentHandler(st))
			groupsGID.DELETE("/appRoleAssignments/:assignmentId", deleteGroupAppRoleAssignmentHandler(st))
			groupsGID.POST("/getMemberObjects", getMemberObjectsHandler(st))
			// Type-cast navigation for members
				groupsGID.GET("/members/microsoft.graph.user", listMembersByTypeHandler(st, "user"))
				groupsGID.GET("/members/microsoft.graph.group", listMembersByTypeHandler(st, "group"))
				// Type-cast navigation for owners
				groupsGID.GET("/owners/microsoft.graph.user", listOwnersByTypeHandler(st, "user"))
			}
		}

		// Applications
		v1.POST("/applications", createApplicationHandler(st))
		v1.GET("/applications", listApplicationsHandler(st))
		v1.GET("/applications/", listApplicationsHandler(st))
		v1.GET("/applications/delta", applicationsDeltaHandler(st))
		v1.GET("/applications/delta/", applicationsDeltaHandler(st))
		v1.GET("/applications/:id", getApplicationHandler(st))
		v1.PATCH("/applications/:id", updateApplicationHandler(st))
		v1.DELETE("/applications/:id", deleteApplicationHandler(st))
		v1.POST("/applications/:id/addPassword", addPasswordHandler(st))
		v1.POST("/applications/:id/removePassword", removePasswordHandler(st))
		v1.POST("/applications/:id/addKey", addKeyHandler(st))
		v1.POST("/applications/:id/removeKey", removeKeyHandler(st))
		v1.GET("/applications/:id/owners", listApplicationOwnersHandler(st))
		v1.POST("/applications/:id/owners/$ref", addApplicationOwnerHandler(st))
		v1.DELETE("/applications/:id/owners/:ownerId/$ref", removeApplicationOwnerHandler(st))
		v1.GET("/applications/:id/extensionProperties", listExtensionPropertiesHandler(st))
		v1.POST("/applications/:id/extensionProperties", createExtensionPropertyHandler(st))
		v1.DELETE("/applications/:id/extensionProperties/:extId", deleteExtensionPropertyHandler(st))
		v1.POST("/applications/:id/setVerifiedPublisher", setVerifiedPublisherHandler(st))

		// Service Principals - handle /servicePrincipals(appId='{appId}') format via middleware (added above)
		v1.POST("/servicePrincipals", createServicePrincipalHandler(st))
		v1.GET("/servicePrincipals", listServicePrincipalsHandler(st))
		sps := v1.Group("/servicePrincipals")
		{
			sps.GET("/", listServicePrincipalsHandler(st))
			spsGID := sps.Group("/:id")
			{
				spsGID.GET("/", getServicePrincipalHandler(st))
				spsGID.GET("", getServicePrincipalHandler(st))
				spsGID.PATCH("/", updateServicePrincipalHandler(st))
				spsGID.PATCH("", updateServicePrincipalHandler(st))
				spsGID.DELETE("/", deleteServicePrincipalHandler(st))
				spsGID.DELETE("", deleteServicePrincipalHandler(st))
				spsGID.GET("/owners", listSPOwnersHandler(st))
				spsGID.POST("/owners/$ref", addSPOwnerHandler(st))
				spsGID.DELETE("/owners/:ownerId/$ref", removeSPOwnerHandler(st))
				spsGID.GET("/memberOf", listSPMemberOfHandler(st))
				spsGID.GET("/transitiveMemberOf", listSPTransitiveMemberOfHandler(st))
				spsGID.GET("/appRoleAssignments", listSPAppRoleAssignmentsHandler(st))
				spsGID.POST("/appRoleAssignments", createSPAppRoleAssignmentHandler(st))
				spsGID.DELETE("/appRoleAssignments/:assignmentId", deleteSPAppRoleAssignmentHandler(st))
				spsGID.GET("/appRoleAssignedTo", listSPAppRoleAssignedToHandler(st))
				spsGID.POST("/appRoleAssignedTo", createSPAppRoleAssignedToHandler(st))
				spsGID.DELETE("/appRoleAssignedTo/:assignmentId", deleteSPAppRoleAssignedToHandler(st))
				spsGID.GET("/oauth2PermissionGrants", listSPOAuth2GrantsHandler(st))
				spsGID.POST("/addPassword", spAddPasswordHandler(st))
				spsGID.POST("/removePassword", spRemovePasswordHandler(st))
				spsGID.POST("/addKey", spAddKeyHandler(st))
				spsGID.POST("/removeKey", spRemoveKeyHandler(st))
				// Policy stubs (return empty lists)
				spsGID.GET("/homeRealmDiscoveryPolicies", listEmptyPoliciesHandler(st, "homeRealmDiscoveryPolicies"))
				spsGID.GET("/claimsMappingPolicies", listEmptyPoliciesHandler(st, "claimsMappingPolicies"))
				spsGID.GET("/tokenIssuancePolicies", listEmptyPoliciesHandler(st, "tokenIssuancePolicies"))
				spsGID.GET("/tokenLifetimePolicies", listEmptyPoliciesHandler(st, "tokenLifetimePolicies"))
			}
		}

		// Directory objects
		v1.POST("/directoryObjects/getByIds", getByIdsHandler(st))
	}

	return r
}

func meHandler(store store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get claims from context
		claims, ok := c.MustGet("claims").(*auth.TokenClaims)
		if !ok {
			writeError(c, http.StatusInternalServerError, "InternalError", "Failed to get token claims")
			return
		}

		// Try to get user by subject (UPN) first
		user, err := store.GetUserByUPN(c.Request.Context(), claims.Subject)
		if err == nil {
			writeJSON(c, http.StatusOK, user)
			return
		}

		// Fall back: if token roles contain "Application", return minimal SP-shaped object
		for _, role := range claims.Roles {
			if role == "Application" {
				writeJSON(c, http.StatusOK, gin.H{
					"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#servicePrincipals/$entity",
					"@odata.type":    "#microsoft.graph.servicePrincipal",
					"id":             claims.Subject,
					"displayName":    claims.Subject,
					"appId":          claims.ClientID,
				})
				return
			}
		}

		writeError(c, http.StatusNotFound, "ResourceNotFound", "User not found")
	}
}

func openIDConfigurationHandler(c *gin.Context) {
	tenantID := c.Param("tenant")
	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	host := c.Request.Host
	baseURL := fmt.Sprintf("%s://%s/%s", scheme, host, tenantID)

	c.JSON(http.StatusOK, gin.H{
		"issuer":                                baseURL,
		"authorization_endpoint":                baseURL + "/oauth2/v2.0/authorize",
		"token_endpoint":                         baseURL + "/oauth2/v2.0/token",
		"jwks_uri":                              baseURL + "/discovery/v2.0/keys",
		"response_types_supported":               []string{"code", "id_token", "token", "token id_token"},
		"subject_types_supported":                []string{"pairwise"},
		"id_token_signing_alg_values_supported":  []string{"HS256"},
		"scopes_supported":                       []string{"openid", "profile", "email", "offline_access"},
		"token_endpoint_auth_methods_supported":  []string{"client_secret_post", "private_key_jwt", "client_secret_basic"},
		"claims_supported":                       []string{"sub", "aud", "exp", "iat", "iss", "auth_time", "acr", "amr", "email", "given_name", "family_name"},
		"request_uri_parameter_supported":        false,
		"request_parameter_supported":            false,
	})
}