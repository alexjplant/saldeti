package handler

import (
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
	r.Use(zerologMiddleware())
	r.Use(gin.Recovery())

	// Token endpoint (no auth)
	r.POST("/:tenant/oauth2/v2.0/token", auth.TokenHandler(st))

	// v1.0 API group (requires auth)
	v1 := r.Group("/v1.0")
	v1.Use(auth.RequireAuth())
	{
		v1.POST("/$batch", batchHandler(r))

		v1.GET("/me", meHandler(st))

		// Users
		users := v1.Group("/users")
		{
			users.GET("/", listUsersHandler(st))
			users.POST("/", createUserHandler(st))
			// Register delta routes BEFORE the :id group to ensure they're matched first
			users.GET("/delta", usersDeltaHandler(st))
			users.GET("/delta/", usersDeltaHandler(st))
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
			}
		}

		// Groups
		groups := v1.Group("/groups")
		{
			groups.GET("/", listGroupsHandler(st))
			groups.POST("/", createGroupHandler(st))
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
				// Type-cast navigation for members
				groupsGID.GET("/members/microsoft.graph.user", listMembersByTypeHandler(st, "user"))
				groupsGID.GET("/members/microsoft.graph.group", listMembersByTypeHandler(st, "group"))
				// Type-cast navigation for owners
				groupsGID.GET("/owners/microsoft.graph.user", listOwnersByTypeHandler(st, "user"))
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