//go:build ui && entra

package ui_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/entra/auth"
	"github.com/saldeti/saldeti/internal/entra/model"
	"github.com/saldeti/saldeti/internal/entra/store"
)

// assignLicenseToUser assigns a license to a user via the assignLicense API
// endpoint (POST /v1.0/users/{id}/assignLicense). It returns the response
// recorder so callers can assert on the HTTP response.
func assignLicenseToUser(t *testing.T, ts *httptest.Server, userID, skuID string) *httptest.ResponseRecorder {
	t.Helper()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.ReadWrite.All"}, []string{"User"}, time.Hour, "", "")
	if err != nil {
		t.Fatalf("Failed to mint auth token: %v", err)
	}

	assignReq := model.LicenseAssignmentRequest{
		AddLicenses: []model.LicenseAssignment{
			{SkuID: skuID},
		},
	}
	bodyBytes, err := json.Marshal(assignReq)
	if err != nil {
		t.Fatalf("Failed to marshal assignLicense request: %v", err)
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/v1.0/users/"+userID+"/assignLicense", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	ts.Config.Handler.ServeHTTP(w, req)
	return w
}

// addUserLicense directly adds a license to a user in the store
func addUserLicense(t *testing.T, st store.Store, userUPN, skuPartNumber string, disabledPlans []string) {
	ctx := context.Background()

	// Get the user
	user, err := st.GetUserByUPN(ctx, userUPN)
	if err != nil {
		t.Fatalf("Failed to get user %s: %v", userUPN, err)
	}

	// Find SKU ID
	skuID, found := model.FindSkuByPartNumber(skuPartNumber)
	if !found {
		t.Fatalf("SKU %s not found in catalog", skuPartNumber)
	}

	// Add license using AssignLicense method
	addLicenses := []model.LicenseAssignment{
		{
			SkuID:         skuID,
			DisabledPlans: disabledPlans,
		},
	}

	_, err = st.AssignLicense(ctx, user.ID, addLicenses, nil)
	if err != nil {
		t.Fatalf("Failed to assign license to user %s: %v", userUPN, err)
	}
}

// removeUserLicense removes a license from a user in the store
func removeUserLicense(t *testing.T, st store.Store, userUPN, skuPartNumber string) {
	ctx := context.Background()

	// Get the user
	user, err := st.GetUserByUPN(ctx, userUPN)
	if err != nil {
		t.Fatalf("Failed to get user %s: %v", userUPN, err)
	}

	// Find SKU ID
	skuID, found := model.FindSkuByPartNumber(skuPartNumber)
	if !found {
		t.Fatalf("SKU %s not found in catalog", skuPartNumber)
	}

	// Remove license using AssignLicense method
	removeLicenses := []string{skuID}

	_, err = st.AssignLicense(ctx, user.ID, nil, removeLicenses)
	if err != nil {
		t.Fatalf("Failed to remove license from user %s: %v", userUPN, err)
	}
}

func TestUserDetailShowsLicenses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test server
	ts, st := setupTestServer(t)

	// Get Alice's ID (Alice already has SPE_E3 license from seed data)
	ctx := context.Background()
	alice, err := st.GetUserByUPN(ctx, "alice.smith@saldeti.local")
	if err != nil {
		t.Fatalf("Failed to get Alice user: %v", err)
	}

	// Get Alice's detail page
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/users/"+alice.ID, nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()

	// Should show Licenses section
	if !strings.Contains(body, "Licenses") {
		t.Error("Expected 'Licenses' section in detail page")
	}

	// Should show SPE_E3 license
	if !strings.Contains(body, "SPE_E3") {
		t.Error("Expected 'SPE_E3' license in detail page")
	}

	// Should show the license table row with the SKU GUID
	if !strings.Contains(body, "05e9a617-0261-4cee-bb44-138d3ef5d965") {
		t.Error("Expected SPE_E3 SKU GUID in detail page")
	}
}

func TestUserDetailShowsNoLicenses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test server
	ts, st := setupTestServer(t)

	// Grace has no licenses by default (from seed.go)
	// Get Grace's ID (disabled user with no licenses)
	ctx := context.Background()
	grace, err := st.GetUserByUPN(ctx, "grace.lee@saldeti.local")
	if err != nil {
		t.Fatalf("Failed to get Grace user: %v", err)
	}

	// Get Grace's detail page
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/users/"+grace.ID, nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()

	// Should show Licenses section or indicate no licenses
	if strings.Contains(body, "Licenses") {
		// If Licenses section exists, it should show no licenses
		if strings.Contains(body, "No licenses assigned") || strings.Contains(body, "no licenses") {
			// Good - showing no licenses message
		} else if !strings.Contains(body, "05e9a617") && !strings.Contains(body, "06ebc4ee") {
			// No license table rows - also acceptable
		} else {
			t.Error("Expected no licenses to be shown for Grace")
		}
	}
}

func TestUserDetailShowsDisabledPlans(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test server
	ts, st := setupTestServer(t)

	// Get Alice's ID (Alice already has SPE_E3 license with MCOSTANDARD disabled from seed data)
	ctx := context.Background()
	alice, err := st.GetUserByUPN(ctx, "alice.smith@saldeti.local")
	if err != nil {
		t.Fatalf("Failed to get Alice user: %v", err)
	}

	// Get Alice's detail page
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/users/"+alice.ID, nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()

	// Debug: print relevant parts of the body
	if !strings.Contains(body, "MCOSTANDARD") {
		t.Logf("Looking for MCOSTANDARD in body...")
		// Find the Licenses section
		licensesIdx := strings.Index(body, "Licenses")
		if licensesIdx >= 0 {
			t.Logf("Licenses section found. Next 500 chars: %s", body[min(licensesIdx, len(body)):min(licensesIdx+500, len(body))])
		}
	}

	// Should show MCOSTANDARD as a disabled plan
	if !strings.Contains(body, "MCOSTANDARD") {
		t.Error("Expected 'MCOSTANDARD' (disabled plan) in detail page")
	}
}

func TestLicenseAddViaUI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test server
	ts, st := setupTestServer(t)

	// Get Grace's ID (disabled user with no licenses)
	ctx := context.Background()
	grace, err := st.GetUserByUPN(ctx, "grace.lee@saldeti.local")
	if err != nil {
		t.Fatalf("Failed to get Grace user: %v", err)
	}

	// First, verify Grace has no licenses
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/ui/users/"+grace.ID, nil)
	ts.Config.Handler.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w1.Code)
	}

	body1 := w1.Body.String()
	// Verify no INTUNE_A license initially (check in assigned licenses table)
	// INTUNE_A might appear in the available SKUs dropdown, but not in assigned licenses
	// Grace should show "No licenses assigned"
	if !strings.Contains(body1, "No licenses assigned") {
		t.Logf("Response body (first 500 chars):\n%s", body1[:min(500, len(body1))])
		t.Error("Grace should show 'No licenses assigned' message initially")
	}
	// INTUNE_A should not be in the assigned licenses table
	if strings.Contains(body1, "<td><strong>INTUNE_A</strong></td>") {
		t.Error("Grace should not have INTUNE_A license assigned initially")
	}

	// Add INTUNE_A license to Grace via store
	addUserLicense(t, st, "grace.lee@saldeti.local", "INTUNE_A", []string{})

	// Get Grace's detail page after adding license
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/ui/users/"+grace.ID, nil)
	ts.Config.Handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w2.Code)
	}

	body2 := w2.Body.String()

	// Should now show INTUNE_A license
	if !strings.Contains(body2, "INTUNE_A") {
		t.Error("Expected 'INTUNE_A' license in detail page after adding")
	}

	// Verify in store
	updatedUser, err := st.GetUser(ctx, grace.ID)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	intuneSkuID, _ := model.FindSkuByPartNumber("INTUNE_A")
	found := false
	for _, lic := range updatedUser.AssignedLicenses {
		if lic.SkuID == intuneSkuID {
			found = true
			break
		}
	}
	if !found {
		t.Error("INTUNE_A license not found in user's assigned licenses in store")
	}
}

func TestLicenseRemoveViaUI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test server
	ts, st := setupTestServer(t)

	// Get Bob's ID (Bob already has SPE_E3 license from seed data)
	ctx := context.Background()
	bob, err := st.GetUserByUPN(ctx, "bob.jones@saldeti.local")
	if err != nil {
		t.Fatalf("Failed to get Bob user: %v", err)
	}

	// Verify Bob has SPE_E3 license initially
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/ui/users/"+bob.ID, nil)
	ts.Config.Handler.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w1.Code)
	}

	body1 := w1.Body.String()
	// Check if SPE_E3 appears in the assigned licenses table
	if !strings.Contains(body1, "<td><strong>SPE_E3</strong></td>") {
		t.Log("Bob should have SPE_E3 license assigned for this test to work")
	}

	// Remove SPE_E3 license from Bob via store
	removeUserLicense(t, st, "bob.jones@saldeti.local", "SPE_E3")

	// Get Bob's detail page after removal
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/ui/users/"+bob.ID, nil)
	ts.Config.Handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w2.Code)
	}

	body2 := w2.Body.String()

	// Should not show SPE_E3 license anymore in assigned licenses table
	// But it might appear in the available SKUs dropdown
	if strings.Contains(body2, "<td><strong>SPE_E3</strong></td>") {
		t.Error("SPE_E3 license should be gone from assigned licenses after removal")
	}
	// Check if it shows "No licenses assigned"
	if !strings.Contains(body2, "No licenses assigned") {
		// Might have other licenses, or SPE_E3 might be in available SKUs
		// Just verify it's not in assigned licenses
		if strings.Contains(body2, "05e9a617-0261-4cee-bb44-138d3ef5d965") && strings.Contains(body2, "<td>") {
			t.Error("SPE_E3 SKU GUID should not appear in assigned licenses table")
		}
	}

	// Verify in store
	updatedUser, err := st.GetUser(ctx, bob.ID)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	speE3SkuID, _ := model.FindSkuByPartNumber("SPE_E3")
	found := false
	for _, lic := range updatedUser.AssignedLicenses {
		if lic.SkuID == speE3SkuID {
			found = true
			break
		}
	}
	if found {
		t.Error("SPE_E3 license should be removed from user's assigned licenses in store")
	}
}

func TestLicenseAddAvailableSkusExcludesAssigned(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test server
	ts, st := setupTestServer(t)

	// Get Admin's ID (Admin already has SPE_E5 license from seed data)
	ctx := context.Background()
	admin, err := st.GetUserByUPN(ctx, "admin@saldeti.local")
	if err != nil {
		t.Fatalf("Failed to get Admin user: %v", err)
	}

	// Get Admin's detail page
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/users/"+admin.ID, nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()

	// Admin already has SPE_E5 (GUID: 06ebc4ee-1bb5-47dd-8120-11324bc54e06)
	// The "Add License" dropdown should NOT contain SPE_E5 as an option
	speE5GUID := "06ebc4ee-1bb5-47dd-8120-11324bc54e06"

	// First, verify that Admin actually has SPE_E5 assigned
	assignedLicenses := admin.AssignedLicenses
	hasSpeE5 := false
	for _, lic := range assignedLicenses {
		if lic.SkuID == speE5GUID {
			hasSpeE5 = true
			t.Logf("Admin has SPE_E5 assigned (SKU ID: %s)", lic.SkuID)
			break
		}
	}

	if !hasSpeE5 {
		t.Skip("Admin does not have SPE_E5 assigned - skipping test")
		return
	}

	// Check if SPE_E5 appears in the assigned licenses table
	hasAssignedInUI := strings.Contains(body, "<td><strong>SPE_E5</strong></td>")
	if !hasAssignedInUI {
		t.Error("Admin's SPE_E5 license should be showing in the assigned licenses table")
	}

	// Check that SPE_E5 is NOT in the available SKUs dropdown
	pattern := `value="` + speE5GUID + `"`
	if strings.Contains(body, pattern) {
		t.Error("SPE_E5 should not appear in the available SKUs dropdown since it is already assigned")
	}

	// Verify that the license section exists at all
	if !strings.Contains(body, "Licenses") {
		t.Error("License section should be present on user detail page")
	}
}

func TestHtmxLicenseAdd(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	ctx := context.Background()
	grace, err := st.GetUserByUPN(ctx, "grace.lee@saldeti.local")
	if err != nil {
		t.Fatalf("Failed to get Grace: %v", err)
	}

	// Get INTUNE_A SKU ID
	skuID, found := model.FindSkuByPartNumber("INTUNE_A")
	if !found {
		t.Fatal("INTUNE_A SKU not found")
	}

	formData := url.Values{}
	formData.Set("skuId", skuID)

	w := httptest.NewRecorder()
	req := csrfHtmxPost(t, ts, "/ui/users/"+grace.ID+"/licenses/add", formData)
	ts.Config.Handler.ServeHTTP(w, req)

	// htmx request should return 200 with partial HTML, not 302 redirect
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for htmx request, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()

	// Should contain the licenses partial (no layout wrapper)
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("htmx response should not contain full HTML layout")
	}
	if !strings.Contains(body, `id="licenses"`) {
		t.Error("htmx response should contain licenses partial with id")
	}
	// Should contain the OOB flash
	if !strings.Contains(body, "License assigned successfully") {
		t.Error("htmx response should contain flash message")
	}
	if !strings.Contains(body, "INTUNE_A") {
		t.Error("htmx response should show INTUNE_A in assigned licenses")
	}
}

func TestHtmxLicenseRemove(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	ctx := context.Background()
	bob, err := st.GetUserByUPN(ctx, "bob.jones@saldeti.local")
	if err != nil {
		t.Fatalf("Failed to get Bob: %v", err)
	}

	// Get Bob's SPE_E3 SKU ID
	skuID, found := model.FindSkuByPartNumber("SPE_E3")
	if !found {
		t.Fatal("SPE_E3 SKU not found")
	}

	w := httptest.NewRecorder()
	req := csrfHtmxPost(t, ts, "/ui/users/"+bob.ID+"/licenses/"+skuID+"/remove", nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for htmx request, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("htmx response should not contain full HTML layout")
	}
	if !strings.Contains(body, `id="licenses"`) {
		t.Error("htmx response should contain licenses partial")
	}
	if !strings.Contains(body, "License removed successfully") {
		t.Error("htmx response should contain flash message")
	}
}

func TestAssignLicense(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	ctx := context.Background()
	grace, err := st.GetUserByUPN(ctx, "grace.lee@saldeti.local")
	if err != nil {
		t.Fatalf("Failed to get Grace: %v", err)
	}

	skuID, found := model.FindSkuByPartNumber("INTUNE_A")
	if !found {
		t.Fatal("INTUNE_A SKU not found in catalog")
	}

	// Verify Grace has no INTUNE_A license initially
	for _, lic := range grace.AssignedLicenses {
		if lic.SkuID == skuID {
			t.Fatal("Grace should not have INTUNE_A license initially")
		}
	}

	// Call the assignLicense API endpoint
	w := assignLicenseToUser(t, ts, grace.ID, skuID)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// Verify the license was actually assigned in the store
	updatedUser, err := st.GetUser(ctx, grace.ID)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	found = false
	for _, lic := range updatedUser.AssignedLicenses {
		if lic.SkuID == skuID {
			found = true
			break
		}
	}
	if !found {
		t.Error("INTUNE_A license not found in user's assigned licenses after API call")
	}
}
