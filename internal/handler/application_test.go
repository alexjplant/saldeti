package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/saldeti/saldeti/internal/auth"
	"github.com/saldeti/saldeti/internal/model"
	"github.com/saldeti/saldeti/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateApplication(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// Create application
	appJSON := `{
		"displayName": "Test App",
		"description": "Test Description",
		"signInAudience": "AzureADandPersonalMicrosoftAccount"
	}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/applications", strings.NewReader(appJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Location"), "/v1.0/applications/")

	var appResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &appResp)
	require.NoError(t, err)

	assert.Equal(t, "https://graph.microsoft.com/v1.0/$metadata#applications/$entity", appResp["@odata.context"])
	assert.Equal(t, "Test App", appResp["displayName"])
	assert.Contains(t, appResp, "id")
	assert.NotEmpty(t, appResp["id"])
}

func TestGetApplication(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create application via store
	app := model.Application{
		DisplayName: "Test App",
		Description: "Test Desc",
	}
	createdApp, err := store.CreateApplication(ctx, app)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.Read.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// Get application by ID
	req, err := http.NewRequest("GET", server.URL+"/v1.0/applications/"+createdApp.ID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var appResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &appResp)
	require.NoError(t, err)

	assert.Equal(t, "https://graph.microsoft.com/v1.0/$metadata#applications/$entity", appResp["@odata.context"])
	assert.Equal(t, createdApp.ID, appResp["id"])
	assert.Equal(t, "Test App", appResp["displayName"])
	assert.Equal(t, "Test Desc", appResp["description"])
}

func TestGetApplicationByAppID(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create application via store
	app := model.Application{
		DisplayName: "Test App",
		Description: "Test Desc",
	}
	createdApp, err := store.CreateApplication(ctx, app)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.Read.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// Get application by AppID
	req, err := http.NewRequest("GET", server.URL+"/v1.0/applications/(appId="+createdApp.AppID+")", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var appResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &appResp)
	require.NoError(t, err)

	assert.Equal(t, "https://graph.microsoft.com/v1.0/$metadata#applications/$entity", appResp["@odata.context"])
	assert.Equal(t, "Test App", appResp["displayName"])
}

func TestGetApplicationNotFound(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.Read.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL+"/v1.0/applications/nonexistent", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestListApplications(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create 3 applications
	for i := 0; i < 3; i++ {
		app := model.Application{
			DisplayName: "Test App " + string(rune('A'+i)),
		}
		_, err := store.CreateApplication(ctx, app)
		require.NoError(t, err)
	}

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.Read.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL+"/v1.0/applications", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	assert.Equal(t, "https://graph.microsoft.com/v1.0/$metadata#applications", listResp["@odata.context"])
	value := listResp["value"].([]interface{})
	assert.Equal(t, 3, len(value))
}

func TestUpdateApplication(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create application via store
	app := model.Application{
		DisplayName: "Test App",
		Description: "Test Desc",
	}
	createdApp, err := store.CreateApplication(ctx, app)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// Update application
	patchJSON := `{"displayName": "Updated App"}`

	req, err := http.NewRequest("PATCH", server.URL+"/v1.0/applications/"+createdApp.ID, strings.NewReader(patchJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var appResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &appResp)
	require.NoError(t, err)

	assert.Equal(t, "Updated App", appResp["displayName"])
}

func TestDeleteApplication(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create application via store
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := store.CreateApplication(ctx, app)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// Delete application
	req, err := http.NewRequest("DELETE", server.URL+"/v1.0/applications/"+createdApp.ID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify it's deleted
	req, err = http.NewRequest("GET", server.URL+"/v1.0/applications/"+createdApp.ID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAddPassword(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create application via store
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := store.CreateApplication(ctx, app)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// Add password
	passwordJSON := `{"displayName": "Test Secret"}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/applications/"+createdApp.ID+"/addPassword", strings.NewReader(passwordJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var pwdResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &pwdResp)
	require.NoError(t, err)

	assert.Contains(t, pwdResp, "secretText")
	assert.Contains(t, pwdResp, "keyId")
	assert.Contains(t, pwdResp, "hint")
}

func TestRemovePassword(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create application via store
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := store.CreateApplication(ctx, app)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// First add a password to get keyId
	passwordJSON := `{"displayName": "Test Secret"}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/applications/"+createdApp.ID+"/addPassword", strings.NewReader(passwordJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var pwdResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &pwdResp)
	require.NoError(t, err)

	keyId := pwdResp["keyId"].(string)

	// Now remove the password
	removeJSON := `{"keyId": "` + keyId + `"}`

	req, err = http.NewRequest("POST", server.URL+"/v1.0/applications/"+createdApp.ID+"/removePassword", strings.NewReader(removeJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestAddKey(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create application via store
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := store.CreateApplication(ctx, app)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// Add key
	keyJSON := `{"displayName": "Test Key", "type": "AsymmetricX509Cert", "usage": "Verify"}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/applications/"+createdApp.ID+"/addKey", strings.NewReader(keyJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var keyResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &keyResp)
	require.NoError(t, err)

	assert.Contains(t, keyResp, "keyId")
}

func TestRemoveKey(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create application via store
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := store.CreateApplication(ctx, app)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// First add a key to get keyId
	keyJSON := `{"displayName": "Test Key", "type": "AsymmetricX509Cert", "usage": "Verify"}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/applications/"+createdApp.ID+"/addKey", strings.NewReader(keyJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var keyResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &keyResp)
	require.NoError(t, err)

	keyId := keyResp["keyId"].(string)

	// Now remove the key
	removeJSON := `{"keyId": "` + keyId + `"}`

	req, err = http.NewRequest("POST", server.URL+"/v1.0/applications/"+createdApp.ID+"/removeKey", strings.NewReader(removeJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestListApplicationOwners(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create user
	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "testuser@example.com",
		Mail:              "testuser@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	// Create application
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := store.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Add owner
	err = store.AddApplicationOwner(ctx, createdApp.ID, createdUser.ID, "user")
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.Read.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL+"/v1.0/applications/"+createdApp.ID+"/owners", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	value := listResp["value"].([]interface{})
	assert.Equal(t, 1, len(value))
}

func TestAddApplicationOwner(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create user
	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "testuser@example.com",
		Mail:              "testuser@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	// Create application
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := store.CreateApplication(ctx, app)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// Add owner
	refJSON := `{"@odata.id": "https://graph.microsoft.com/v1.0/directoryObjects/` + createdUser.ID + `"}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/applications/"+createdApp.ID+"/owners/$ref", strings.NewReader(refJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestRemoveApplicationOwner(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create user
	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "testuser@example.com",
		Mail:              "testuser@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	// Create application
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := store.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Add owner
	err = store.AddApplicationOwner(ctx, createdApp.ID, createdUser.ID, "user")
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// Remove owner
	req, err := http.NewRequest("DELETE", server.URL+"/v1.0/applications/"+createdApp.ID+"/owners/"+createdUser.ID+"/$ref", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestApplicationsDelta(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create an application
	app := model.Application{
		DisplayName: "Test App",
	}
	_, err := store.CreateApplication(ctx, app)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.Read.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL+"/v1.0/applications/delta", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var deltaResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &deltaResp)
	require.NoError(t, err)

	assert.Contains(t, deltaResp, "@odata.context")
	assert.Contains(t, deltaResp, "@odata.deltaLink")
	assert.Contains(t, deltaResp, "value")

	value := deltaResp["value"].([]interface{})
	assert.GreaterOrEqual(t, len(value), 1)
}

func TestListExtensionProperties(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create application
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := store.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Create extension property
	ep := model.ExtensionProperty{
		Name:          "extension_test",
		DataType:      "String",
		TargetObjects: []string{"User"},
	}
	_, err = store.CreateExtensionProperty(ctx, createdApp.ID, ep)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.Read.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL+"/v1.0/applications/"+createdApp.ID+"/extensionProperties", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	assert.Equal(t, "https://graph.microsoft.com/v1.0/$metadata#extensionProperties", listResp["@odata.context"])
	value := listResp["value"].([]interface{})
	assert.Equal(t, 1, len(value))
}

func TestCreateExtensionProperty(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create application
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := store.CreateApplication(ctx, app)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// Create extension property
	epJSON := `{"name": "extension_test", "dataType": "String", "targetObjects": ["User"]}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/applications/"+createdApp.ID+"/extensionProperties", strings.NewReader(epJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Location"), "/v1.0/applications/"+createdApp.ID+"/extensionProperties/")
}

func TestDeleteExtensionProperty(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create application
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := store.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Create extension property
	ep := model.ExtensionProperty{
		Name:          "extension_test",
		DataType:      "String",
		TargetObjects: []string{"User"},
	}
	createdEP, err := store.CreateExtensionProperty(ctx, createdApp.ID, ep)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// Delete extension property
	req, err := http.NewRequest("DELETE", server.URL+"/v1.0/applications/"+createdApp.ID+"/extensionProperties/"+createdEP.ID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestSetVerifiedPublisher(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create application
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := store.CreateApplication(ctx, app)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// Set verified publisher
	vpJSON := `{"displayName": "Test Publisher", "verifiedPublisherId": "pub123"}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/applications/"+createdApp.ID+"/setVerifiedPublisher", strings.NewReader(vpJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
