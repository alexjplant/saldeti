//go:build ui

package ui_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/handler"
	"github.com/saldeti/saldeti/internal/seed"
	"github.com/saldeti/saldeti/internal/store"
	ui "github.com/saldeti/saldeti/internal/ui"
)

func setupTestServer(t *testing.T) (*httptest.Server, store.Store) {
	st := store.NewMemoryStore()
	if err := seed.Seed(st); err != nil {
		t.Fatalf("Failed to seed data: %v", err)
	}

	// Create engine with API routes
	engine := handler.NewRouter(st)

	// Create unstarted server to get a port
	ts := httptest.NewUnstartedServer(engine)
	port := ts.Listener.Addr().(*net.TCPAddr).Port

	// Register UI routes on the same engine
	ui.RegisterUIRoutes(engine, st, port)

	ts.Start()
	t.Cleanup(func() { ts.Close() })
	return ts, st
}

// Helper function to login and get session cookie
func loginAndGetSession(t *testing.T, ts *httptest.Server) *http.Cookie {
	w := httptest.NewRecorder()

	formData := url.Values{}
	formData.Set("username", "admin@saldeti.local")
	formData.Set("password", "Simulator123!")

	req, _ := http.NewRequest("POST", ts.URL+"/ui/login", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("Login failed: expected status %d, got %d", http.StatusFound, w.Code)
	}

	var sessionCookie *http.Cookie
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "saldeti_session" {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("Expected session cookie to be set after login")
	}

	return sessionCookie
}

func init() {
	// Change working directory to project root so templates can be found
	// This is needed because tests run from the internal/ui_test directory
	// but templates are in the project root's templates/ directory
}

func TestFlashHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test SetFlash
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = httptest.NewRequest("GET", "/test", nil)

	ui.SetFlash(c1, ui.FlashSuccess, "Test flash message")

	// Verify flash cookie was set
	cookies1 := w1.Result().Cookies()
	if len(cookies1) != 1 || cookies1[0].Name != "saldeti_flash" {
		t.Fatalf("Expected 1 flash cookie, got %d", len(cookies1))
	}

	// Test GetFlash - simulate a new request with the flash cookie
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest("GET", "/test", nil)
	c2.Request.AddCookie(cookies1[0])

	flash := ui.GetFlash(c2)

	if flash == nil {
		t.Fatal("Expected flash message, got nil")
	}

	if flash.Level != ui.FlashSuccess {
		t.Errorf("Expected flash level %s, got %s", ui.FlashSuccess, flash.Level)
	}

	if flash.Message != "Test flash message" {
		t.Errorf("Expected flash message 'Test flash message', got '%s'", flash.Message)
	}

	// Test that flash is cleared - GetFlash should set MaxAge=-1
	cookies2 := w2.Result().Cookies()
	foundClearedCookie := false
	for _, cookie := range cookies2 {
		if cookie.Name == "saldeti_flash" && cookie.MaxAge == -1 {
			foundClearedCookie = true
			break
		}
	}

	if !foundClearedCookie {
		// Check what cookies we actually have
		t.Logf("Response cookies after GetFlash:")
		for _, cookie := range cookies2 {
			t.Logf("  %s: MaxAge=%d", cookie.Name, cookie.MaxAge)
		}
		t.Error("Expected flash cookie to be cleared (MaxAge=-1)")
	}

	// Test that calling GetFlash again on the same request returns nil
	flash2 := ui.GetFlash(c2)
	if flash2 != nil {
		t.Error("Expected flash to be nil after first read, got another message")
	}
}

func TestSetFlashDanger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	ui.SetFlash(c, ui.FlashDanger, "Error message")

	// Transfer cookie to request for testing
	for _, cookie := range w.Result().Cookies() {
		c.Request.AddCookie(cookie)
	}

	flash := ui.GetFlash(c)

	if flash == nil {
		t.Fatal("Expected flash message, got nil")
	}

	if flash.Level != ui.FlashDanger {
		t.Errorf("Expected flash level %s, got %s", ui.FlashDanger, flash.Level)
	}
}

func TestSetFlashInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	ui.SetFlash(c, ui.FlashInfo, "Info message")

	// Transfer cookie to request for testing
	for _, cookie := range w.Result().Cookies() {
		c.Request.AddCookie(cookie)
	}

	flash := ui.GetFlash(c)

	if flash == nil {
		t.Fatal("Expected flash message, got nil")
	}

	if flash.Level != ui.FlashInfo {
		t.Errorf("Expected flash level %s, got %s", ui.FlashInfo, flash.Level)
	}
}
