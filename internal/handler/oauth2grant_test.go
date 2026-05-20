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

func TestOAuth2GrantCreate(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	server := httptest.NewServer(router)
	defer server.Close()

	// Mint token with correct scope
	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"DelegatedPermissionGrant.ReadWrite.All"}, []string{"Directory"}, time.Hour, "", "")
	require.NoError(t, err)

	// POST to /v1.0/oauth2PermissionGrants
	body := `{"clientId":"test-client-id","consentType":"AllPrincipals","resourceId":"resource-123","scope":"User.Read Mail.Read"}`
	req, err := http.NewRequest("POST", server.URL+"/v1.0/oauth2PermissionGrants", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 201, Location header, response body has id, clientId, scope, @odata.context
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Location"), "/v1.0/oauth2PermissionGrants/")

	var grantResp map[string]interface{}
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(bodyBytes, &grantResp)
	require.NoError(t, err)

	assert.Contains(t, grantResp, "id")
	assert.NotEmpty(t, grantResp["id"])
	assert.Equal(t, "test-client-id", grantResp["clientId"])
	assert.Equal(t, "AllPrincipals", grantResp["consentType"])
	assert.Equal(t, "resource-123", grantResp["resourceId"])
	assert.Equal(t, "User.Read Mail.Read", grantResp["scope"])
	assert.Equal(t, "https://graph.microsoft.com/v1.0/$metadata#oauth2PermissionGrants/$entity", grantResp["@odata.context"])
}

func TestOAuth2GrantList(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create 2 grants via store directly
	grant1 := model.OAuth2PermissionGrant{
		ClientID:    "client-1",
		ConsentType: "AllPrincipals",
		ResourceID:  "resource-1",
		Scope:       "User.Read",
	}
	grant2 := model.OAuth2PermissionGrant{
		ClientID:    "client-2",
		ConsentType: "Principal",
		PrincipalID: "user-1",
		ResourceID:  "resource-2",
		Scope:       "Mail.Read",
	}
	_, err := st.CreateOAuth2PermissionGrant(ctx, grant1)
	require.NoError(t, err)
	_, err = st.CreateOAuth2PermissionGrant(ctx, grant2)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"DelegatedPermissionGrant.ReadWrite.All"}, []string{"Directory"}, time.Hour, "", "")
	require.NoError(t, err)

	// GET /v1.0/oauth2PermissionGrants
	req, err := http.NewRequest("GET", server.URL+"/v1.0/oauth2PermissionGrants", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 200, @odata.context, value array has 2 items
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	assert.Equal(t, "https://graph.microsoft.com/v1.0/$metadata#oauth2PermissionGrants", listResp["@odata.context"])
	grants := listResp["value"].([]interface{})
	assert.Len(t, grants, 2)
}

func TestOAuth2GrantGet(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create a grant via store
	grant := model.OAuth2PermissionGrant{
		ClientID:    "test-client-id",
		ConsentType: "AllPrincipals",
		ResourceID:  "resource-123",
		Scope:       "User.Read Mail.Read",
	}
	createdGrant, err := st.CreateOAuth2PermissionGrant(ctx, grant)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"DelegatedPermissionGrant.ReadWrite.All"}, []string{"Directory"}, time.Hour, "", "")
	require.NoError(t, err)

	// GET /v1.0/oauth2PermissionGrants/{id}
	req, err := http.NewRequest("GET", server.URL+"/v1.0/oauth2PermissionGrants/"+createdGrant.ID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 200, @odata.context contains "$entity", fields match
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var grantResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &grantResp)
	require.NoError(t, err)

	assert.Contains(t, grantResp["@odata.context"], "$entity")
	assert.Equal(t, createdGrant.ID, grantResp["id"])
	assert.Equal(t, "test-client-id", grantResp["clientId"])
	assert.Equal(t, "AllPrincipals", grantResp["consentType"])
	assert.Equal(t, "resource-123", grantResp["resourceId"])
	assert.Equal(t, "User.Read Mail.Read", grantResp["scope"])
}

func TestOAuth2GrantGetNotFound(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"DelegatedPermissionGrant.ReadWrite.All"}, []string{"Directory"}, time.Hour, "", "")
	require.NoError(t, err)

	// GET /v1.0/oauth2PermissionGrants/nonexistent
	req, err := http.NewRequest("GET", server.URL+"/v1.0/oauth2PermissionGrants/nonexistent", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 404
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestOAuth2GrantUpdate(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create a grant
	grant := model.OAuth2PermissionGrant{
		ClientID:    "test-client-id",
		ConsentType: "AllPrincipals",
		ResourceID:  "resource-123",
		Scope:       "User.Read",
	}
	createdGrant, err := st.CreateOAuth2PermissionGrant(ctx, grant)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"DelegatedPermissionGrant.ReadWrite.All"}, []string{"Directory"}, time.Hour, "", "")
	require.NoError(t, err)

	// PATCH it with {"scope":"User.Read.Write"}
	patchBody := `{"scope":"User.Read.Write"}`
	req, err := http.NewRequest("PATCH", server.URL+"/v1.0/oauth2PermissionGrants/"+createdGrant.ID, strings.NewReader(patchBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 200, updated scope in response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var grantResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &grantResp)
	require.NoError(t, err)

	assert.Equal(t, "User.Read.Write", grantResp["scope"])
}

func TestOAuth2GrantDelete(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create a grant
	grant := model.OAuth2PermissionGrant{
		ClientID:    "test-client-id",
		ConsentType: "AllPrincipals",
		ResourceID:  "resource-123",
		Scope:       "User.Read",
	}
	createdGrant, err := st.CreateOAuth2PermissionGrant(ctx, grant)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"DelegatedPermissionGrant.ReadWrite.All"}, []string{"Directory"}, time.Hour, "", "")
	require.NoError(t, err)

	// DELETE it
	req, err := http.NewRequest("DELETE", server.URL+"/v1.0/oauth2PermissionGrants/"+createdGrant.ID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 204
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// GET it again, assert 404
	getReq, err := http.NewRequest("GET", server.URL+"/v1.0/oauth2PermissionGrants/"+createdGrant.ID, nil)
	require.NoError(t, err)
	getReq.Header.Set("Authorization", "Bearer "+token)

	getResp, err := http.DefaultClient.Do(getReq)
	require.NoError(t, err)
	defer getResp.Body.Close()

	assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
}
