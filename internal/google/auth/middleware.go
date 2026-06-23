package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// RequireAuth returns middleware that validates Google OAuth2 tokens.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			writeAuthError(c, 401, "Invalid token", "UNAUTHENTICATED")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			writeAuthError(c, 401, "Invalid token", "UNAUTHENTICATED")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := ValidateToken(tokenString)
		if err != nil {
			writeAuthError(c, 401, "Invalid token", "UNAUTHENTICATED")
			c.Abort()
			return
		}

		// Add claims to context
		c.Set("claims", claims)
		c.Next()
	}
}

func writeAuthError(c *gin.Context, code int, message string, status string) {
	c.JSON(code, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
			"status":  status,
		},
	})
}
