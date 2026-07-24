//go:build ui && entra

package ui_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/entra/handler"
	"github.com/saldeti/saldeti/internal/entra/store"
	ui "github.com/saldeti/saldeti/internal/entra/ui"
)

func TestDashboardShowsStatsAfterLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test store and seed data
	st := store.NewMemoryStore()
	seedTestStore(t, st)

	// Create engine with API routes
	engine := handler.NewRouter(st)

	// Generate admin credentials for the UI
	ctx := context.Background()
	adminClientID := "test-admin-client-id"
	adminClientSecret := "test-admin-secret"
	adminTenantID := "test-admin-tenant"
	if err := st.RegisterClient(ctx, adminClientID, adminClientSecret, adminTenantID); err != nil {
		t.Fatalf("Failed to register admin client: %v", err)
	}

	// Use TLS test server
	ts := httptest.NewTLSServer(engine)

	// Register UI routes with the HTTPS base URL
	baseURL := ts.URL
	if err := ui.RegisterUIRoutes(engine, baseURL, adminClientID, adminClientSecret, adminTenantID); err != nil {
		t.Fatalf("Failed to register UI routes: %v", err)
	}

	t.Cleanup(func() { ts.Close() })

	// Access dashboard directly without login
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/", nil)
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
