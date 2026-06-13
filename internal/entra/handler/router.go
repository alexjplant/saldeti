package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/saldeti/saldeti/internal/entra/auth"
	"github.com/saldeti/saldeti/internal/entra/store"
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
			return
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
			return
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
		users := v1.Group("/users")
		{
			users.GET("", listUsersHandler(st))
			users.GET("/", listUsersHandler(st))
			users.POST("", createUserHandler(st))
			// Register delta routes BEFORE the :id group to ensure they're matched first
			users.GET("/delta", usersDeltaHandler(st))
			users.GET("/delta/", usersDeltaHandler(st))
			users.GET("/delta()", usersDeltaHandler(st)) // SDK sometimes calls with parentheses
			usersUID := users.Group("/:id")
			{
				usersUID.GET("/", getUserHandler(st))
				usersUID.GET("", getUserHandler(st))
				usersUID.PATCH("/", updateUserHandler(st))
				usersUID.PATCH("", updateUserHandler(st))
				usersUID.DELETE("/", deleteUserHandler(st))
				usersUID.DELETE("", deleteUserHandler(st))
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
		groups := v1.Group("/groups")
		{
			groups.GET("", listGroupsHandler(st))
			groups.GET("/", listGroupsHandler(st))
			groups.POST("", createGroupHandler(st))
			groups.GET("/delta", groupsDeltaHandler(st))
			groups.GET("/delta/", groupsDeltaHandler(st))
			groupsGID := groups.Group("/:id")
			{
				groupsGID.GET("/", getGroupHandler(st))
				groupsGID.GET("", getGroupHandler(st))
				groupsGID.PATCH("/", updateGroupHandler(st))
				groupsGID.PATCH("", updateGroupHandler(st))
				groupsGID.DELETE("/", deleteGroupHandler(st))
				groupsGID.DELETE("", deleteGroupHandler(st))
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
		apps := v1.Group("/applications")
		{
			apps.GET("", listApplicationsHandler(st))
			apps.GET("/", listApplicationsHandler(st))
			apps.POST("", createApplicationHandler(st))
			apps.GET("/delta", applicationsDeltaHandler(st))
			apps.GET("/delta/", applicationsDeltaHandler(st))
			appsUID := apps.Group("/:id")
			{
				appsUID.GET("", getApplicationHandler(st))
				appsUID.GET("/", getApplicationHandler(st))
				appsUID.PATCH("", updateApplicationHandler(st))
				appsUID.PATCH("/", updateApplicationHandler(st))
				appsUID.DELETE("", deleteApplicationHandler(st))
				appsUID.DELETE("/", deleteApplicationHandler(st))
				appsUID.POST("addPassword", addPasswordHandler(st))
				appsUID.POST("removePassword", removePasswordHandler(st))
				appsUID.POST("addKey", addKeyHandler(st))
				appsUID.POST("removeKey", removeKeyHandler(st))
				appsUID.GET("owners", listApplicationOwnersHandler(st))
				appsUID.POST("owners/$ref", addApplicationOwnerHandler(st))
				appsUID.DELETE("owners/:ownerId/$ref", removeApplicationOwnerHandler(st))
				appsUID.GET("extensionProperties", listExtensionPropertiesHandler(st))
				appsUID.POST("extensionProperties", createExtensionPropertyHandler(st))
				appsUID.DELETE("extensionProperties/:extId", deleteExtensionPropertyHandler(st))
				appsUID.POST("setVerifiedPublisher", setVerifiedPublisherHandler(st))
			}
		}

		// Service Principals - handle /servicePrincipals(appId='{appId}') format via middleware (added above)
		sps := v1.Group("/servicePrincipals")
		{
			sps.GET("", listServicePrincipalsHandler(st))
			sps.GET("/", listServicePrincipalsHandler(st))
			sps.POST("", createServicePrincipalHandler(st))
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

func meHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get claims from context
		claims, ok := c.MustGet("claims").(*auth.TokenClaims)
		if !ok {
			writeError(c, http.StatusInternalServerError, "InternalError", "Failed to get token claims")
			return
		}

		// Try to get user by subject (UPN) first
		user, err := st.GetUserByUPN(c.Request.Context(), claims.Subject)
		if err == nil {
			response, err := buildEntityResponse("https://graph.microsoft.com/v1.0/$metadata#users/$entity", user)
			if err != nil {
				writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
				return
			}

			writeJSON(c, http.StatusOK, response)
			return
		}

		// Fall back: if token roles contain "Application", look up the real SP
		for _, role := range claims.Roles {
			if role == "Application" {
				sp, err := st.GetServicePrincipalByAppID(c.Request.Context(), claims.ClientID)
				if err != nil {
					if errors.Is(err, store.ErrServicePrincipalNotFound) {
						writeError(c, http.StatusNotFound, "ResourceNotFound", "Service principal not found")
					} else {
						writeError(c, http.StatusInternalServerError, "InternalError", "Failed to get service principal")
					}
					return
				}

				response, err := buildEntityResponse("https://graph.microsoft.com/v1.0/$metadata#servicePrincipals/$entity", sp)
				if err != nil {
					writeError(c, http.StatusInternalServerError, "Service_InternalServerError", "Failed to serialize response.")
					return
				}

				writeJSON(c, http.StatusOK, response)
				return
			}
		}

		writeError(c, http.StatusNotFound, "ResourceNotFound", "User not found")
	}
}

func openIDConfigurationHandler(c *gin.Context) {
	tenantID := c.Param("tenant")
	baseURL := getBaseURL(c) + "/" + tenantID

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