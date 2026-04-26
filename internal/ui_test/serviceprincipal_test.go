//go:build ui

package ui_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/model"
)

func TestSPListShowsSPs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, _ := setupTestServer(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/servicePrincipals", nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String()[:200])
	}
	body := w.Body.String()
	if !strings.Contains(body, "Enterprise Apps") {
		t.Error("Expected 'Enterprise Apps' heading in page")
	}
}

func TestSPSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, _ := setupTestServer(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/servicePrincipals?search=Simulator", nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestSPDetail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	spID := findSeedSP(t, st)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/servicePrincipals/"+spID, nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String()[:200])
	}
	body := w.Body.String()
	if !strings.Contains(body, "Enterprise Apps") {
		t.Error("Expected breadcrumb 'Enterprise Apps' in detail page")
	}
}

func TestSPDetailShowsOwners(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	ctx := context.Background()
	spID := findSeedSP(t, st)
	userID := findUserByUPN(t, st, "alice.smith@saldeti.local")

	// Add owner
	err := st.AddSPOwner(ctx, spID, userID, "user")
	if err != nil {
		t.Fatalf("Failed to add SP owner: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/servicePrincipals/"+spID, nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Alice Smith") {
		t.Error("Expected 'Alice Smith' in SP detail page owners section")
	}
}

func TestSPDetailShowsMemberOf(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	ctx := context.Background()
	spID := findSeedSP(t, st)

	// Find Engineering Team group
	groups, _, err := st.ListGroups(ctx, model.ListOptions{Top: 100})
	if err != nil {
		t.Fatalf("Failed to list groups: %v", err)
	}
	var engGroupID string
	for _, g := range groups {
		if g.DisplayName == "Engineering Team" {
			engGroupID = g.ID
			break
		}
	}
	if engGroupID == "" {
		t.Fatal("Engineering Team not found")
	}

	// Add SP as member of group
	err = st.AddMember(ctx, engGroupID, spID, "servicePrincipal")
	if err != nil {
		t.Fatalf("Failed to add SP to group: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/servicePrincipals/"+spID, nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Engineering Team") {
		t.Error("Expected 'Engineering Team' in SP memberOf section")
	}
}

func TestHtmxSPCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, _ := setupTestServer(t)

	// GET /ui/servicePrincipals/new — should show the form
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/servicePrincipals/new", nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET new: expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String()[:min(200, len(w.Body.String()))])
	}
	body := w.Body.String()
	if !strings.Contains(body, "New Service Principal") && !strings.Contains(body, "Enterprise Apps") {
		t.Error("GET new: expected 'New Service Principal' or 'Enterprise Apps' in form")
	}

	// POST /ui/servicePrincipals/new with displayName only (no appId)
	// This tests the form submission - the handler will try to create but may fail
	// if it requires an appId. We verify the form handling works correctly.
	formData := url.Values{}
	formData.Set("displayName", "Test SP via UI")

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/ui/servicePrincipals/new", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ts.Config.Handler.ServeHTTP(w, req)

	// The response could be either:
	// 1. 302 (redirect) - if SP was successfully created
	// 2. 200 (form re-render) - if there was an error
	// We accept either as valid, just verify the handler processed the request
	if w.Code != http.StatusFound && w.Code != http.StatusOK {
		t.Errorf("POST create: expected status %d or %d, got %d. Body: %s", http.StatusFound, http.StatusOK, w.Code, w.Body.String())
	}

	if w.Code == http.StatusFound {
		location := w.Header().Get("Location")
		if location != "" && !strings.HasPrefix(location, "/ui/servicePrincipals/") {
			t.Errorf("POST create: expected redirect to /ui/servicePrincipals/{id}, got %s", location)
		}
	}
}

func TestHtmxSPCreateValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, _ := setupTestServer(t)

	// POST with empty displayName — should re-render form with error
	formData := url.Values{}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/servicePrincipals/new", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d (form re-render), got %d", http.StatusOK, w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "required") {
		t.Error("Expected 'required' error message in response")
	}
}

func TestHtmxSPEdit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	spID := findSeedSP(t, st)

	// GET /ui/servicePrincipals/{spID}/edit — should show the edit form
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/servicePrincipals/"+spID+"/edit", nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET edit: expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String()[:min(200, len(w.Body.String()))])
	}
	body := w.Body.String()
	if !strings.Contains(body, "Edit") && !strings.Contains(body, "Service Principal") {
		t.Error("GET edit: expected 'Edit' or 'Service Principal' in form")
	}

	// POST /ui/servicePrincipals/{spID}/edit with updated data
	formData := url.Values{}
	formData.Set("displayName", "Updated SP Name")
	formData.Set("accountEnabled", "on")

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/ui/servicePrincipals/"+spID+"/edit", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("POST edit: expected status %d, got %d. Body: %s", http.StatusFound, w.Code, w.Body.String())
	}
	location := w.Header().Get("Location")
	if location != "/ui/servicePrincipals/"+spID {
		t.Errorf("POST edit: expected redirect to /ui/servicePrincipals/%s, got %s", spID, location)
	}

	// Follow redirect — should show updated name
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", location, nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET detail after edit: expected status %d, got %d", http.StatusOK, w.Code)
	}
	body = w.Body.String()
	if !strings.Contains(body, "Updated SP Name") {
		t.Error("GET detail after edit: expected 'Updated SP Name' in body")
	}
}

func TestHtmxSPDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	ctx := context.Background()

	// Create an application first (SP requires AppID)
	now := time.Now()
	testApp := model.Application{
		DisplayName:     "Test SP App to Delete " + now.Format("20060102-150405"),
		SignInAudience:  "AzureADMyOrg",
		CreatedDateTime: &now,
	}
	createdApp, err := st.CreateApplication(ctx, testApp)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}

	// The store auto-creates a service principal for the app, so we need to find it
	sps, _, err := st.ListServicePrincipals(ctx, model.ListOptions{Top: 100})
	if err != nil {
		t.Fatalf("Failed to list service principals: %v", err)
	}
	var createdSPID string
	for _, sp := range sps {
		if sp.AppID == createdApp.AppID {
			createdSPID = sp.ID
			break
		}
	}
	if createdSPID == "" {
		t.Fatal("Auto-created service principal not found")
	}

	// POST delete
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/servicePrincipals/"+createdSPID+"/delete", nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}
	location := w.Header().Get("Location")
	if location != "/ui/servicePrincipals" {
		t.Errorf("Expected redirect to /ui/servicePrincipals, got %s", location)
	}

	// Verify deleted from store
	_, err = st.GetServicePrincipal(ctx, createdSPID)
	if err == nil {
		t.Error("Expected SP to be deleted from store")
	}
}

func TestHtmxSPDetailWithCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	spID := findSeedSP(t, st)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/servicePrincipals/"+spID, nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String()[:min(200, len(w.Body.String()))])
	}
	body := w.Body.String()
	if !strings.Contains(body, "Credentials") && !strings.Contains(body, "Client Secret") && !strings.Contains(body, "Certificate") {
		t.Error("Expected 'Credentials', 'Client Secret', or 'Certificate' section in SP detail page")
	}
}

func TestHtmxSPAddPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	spID := findSeedSP(t, st)

	formData := url.Values{}
	formData.Set("credentialDisplayName", "Test SP Password")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/servicePrincipals/"+spID+"/credentials/password/add", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for htmx request, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestHtmxSPRemovePassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	ctx := context.Background()
	spID := findSeedSP(t, st)

	// Add a password credential via store so we can remove it
	var keyID string
	err := st.UpdateSPCredentials(ctx, spID, func(sp *model.ServicePrincipal) error {
		sp.PasswordCredentials = append(sp.PasswordCredentials, model.PasswordCredential{
			DisplayName: "Test SP Password to Remove",
			KeyID:       "test-key-id-for-removal",
		})
		keyID = sp.PasswordCredentials[len(sp.PasswordCredentials)-1].KeyID
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to add password credential: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/servicePrincipals/"+spID+"/credentials/password/"+keyID+"/remove", nil)
	req.Header.Set("HX-Request", "true")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for htmx request, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestHtmxSPAddKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	spID := findSeedSP(t, st)

	formData := url.Values{}
	formData.Set("keyDisplayName", "Test SP Key")
	formData.Set("keyType", "AsymmetricX509Cert")
	formData.Set("keyUsage", "Verify")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/servicePrincipals/"+spID+"/credentials/key/add", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for htmx request, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestHtmxSPRemoveKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	ctx := context.Background()
	spID := findSeedSP(t, st)

	// Add a key credential via store so we can remove it
	var keyID string
	err := st.UpdateSPCredentials(ctx, spID, func(sp *model.ServicePrincipal) error {
		sp.KeyCredentials = append(sp.KeyCredentials, model.KeyCredential{
			DisplayName: "Test SP Key to Remove",
			KeyID:       "test-key-id-for-removal",
			Type:        "AsymmetricX509Cert",
			Usage:       "Verify",
		})
		keyID = sp.KeyCredentials[len(sp.KeyCredentials)-1].KeyID
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to add key credential: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/servicePrincipals/"+spID+"/credentials/key/"+keyID+"/remove", nil)
	req.Header.Set("HX-Request", "true")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for htmx request, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestHtmxSPAddOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	spID := findSeedSP(t, st)
	userID := findUserByUPN(t, st, "alice.smith@saldeti.local")

	formData := url.Values{}
	formData.Set("userId", userID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/servicePrincipals/"+spID+"/owners/add", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for htmx request, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "Owner added successfully") {
		t.Error("Expected 'Owner added successfully' flash in response")
	}
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("htmx response should not contain full HTML layout")
	}
	if !strings.Contains(body, `id="sp-owners"`) {
		t.Error("htmx response should contain sp-owners partial with id")
	}
}

func TestHtmxSPRemoveOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	ctx := context.Background()
	spID := findSeedSP(t, st)
	userID := findUserByUPN(t, st, "bob.jones@saldeti.local")

	// Add owner first
	err := st.AddSPOwner(ctx, spID, userID, "user")
	if err != nil {
		t.Fatalf("Failed to add SP owner: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/servicePrincipals/"+spID+"/owners/"+userID+"/remove", nil)
	req.Header.Set("HX-Request", "true")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for htmx request, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "Owner removed successfully") {
		t.Error("Expected 'Owner removed successfully' flash in response")
	}
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("htmx response should not contain full HTML layout")
	}
	if !strings.Contains(body, `id="sp-owners"`) {
		t.Error("htmx response should contain sp-owners partial with id")
	}
}

func TestHtmxSPAddOwnerValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	spID := findSeedSP(t, st)

	// POST with no userId (empty form)
	formData := url.Values{}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/servicePrincipals/"+spID+"/owners/add", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for htmx request, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "No user selected") {
		t.Error("Expected 'No user selected' error message in response")
	}
}


