//go:build ui

package ui_test

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/auth"
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

	// Register admin client for UI first
	ctx := context.Background()
	adminClientID := "test-admin-client-id"
	adminClientSecret := "test-admin-secret"
	adminTenantID := "test-admin-tenant"
	if err := st.RegisterClient(ctx, adminClientID, adminClientSecret, adminTenantID); err != nil {
		t.Fatalf("Failed to register admin client: %v", err)
	}

	// Create engine with API routes
	engine := handler.NewRouter(st)

	// Use TLS test server
	ts := httptest.NewTLSServer(engine)

	// Register UI routes with the HTTPS base URL
	baseURL := ts.URL
	ui.RegisterUIRoutes(engine, baseURL, adminClientID, adminClientSecret, adminTenantID)

	t.Cleanup(func() { ts.Close() })
	return ts, st
}

func TestMain(m *testing.M) {
	// Configure default HTTP transport to trust self-signed cert for UI tests.
	// This is needed because azidentity creates its own HTTP client that uses
	// the default transport.
	if transport, ok := http.DefaultTransport.(*http.Transport); ok {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	auth.SetSigningKey([]byte("test-signing-key-32-bytes-long"))
	m.Run()
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
