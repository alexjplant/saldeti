package ui

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/auth"
)

func LoginHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			h.render(c, "templates/login.html", gin.H{
				"ActiveNav": "login",
			})
			return
		}

		// POST: Handle login form submission
		if err := c.Request.ParseForm(); err != nil {
			h.render(c, "templates/login.html", gin.H{
				"ActiveNav": "login",
				"Error":     "Failed to parse form",
			})
			return
		}

		username := c.PostForm("username")
		password := c.PostForm("password")

		// Look up user by UPN
		user, err := h.store.GetUserByUPN(c.Request.Context(), username)
		if err != nil {
			h.render(c, "templates/login.html", gin.H{
				"ActiveNav": "login",
				"Error":     "Invalid username or password",
			})
			return
		}

		// Validate password using stored PasswordProfile
		if user.PasswordProfile == nil || user.PasswordProfile.Password != password {
			h.render(c, "templates/login.html", gin.H{
				"ActiveNav": "login",
				"Error":     "Invalid username or password",
			})
			return
		}

		// Mint JWT token using stored tenant ID from credentials
		tenantID := h.cred.GetTenantID()
		token, err := auth.MintToken(tenantID, "ui-session", user.UserPrincipalName, []string{"User.Read"}, []string{}, 24*time.Hour)
		if err != nil {
			h.render(c, "templates/login.html", gin.H{
				"ActiveNav": "login",
				"Error":     "Failed to create session",
			})
			return
		}

		// Set session cookie - set Secure=true only when request is HTTPS
		c.SetSameSite(http.SameSiteLaxMode)
		secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
		c.SetCookie(sessionCookieName, token, 86400, "/ui", "", secure, true)

		// Set flash message and redirect
		SetFlash(c, FlashSuccess, "Welcome back!")
		c.Redirect(http.StatusFound, "/ui")
	}
}

func LogoutHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Clear session cookie
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie("saldeti_session", "", -1, "/ui", "", false, true)

		// Set flash message and redirect
		SetFlash(c, FlashInfo, "You have been logged out")
		c.Redirect(http.StatusFound, "/ui/login")
		c.Abort() // Abort to prevent further handlers from running
	}
}
