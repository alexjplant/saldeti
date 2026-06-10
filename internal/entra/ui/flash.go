package ui

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type FlashLevel string

const (
	FlashSuccess FlashLevel = "success"
	FlashDanger  FlashLevel = "danger"
	FlashInfo    FlashLevel = "info"
)

type Flash struct {
	Level   FlashLevel
	Message string
}

const flashCookieName = "saldeti_flash"
const flashContextKey = "flash_read"

func SetFlash(c *gin.Context, level FlashLevel, message string) {
	value := url.QueryEscape(string(level) + ":" + message)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(flashCookieName, value, 60, "/ui", "", false, true)
}

func GetFlash(c *gin.Context) *Flash {
	// Check if we've already read the flash in this request
	if _, exists := c.Get(flashContextKey); exists {
		return nil
	}

	cookie, err := c.Cookie(flashCookieName)
	if err != nil || cookie == "" {
		return nil
	}

	// Mark that we've read the flash
	c.Set(flashContextKey, true)

	// Clear the flash cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(flashCookieName, "", -1, "/ui", "", false, true)

	// Parse the flash value
	decoded, err := url.QueryUnescape(cookie)
	if err != nil {
		return nil
	}

	parts := strings.SplitN(decoded, ":", 2)
	if len(parts) != 2 {
		return nil
	}

	return &Flash{
		Level:   FlashLevel(parts[0]),
		Message: parts[1],
	}
}
