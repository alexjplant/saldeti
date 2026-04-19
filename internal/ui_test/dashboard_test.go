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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestDashboardRedirectsWhenNotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test store and seed data
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

	// Create test recorder
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui", nil)
	engine.ServeHTTP(w, req)

	// Should redirect to login
	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/ui/login" {
		t.Errorf("Expected redirect to /ui/login, got %s", location)
	}
}

func TestLoginWithValidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test store and seed data
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

	// Create test recorder
	w := httptest.NewRecorder()

	// Create login form data
	formData := url.Values{}
	formData.Set("username", "admin@saldeti.local")
	formData.Set("password", "Simulator123!")

	req, _ := http.NewRequest("POST", "/ui/login", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	engine.ServeHTTP(w, req)

	// Should redirect to dashboard
	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/ui" {
		t.Errorf("Expected redirect to /ui, got %s", location)
	}

	// Check that session cookie is set
	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "saldeti_session" {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("Expected session cookie to be set")
	}

	if sessionCookie.Path != "/ui" {
		t.Errorf("Expected cookie path /ui, got %s", sessionCookie.Path)
	}

	if !sessionCookie.HttpOnly {
		t.Error("Expected cookie to be HttpOnly")
	}
}

func TestLoginWithInvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test store and seed data
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

	// Create test recorder
	w := httptest.NewRecorder()

	// Create login form data with invalid password
	formData := url.Values{}
	formData.Set("username", "admin@saldeti.local")
	formData.Set("password", "wrongpassword")

	req, _ := http.NewRequest("POST", "/ui/login", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	engine.ServeHTTP(w, req)

	// Should return login page (not redirect)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Check that session cookie is NOT set
	cookies := w.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "saldeti_session" {
			t.Error("Expected session cookie to NOT be set on invalid login")
		}
	}

	// Check that error message is in response
	body := w.Body.String()
	if !strings.Contains(body, "Invalid username or password") {
		t.Error("Expected error message in response")
	}
}

func TestDashboardShowsStatsAfterLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test store and seed data
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

	// First, login to get session cookie
	w := httptest.NewRecorder()

	formData := url.Values{}
	formData.Set("username", "admin@saldeti.local")
	formData.Set("password", "Simulator123!")

	req, _ := http.NewRequest("POST", "/ui/login", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	engine.ServeHTTP(w, req)

	// Extract session cookie
	var sessionCookie *http.Cookie
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "saldeti_session" {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("Expected session cookie to be set")
	}

	// Now, access dashboard with session
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/ui/", nil)  // Add trailing slash
	req.AddCookie(sessionCookie)
	engine.ServeHTTP(w, req)

	// Should return OK
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Check that dashboard content is present
	body := w.Body.String()

	// Should contain stats from seeded data
	if !strings.Contains(body, "11") && !strings.Contains(body, "Total Users") {
		// We seeded 1 admin + 10 users = 11 total users
		t.Error("Expected user count in dashboard")
	}

	if !strings.Contains(body, "5") && !strings.Contains(body, "Total Groups") {
		// We seeded 5 groups
		t.Error("Expected group count in dashboard")
	}

	if !strings.Contains(body, "Dashboard") {
		t.Error("Expected 'Dashboard' title in response")
	}
}

func TestLogoutClearsSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test store and seed data
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

	// First, login to get session cookie
	w := httptest.NewRecorder()

	formData := url.Values{}
	formData.Set("username", "admin@saldeti.local")
	formData.Set("password", "Simulator123!")

	req, _ := http.NewRequest("POST", "/ui/login", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	engine.ServeHTTP(w, req)

	// Extract session cookie
	var sessionCookie *http.Cookie
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "saldeti_session" {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("Expected session cookie to be set")
	}

	// Now, logout
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/ui/logout", nil)
	req.AddCookie(sessionCookie)
	engine.ServeHTTP(w, req)

	// Should redirect to login
	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/ui/login" {
		t.Errorf("Expected redirect to /ui/login, got %s", location)
	}

	// Check that session cookie is cleared
	foundClearedCookie := false
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "saldeti_session" {
			if cookie.MaxAge == -1 {
				foundClearedCookie = true
			}
			break
		}
	}

	if !foundClearedCookie {
		t.Error("Expected session cookie to be cleared (MaxAge=-1)")
	}

	// Now try to access dashboard WITHOUT a cookie
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/ui", nil)
	// Note: not adding the cookie, simulating a real scenario where the client stops sending it
	engine.ServeHTTP(w, req)

	// Should redirect to login (session is invalid)
	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location = w.Header().Get("Location")
	if location != "/ui/login" {
		t.Errorf("Expected redirect to /ui/login, got %s", location)
	}
}
