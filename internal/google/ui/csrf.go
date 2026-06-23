package ui

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

const csrfCookieName = "saldeti_google_csrf"
const csrfFormField = "csrf_token"
const csrfContextKey = "csrf_token"

func generateCSRFToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func csrfMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isUnsafe := c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut ||
			c.Request.Method == http.MethodPatch || c.Request.Method == http.MethodDelete

		// For unsafe methods: validate against the EXISTING cookie before generating a new token.
		// htmx requests automatically include the cookie, so they can skip CSRF validation.
		if isUnsafe && c.GetHeader("HX-Request") != "true" && c.GetHeader("Authorization") == "" {
			formToken := c.PostForm(csrfFormField)
			cookieToken, err := c.Cookie(csrfCookieName)
			if err != nil || formToken == "" || formToken != cookieToken {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		// Reuse existing cookie value if available, otherwise generate a fresh token.
		// Only generating on first visit prevents static-file responses (JS, CSS) from
		// overwriting the cookie between page render and form submit.
		token, _ := c.Cookie(csrfCookieName)
		if token == "" {
			token = generateCSRFToken()
		}
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(csrfCookieName, token, 3600, "/google-ui", "", false, false)
		c.Set(csrfContextKey, token)

		c.Next()
	}
}
