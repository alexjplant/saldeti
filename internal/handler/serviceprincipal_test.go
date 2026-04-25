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

func findSPByAppID(st store.Store, appID string) (*model.ServicePrincipal, error) {
	return st.GetServicePrincipalByAppID(context.Background(), appID)
}

func TestServicePrincipalList(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create 3 apps (which auto-creates SPs)
	for i := 0; i < 3; i++ {
		app := model.Application{
			DisplayName: "Test App " + string(rune('A'+i)),
		}
		_, err := st.CreateApplication(ctx, app)
		require.NoError(t, err)
	}

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.Read.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL+"/v1.0/servicePrincipals", nil)
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

	assert.Equal(t, "https://graph.microsoft.com/v1.0/$metadata#servicePrincipals", listResp["@odata.context"])
	value := listResp["value"].([]interface{})
	assert.GreaterOrEqual(t, len(value), 3)
}

func TestServicePrincipalListWithFilter(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create an app with specific displayName
	app := model.Application{
		DisplayName: "Test App 1",
	}
	createdApp, err := st.CreateApplication(ctx, app)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.Read.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// List all service principals without filter, then verify in code
	req, err := http.NewRequest("GET", server.URL+"/v1.0/servicePrincipals", nil)
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

	// Verify the SP for our app exists
	value := listResp["value"].([]interface{})
	found := false
	for _, v := range value {
		sp := v.(map[string]interface{})
		if sp["appId"] == createdApp.AppID {
			found = true
			break
		}
	}
	assert.True(t, found, "Service principal for created app should exist")
}

func TestServicePrincipalListWithTop(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create 5 apps
	for i := 0; i < 5; i++ {
		app := model.Application{
			DisplayName: "Test App " + string(rune('A'+i)),
		}
		_, err := st.CreateApplication(ctx, app)
		require.NoError(t, err)
	}

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.Read.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL+"/v1.0/servicePrincipals?$top=2", nil)
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
	assert.Equal(t, 2, len(value))
}

func TestServicePrincipalGet(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create an app (auto-creates SP)
	app := model.Application{
		DisplayName: "Test App",
		Description: "Test Desc",
	}
	createdApp, err := st.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Find the SP
	sp, err := findSPByAppID(st, createdApp.AppID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.Read.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL+"/v1.0/servicePrincipals/"+sp.ID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var spResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &spResp)
	require.NoError(t, err)

	assert.Equal(t, "https://graph.microsoft.com/v1.0/$metadata#servicePrincipals/$entity", spResp["@odata.context"])
	assert.Equal(t, sp.ID, spResp["id"])
	assert.Equal(t, "Test App", spResp["displayName"])
}

func TestServicePrincipalGetByAppID(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create an app (auto-creates SP)
	app := model.Application{
		DisplayName: "Test App",
		Description: "Test Desc",
	}
	createdApp, err := st.CreateApplication(ctx, app)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.Read.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL+"/v1.0/servicePrincipals/(appId="+createdApp.AppID+")", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var spResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &spResp)
	require.NoError(t, err)

	assert.Equal(t, "https://graph.microsoft.com/v1.0/$metadata#servicePrincipals/$entity", spResp["@odata.context"])
	assert.Equal(t, "Test App", spResp["displayName"])
}

func TestServicePrincipalGetNotFound(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.Read.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL+"/v1.0/servicePrincipals/nonexistent", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestServicePrincipalCreate(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create an app via store
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := st.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Delete the auto-created SP so we can manually create one
	sp, err := findSPByAppID(st, createdApp.AppID)
	require.NoError(t, err)
	err = st.DeleteServicePrincipal(ctx, sp.ID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	spJSON := `{"appId": "` + createdApp.AppID + `"}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/servicePrincipals", strings.NewReader(spJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Location"), "/v1.0/servicePrincipals/")

	var spResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &spResp)
	require.NoError(t, err)

	assert.Equal(t, "https://graph.microsoft.com/v1.0/$metadata#servicePrincipals/$entity", spResp["@odata.context"])
	assert.Contains(t, spResp, "id")
	assert.NotEmpty(t, spResp["id"])
	assert.Equal(t, createdApp.AppID, spResp["appId"])
}

func TestServicePrincipalCreateDuplicate(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create an app (auto-creates SP)
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := st.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Delete the auto-created SP so we can manually create one
	sp, err := findSPByAppID(st, createdApp.AppID)
	require.NoError(t, err)
	err = st.DeleteServicePrincipal(ctx, sp.ID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	spJSON := `{"appId": "` + createdApp.AppID + `"}`

	// First attempt should succeed
	req, err := http.NewRequest("POST", server.URL+"/v1.0/servicePrincipals", strings.NewReader(spJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Second attempt should fail with 409
	req2, err := http.NewRequest("POST", server.URL+"/v1.0/servicePrincipals", strings.NewReader(spJSON))
	require.NoError(t, err)
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusConflict, resp2.StatusCode)
}

func TestServicePrincipalCreateNoAppId(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	spJSON := `{}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/servicePrincipals", strings.NewReader(spJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestServicePrincipalCreateAppNotFound(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	spJSON := `{"appId": "nonexistent"}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/servicePrincipals", strings.NewReader(spJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestServicePrincipalUpdate(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create an app (auto-creates SP)
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := st.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Find the SP
	sp, err := findSPByAppID(st, createdApp.AppID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	patchJSON := `{"displayName": "Updated SP"}`

	req, err := http.NewRequest("PATCH", server.URL+"/v1.0/servicePrincipals/"+sp.ID, strings.NewReader(patchJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var spResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &spResp)
	require.NoError(t, err)

	assert.Equal(t, "Updated SP", spResp["displayName"])
}

func TestServicePrincipalDelete(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create an app
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := st.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Find the SP
	sp, err := findSPByAppID(st, createdApp.AppID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("DELETE", server.URL+"/v1.0/servicePrincipals/"+sp.ID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify it's deleted
	req2, err := http.NewRequest("GET", server.URL+"/v1.0/servicePrincipals/"+sp.ID, nil)
	require.NoError(t, err)
	req2.Header.Set("Authorization", "Bearer "+token)

	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp2.StatusCode)
}

func TestServicePrincipalListOwners(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create user
	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "testuser@example.com",
		Mail:              "testuser@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := st.CreateUser(ctx, user)
	require.NoError(t, err)

	// Create app (auto-creates SP)
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := st.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Find SP
	sp, err := findSPByAppID(st, createdApp.AppID)
	require.NoError(t, err)

	// Add owner via store
	err = st.AddSPOwner(ctx, sp.ID, createdUser.ID, "user")
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.Read.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL+"/v1.0/servicePrincipals/"+sp.ID+"/owners", nil)
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

func TestServicePrincipalAddOwner(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create user
	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "testuser@example.com",
		Mail:              "testuser@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := st.CreateUser(ctx, user)
	require.NoError(t, err)

	// Create app
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := st.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Find SP
	sp, err := findSPByAppID(st, createdApp.AppID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	refJSON := `{"@odata.id": "https://graph.microsoft.com/v1.0/directoryObjects/` + createdUser.ID + `"}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/servicePrincipals/"+sp.ID+"/owners/$ref", strings.NewReader(refJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestServicePrincipalRemoveOwner(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create user
	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "testuser@example.com",
		Mail:              "testuser@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := st.CreateUser(ctx, user)
	require.NoError(t, err)

	// Create app
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := st.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Find SP
	sp, err := findSPByAppID(st, createdApp.AppID)
	require.NoError(t, err)

	// Add owner via store
	err = st.AddSPOwner(ctx, sp.ID, createdUser.ID, "user")
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("DELETE", server.URL+"/v1.0/servicePrincipals/"+sp.ID+"/owners/"+createdUser.ID+"/$ref", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestServicePrincipalListMemberOf(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create a group
	mailEnabled := false
	securityEnabled := true
	group := model.Group{
		DisplayName:       "Test Group",
		MailEnabled:       &mailEnabled,
		SecurityEnabled:   &securityEnabled,
		MailNickname:      "testgroup",
	}
	createdGroup, err := st.CreateGroup(ctx, group)
	require.NoError(t, err)

	// Create an app (auto-creates SP)
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := st.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Find SP
	sp, err := findSPByAppID(st, createdApp.AppID)
	require.NoError(t, err)

	// Add SP as member of group
	err = st.AddMember(ctx, createdGroup.ID, sp.ID, "servicePrincipal")
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.Read.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL+"/v1.0/servicePrincipals/"+sp.ID+"/memberOf", nil)
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

func TestServicePrincipalListTransitiveMemberOf(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create nested groups
	mailEnabled := false
	securityEnabled := true
	parentGroup := model.Group{
		DisplayName:       "Parent Group",
		MailEnabled:       &mailEnabled,
		SecurityEnabled:   &securityEnabled,
		MailNickname:      "parentgroup",
	}
	createdParent, err := st.CreateGroup(ctx, parentGroup)
	require.NoError(t, err)

	childGroup := model.Group{
		DisplayName:       "Child Group",
		MailEnabled:       &mailEnabled,
		SecurityEnabled:   &securityEnabled,
		MailNickname:      "childgroup",
	}
	createdChild, err := st.CreateGroup(ctx, childGroup)
	require.NoError(t, err)

	// Add child as member of parent
	err = st.AddMember(ctx, createdParent.ID, createdChild.ID, "group")
	require.NoError(t, err)

	// Create an app (auto-creates SP)
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := st.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Find SP
	sp, err := findSPByAppID(st, createdApp.AppID)
	require.NoError(t, err)

	// Add SP as member of child group
	err = st.AddMember(ctx, createdChild.ID, sp.ID, "servicePrincipal")
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.Read.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL+"/v1.0/servicePrincipals/"+sp.ID+"/transitiveMemberOf", nil)
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
	assert.GreaterOrEqual(t, len(value), 1)
}

func TestServicePrincipalAppRoleAssignmentCRUD(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create two apps with appRoles
	enabled := true
	resourceApp := model.Application{
		DisplayName: "Resource App",
		AppRoles: []model.AppRole{
			{
				ID:                 "00000000-0000-0000-0000-000000000001",
				AllowedMemberTypes: []string{"Application"},
				DisplayName:        "Test Role",
				IsEnabled:          &enabled,
				Value:              "test.role",
			},
		},
	}
	createdResourceApp, err := st.CreateApplication(ctx, resourceApp)
	require.NoError(t, err)

	principalApp := model.Application{
		DisplayName: "Principal App",
	}
	createdPrincipalApp, err := st.CreateApplication(ctx, principalApp)
	require.NoError(t, err)

	// Find both SPs
	resourceSP, err := findSPByAppID(st, createdResourceApp.AppID)
	require.NoError(t, err)

	principalSP, err := findSPByAppID(st, createdPrincipalApp.AppID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"AppRoleAssignment.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// Create app role assignment
	assignmentJSON := `{
		"principalId": "` + principalSP.ID + `",
		"resourceId": "` + resourceSP.ID + `",
		"appRoleId": "00000000-0000-0000-0000-000000000001"
	}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/servicePrincipals/"+principalSP.ID+"/appRoleAssignments", strings.NewReader(assignmentJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Get assignment ID
	var assignmentResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &assignmentResp)
	require.NoError(t, err)

	assignmentID := assignmentResp["id"].(string)

	// List assignments
	req2, err := http.NewRequest("GET", server.URL+"/v1.0/servicePrincipals/"+principalSP.ID+"/appRoleAssignments", nil)
	require.NoError(t, err)
	req2.Header.Set("Authorization", "Bearer "+token)

	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	var listResp map[string]interface{}
	body2, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body2, &listResp)
	require.NoError(t, err)

	value := listResp["value"].([]interface{})
	assert.GreaterOrEqual(t, len(value), 1)

	// Delete assignment
	req3, err := http.NewRequest("DELETE", server.URL+"/v1.0/servicePrincipals/"+principalSP.ID+"/appRoleAssignments/"+assignmentID, nil)
	require.NoError(t, err)
	req3.Header.Set("Authorization", "Bearer "+token)

	resp3, err := http.DefaultClient.Do(req3)
	require.NoError(t, err)
	defer resp3.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp3.StatusCode)
}

func TestServicePrincipalAppRoleAssignedToCRUD(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create two apps with appRoles
	enabled := true
	resourceApp := model.Application{
		DisplayName: "Resource App",
		AppRoles: []model.AppRole{
			{
				ID:                 "00000000-0000-0000-0000-000000000001",
				AllowedMemberTypes: []string{"Application"},
				DisplayName:        "Test Role",
				IsEnabled:          &enabled,
				Value:              "test.role",
			},
		},
	}
	createdResourceApp, err := st.CreateApplication(ctx, resourceApp)
	require.NoError(t, err)

	principalApp := model.Application{
		DisplayName: "Principal App",
	}
	createdPrincipalApp, err := st.CreateApplication(ctx, principalApp)
	require.NoError(t, err)

	// Find both SPs
	resourceSP, err := findSPByAppID(st, createdResourceApp.AppID)
	require.NoError(t, err)

	principalSP, err := findSPByAppID(st, createdPrincipalApp.AppID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"AppRoleAssignment.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// Create app role assigned to (via resource endpoint)
	assignmentJSON := `{
		"principalId": "` + principalSP.ID + `",
		"resourceId": "` + resourceSP.ID + `",
		"appRoleId": "00000000-0000-0000-0000-000000000001"
	}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/servicePrincipals/"+resourceSP.ID+"/appRoleAssignedTo", strings.NewReader(assignmentJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Get assignment ID
	var assignmentResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &assignmentResp)
	require.NoError(t, err)

	assignmentID := assignmentResp["id"].(string)

	// List assigned to
	req2, err := http.NewRequest("GET", server.URL+"/v1.0/servicePrincipals/"+resourceSP.ID+"/appRoleAssignedTo", nil)
	require.NoError(t, err)
	req2.Header.Set("Authorization", "Bearer "+token)

	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	var listResp map[string]interface{}
	body2, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body2, &listResp)
	require.NoError(t, err)

	value := listResp["value"].([]interface{})
	assert.GreaterOrEqual(t, len(value), 1)

	// Delete assignment
	req3, err := http.NewRequest("DELETE", server.URL+"/v1.0/servicePrincipals/"+resourceSP.ID+"/appRoleAssignedTo/"+assignmentID, nil)
	require.NoError(t, err)
	req3.Header.Set("Authorization", "Bearer "+token)

	resp3, err := http.DefaultClient.Do(req3)
	require.NoError(t, err)
	defer resp3.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp3.StatusCode)
}

func TestServicePrincipalListOAuth2Grants(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create an app (auto-creates SP)
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := st.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Find SP
	sp, err := findSPByAppID(st, createdApp.AppID)
	require.NoError(t, err)

  	// Create an OAuth2PermissionGrant with clientId = SP's object ID
 	grant := model.OAuth2PermissionGrant{
 		ClientID:    sp.ID,
 		ConsentType: "AllPrincipals",
		ResourceID:  sp.ID,
		Scope:       "User.Read",
	}
	_, err = st.CreateOAuth2PermissionGrant(ctx, grant)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"DelegatedPermissionGrant.Read.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL+"/v1.0/servicePrincipals/"+sp.ID+"/oauth2PermissionGrants", nil)
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
	assert.GreaterOrEqual(t, len(value), 1)
}

func TestServicePrincipalAddPassword(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create app
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := st.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Find SP
	sp, err := findSPByAppID(st, createdApp.AppID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	passwordJSON := `{"displayName": "Test Secret"}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/servicePrincipals/"+sp.ID+"/addPassword", strings.NewReader(passwordJSON))
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

func TestServicePrincipalRemovePassword(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create app
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := st.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Find SP
	sp, err := findSPByAppID(st, createdApp.AppID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// First add a password to get keyId
	passwordJSON := `{"displayName": "Test Secret"}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/servicePrincipals/"+sp.ID+"/addPassword", strings.NewReader(passwordJSON))
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

	req2, err := http.NewRequest("POST", server.URL+"/v1.0/servicePrincipals/"+sp.ID+"/removePassword", strings.NewReader(removeJSON))
	require.NoError(t, err)
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp2.StatusCode)
}

func TestServicePrincipalAddKey(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create app
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := st.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Find SP
	sp, err := findSPByAppID(st, createdApp.AppID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	keyJSON := `{"displayName": "Test Key", "type": "AsymmetricX509Cert", "usage": "Verify"}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/servicePrincipals/"+sp.ID+"/addKey", strings.NewReader(keyJSON))
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

func TestServicePrincipalRemoveKey(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create app
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := st.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Find SP
	sp, err := findSPByAppID(st, createdApp.AppID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// First add a key to get keyId
	keyJSON := `{"displayName": "Test Key", "type": "AsymmetricX509Cert", "usage": "Verify"}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/servicePrincipals/"+sp.ID+"/addKey", strings.NewReader(keyJSON))
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

	req2, err := http.NewRequest("POST", server.URL+"/v1.0/servicePrincipals/"+sp.ID+"/removeKey", strings.NewReader(removeJSON))
	require.NoError(t, err)
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp2.StatusCode)
}

func TestServicePrincipalEmptyPolicies(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create app
	app := model.Application{
		DisplayName: "Test App",
	}
	createdApp, err := st.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Find SP
	sp, err := findSPByAppID(st, createdApp.AppID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Application.Read.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	policies := []string{
		"homeRealmDiscoveryPolicies",
		"claimsMappingPolicies",
		"tokenIssuancePolicies",
		"tokenLifetimePolicies",
	}

	for _, policy := range policies {
		req, err := http.NewRequest("GET", server.URL+"/v1.0/servicePrincipals/"+sp.ID+"/"+policy, nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var policyResp map[string]interface{}
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		err = json.Unmarshal(body, &policyResp)
		require.NoError(t, err)

		value := policyResp["value"].([]interface{})
		assert.Equal(t, 0, len(value))
		assert.Contains(t, policyResp, "@odata.context")
	}
}
