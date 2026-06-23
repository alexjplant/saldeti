package ui

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

const csrfCookieName = "saldeti_csrf"
const csrfFormField = "csrf_token"
const csrfContextKey = "csrf_token"

func generateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func csrfMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isUnsafe := c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut ||
			c.Request.Method == http.MethodPatch || c.Request.Method == http.MethodDelete

		// For unsafe methods: validate CSRF token — the form field must match the cookie.
		// API requests authenticating via Authorization header are exempt.
		// htmx requests are NOT exempt: every htmx form submission includes the CSRF
		// hidden input, and SameSite=Lax prevents a cross-site attacker from reading
		// the victim's CSRF cookie to forge a matching token.
		if isUnsafe && c.GetHeader("Authorization") == "" {
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
			t, err := generateCSRFToken()
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			token = t
		}
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(csrfCookieName, token, 3600, "/ui", "", true, false)
		c.Set(csrfContextKey, token)

		c.Next()
	}
}
