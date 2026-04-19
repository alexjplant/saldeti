package ui

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/auth"
)

const sessionCookieName = "saldeti_session"

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read session cookie
		tokenString, err := c.Cookie(sessionCookieName)
		if err != nil || tokenString == "" {
			c.Redirect(http.StatusFound, "/ui/login")
			c.Abort()
			return
		}

		// Validate JWT token
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.Redirect(http.StatusFound, "/ui/login")
			c.Abort()
			return
		}

		// Set user in context
		c.Set("ui_user", claims.Subject)
		c.Next()
	}
}
