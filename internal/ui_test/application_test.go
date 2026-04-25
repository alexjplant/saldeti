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
	"github.com/saldeti/saldeti/internal/store"
)

// Ensure store package is used
var _ store.Store = nil

func findSeedApplication(t *testing.T, st store.Store) string {
	t.Helper()
	ctx := context.Background()
	apps, _, err := st.ListApplications(ctx, model.ListOptions{Top: 100})
	if err != nil {
		t.Fatalf("Failed to list applications: %v", err)
	}
	for _, app := range apps {
		if app.DisplayName == "Saldeti Simulator App" {
			return app.ID
		}
	}
	// Return first app if seed name not found
	for _, app := range apps {
		return app.ID
	}
	t.Fatal("No applications found in store")
	return ""
}

func findSeedSP(t *testing.T, st store.Store) string {
	t.Helper()
	ctx := context.Background()
	sps, _, err := st.ListServicePrincipals(ctx, model.ListOptions{Top: 100})
	if err != nil {
		t.Fatalf("Failed to list SPs: %v", err)
	}
	for _, sp := range sps {
		if strings.Contains(sp.DisplayName, "Simulator") {
			return sp.ID
		}
	}
	for _, sp := range sps {
		return sp.ID
	}
	t.Fatal("No service principals found in store")
	return ""
}

func findUserByUPN(t *testing.T, st store.Store, upn string) string {
	t.Helper()
	ctx := context.Background()
	users, _, err := st.ListUsers(ctx, model.ListOptions{Top: 100})
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}
	for _, u := range users {
		if u.UserPrincipalName == upn {
			return u.ID
		}
	}
	t.Fatalf("User %s not found", upn)
	return ""
}

func TestApplicationListShowsApplications(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, _ := setupTestServer(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/applications", nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "App Registrations") {
		t.Error("Expected 'App Registrations' heading in page")
	}
}

func TestApplicationSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, _ := setupTestServer(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/applications?search=Simulator", nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestApplicationDetail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	appID := findSeedApplication(t, st)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/applications/"+appID, nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "App Registrations") {
		t.Error("Expected breadcrumb 'App Registrations' in detail page")
	}
}

func TestApplicationCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	formData := url.Values{}
	formData.Set("displayName", "Test App via UI")
	formData.Set("description", "A test application")
	formData.Set("signInAudience", "AzureADMyOrg")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/applications/new", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusFound, w.Code, w.Body.String())
	}

	location := w.Header().Get("Location")
	if !strings.HasPrefix(location, "/ui/applications/") {
		t.Errorf("Expected redirect to /ui/applications/{id}, got %s", location)
	}

	// Follow redirect
	newAppID := strings.TrimPrefix(location, "/ui/applications/")
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", location, nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Test App via UI") {
		t.Error("Expected 'Test App via UI' in detail page")
	}

	// Verify in store
	ctx := context.Background()
	_, err := st.GetApplication(ctx, newAppID)
	if err != nil {
		t.Errorf("Expected app %s to exist in store: %v", newAppID, err)
	}
}

func TestApplicationCreateValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, _ := setupTestServer(t)

	formData := url.Values{}
	// Missing displayName

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/applications/new", strings.NewReader(formData.Encode()))
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

func TestApplicationEdit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	appID := findSeedApplication(t, st)

	formData := url.Values{}
	formData.Set("displayName", "Updated Simulator App")
	formData.Set("description", "Updated description")
	formData.Set("signInAudience", "AzureADMyOrg")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/applications/"+appID+"/edit", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusFound, w.Code, w.Body.String())
	}

	location := w.Header().Get("Location")
	if location != "/ui/applications/"+appID {
		t.Errorf("Expected redirect to /ui/applications/%s, got %s", appID, location)
	}

	// Follow redirect
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", location, nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Updated Simulator App") {
		t.Error("Expected 'Updated Simulator App' in detail page")
	}
}

func TestApplicationDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	ctx := context.Background()
	now := time.Now()
	testApp := model.Application{
		DisplayName:     "Test App to Delete",
		SignInAudience:  "AzureADMyOrg",
		CreatedDateTime: &now,
	}
	created, err := st.CreateApplication(ctx, testApp)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/applications/"+created.ID+"/delete", nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/ui/applications" {
		t.Errorf("Expected redirect to /ui/applications, got %s", location)
	}

	// Verify deletion
	_, err = st.GetApplication(ctx, created.ID)
	if err == nil {
		t.Error("Expected app to be deleted from store")
	}
}

func TestHtmxApplicationAddPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	appID := findSeedApplication(t, st)

	formData := url.Values{}
	formData.Set("displayName", "Test Password")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/applications/"+appID+"/credentials/password/add", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for htmx request, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestHtmxApplicationAddOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	appID := findSeedApplication(t, st)
	userID := findUserByUPN(t, st, "alice.smith@saldeti.local")

	formData := url.Values{}
	formData.Set("userId", userID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/applications/"+appID+"/owners/add", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for htmx request, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestHtmxApplicationRemoveOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	ctx := context.Background()
	appID := findSeedApplication(t, st)
	userID := findUserByUPN(t, st, "bob.jones@saldeti.local")

	// Add owner first
	err := st.AddApplicationOwner(ctx, appID, userID, "user")
	if err != nil {
		t.Fatalf("Failed to add owner: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/applications/"+appID+"/owners/"+userID+"/remove", nil)
	req.Header.Set("HX-Request", "true")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for htmx request, got %d", http.StatusOK, w.Code)
	}
}

func TestHtmxApplicationRemovePassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	ctx := context.Background()
	now := time.Now()
	testApp := model.Application{
		DisplayName:     "Test Remove Password App",
		SignInAudience:  "AzureADMyOrg",
		CreatedDateTime: &now,
	}
	created, err := st.CreateApplication(ctx, testApp)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}

	// Add a password credential first
	cred, err := st.AddApplicationPassword(ctx, created.ID, model.PasswordCredential{
		DisplayName: "Test Password to Remove",
	})
	if err != nil {
		t.Fatalf("Failed to add password: %v", err)
	}

	// Now remove it via UI
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/applications/"+created.ID+"/credentials/password/"+cred.KeyID+"/remove", nil)
	req.Header.Set("HX-Request", "true")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for htmx request, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestHtmxApplicationAddKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	appID := findSeedApplication(t, st)

	formData := url.Values{}
	formData.Set("keyDisplayName", "Test Key")
	formData.Set("keyType", "AsymmetricX509Cert")
	formData.Set("keyUsage", "Verify")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/applications/"+appID+"/credentials/key/add", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for htmx request, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestHtmxApplicationRemoveKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	ctx := context.Background()
	now := time.Now()
	testApp := model.Application{
		DisplayName:     "Test Remove Key App",
		SignInAudience:  "AzureADMyOrg",
		CreatedDateTime: &now,
	}
	created, err := st.CreateApplication(ctx, testApp)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}

	// Add a key credential first
	cred, err := st.AddApplicationKey(ctx, created.ID, model.KeyCredential{
		DisplayName: "Test Key to Remove",
		Type:        "AsymmetricX509Cert",
		Usage:       "Verify",
	})
	if err != nil {
		t.Fatalf("Failed to add key: %v", err)
	}

	// Now remove it via UI
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/applications/"+created.ID+"/credentials/key/"+cred.KeyID+"/remove", nil)
	req.Header.Set("HX-Request", "true")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for htmx request, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestHtmxApplicationCreateExtension(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	appID := findSeedApplication(t, st)

	formData := url.Values{}
	formData.Set("name", "testExtension")
	formData.Set("dataType", "String")
	formData.Set("targetObjects", "User")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/applications/"+appID+"/extensions/create", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for htmx request, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "Extension property created successfully") {
		t.Error("Expected 'Extension property created successfully' flash in response")
	}
}

func TestHtmxApplicationDeleteExtension(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	ctx := context.Background()
	appID := findSeedApplication(t, st)

	// Create extension property via store
	created, err := st.CreateExtensionProperty(ctx, appID, model.ExtensionProperty{
		Name:          "Test Extension to Delete",
		DataType:      "String",
		TargetObjects: []string{"User"},
	})
	if err != nil {
		t.Fatalf("Failed to create extension property: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ui/applications/"+appID+"/extensions/"+created.ID+"/delete", nil)
	req.Header.Set("HX-Request", "true")
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for htmx request, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestHtmxApplicationDetailWithExtensions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	ctx := context.Background()
	appID := findSeedApplication(t, st)

	// Create extension property via store
	_, err := st.CreateExtensionProperty(ctx, appID, model.ExtensionProperty{
		Name:          "Test Extension",
		DataType:      "String",
		TargetObjects: []string{"User"},
	})
	if err != nil {
		t.Fatalf("Failed to create extension property: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/applications/"+appID, nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Extension Properties") {
		t.Error("Expected 'Extension Properties' section in application detail page")
	}
}

func TestApplicationCreateForm(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, _ := setupTestServer(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/applications/new", nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "displayName") {
		t.Error("Expected 'displayName' form field in page")
	}
}

func TestApplicationEditForm(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts, st := setupTestServer(t)

	appID := findSeedApplication(t, st)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/applications/"+appID+"/edit", nil)
	ts.Config.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "Saldeti Simulator App") {
		t.Error("Expected 'Saldeti Simulator App' display name in edit form")
	}
}
 

