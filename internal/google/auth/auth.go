package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/google/store"
)

var signingKey []byte

// SetSigningKey sets the JWT signing key for Google token generation.
func SetSigningKey(key []byte) {
	signingKey = key
}

// TokenHandler returns a gin handler for the Google OAuth2 token endpoint.
// This is a placeholder that returns 501 Not Implemented.
func TokenHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":             "not_implemented",
			"error_description": "Google OAuth2 token endpoint is not yet implemented",
		})
	}
}

// RequireAuth returns middleware that validates Google OAuth2 tokens.
// This is a placeholder that passes all requests through.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement Google OAuth2 token validation
		c.Next()
	}
}