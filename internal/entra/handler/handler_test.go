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

	"github.com/saldeti/saldeti/internal/entra/auth"
	"github.com/saldeti/saldeti/internal/entra/model"
	"github.com/saldeti/saldeti/internal/entra/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Set a test signing key for all tests
	auth.SetSigningKey([]byte("test-signing-key-32-bytes-long"))
	m.Run()
}

func TestTokenEndpoint(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)

	// Seed a client
	ctx := context.Background()
	err := store.RegisterClient(ctx, "test-client", "test-secret", "test-tenant")
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	// Test client_credentials grant
	formData := strings.NewReader("grant_type=client_credentials&client_id=test-client&client_secret=test-secret&scope=User.Read")
	req, err := http.NewRequest("POST", server.URL+"/test-tenant/oauth2/v2.0/token", formData)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var tokenResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &tokenResp)
	require.NoError(t, err)

	assert.Equal(t, "Bearer", tokenResp["token_type"])
	assert.Equal(t, float64(3600), tokenResp["expires_in"])
	assert.Contains(t, tokenResp, "access_token")
	// client_credentials grant does not return refresh_token
	assert.NotContains(t, tokenResp, "refresh_token")

	// Test invalid client credentials
	formData = strings.NewReader("grant_type=client_credentials&client_id=wrong&client_secret=wrong&scope=User.Read")
	req, err = http.NewRequest("POST", server.URL+"/test-tenant/oauth2/v2.0/token", formData)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestMeEndpoint(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)

	// Create a test user
	ctx := context.Background()
	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "test@example.com",
		Mail:              "test@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	// Mint a token for the user
	token, err := auth.MintToken("test-tenant", "test-client", "test@example.com", []string{"User.Read"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Test /me endpoint with valid token
	req, err := http.NewRequest("GET", server.URL+"/v1.0/me", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var userResp model.User
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &userResp)
	require.NoError(t, err)

	assert.Equal(t, createdUser.ID, userResp.ID)
	assert.Equal(t, "test@example.com", userResp.UserPrincipalName)
	assert.Equal(t, "Test User", userResp.DisplayName)

	// Test /me endpoint without token
	req, err = http.NewRequest("GET", server.URL+"/v1.0/me", nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Test /me endpoint with invalid token
	req, err = http.NewRequest("GET", server.URL+"/v1.0/me", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestTokenGrantTypes(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)

	// Seed a client for auth code and refresh token tests
	ctx := context.Background()
	err := store.RegisterClient(ctx, "test-client", "test-secret", "test-tenant")
	require.NoError(t, err)

	// Create a test user for authorization_code grant
	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "test@example.com",
		Mail:              "test@example.com",
		AccountEnabled:    &accountEnabled,
	}
	_, err = store.CreateUser(ctx, user)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	testCases := []struct {
		name       string
		formData   string
		expectCode int
	}{
		{
			name:       "authorization_code grant",
			formData:   "grant_type=authorization_code&client_id=test-client&client_secret=test-secret&code=test@example.com&scope=User.Read",
			expectCode: http.StatusOK,
		},
		{
			name:       "refresh_token grant",
			formData:   "grant_type=refresh_token&client_id=test-client&refresh_token=invalidtoken&scope=User.Read",
			expectCode: http.StatusBadRequest, // Invalid refresh token
		},
		{
			name:       "unsupported grant type",
			formData:   "grant_type=password&username=test&password=test",
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "missing required fields",
			formData:   "grant_type=client_credentials",
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formData := strings.NewReader(tc.formData)
			req, err := http.NewRequest("POST", server.URL+"/test-tenant/oauth2/v2.0/token", formData)
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectCode, resp.StatusCode)
		})
	}
}

func TestMeEndpointWithClientCredentials(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Register a client so the token endpoint works
	err := st.RegisterClient(ctx, "test-client", "test-secret", "test-tenant")
	require.NoError(t, err)

	// Create an application (which auto-creates an SP)
	app := model.Application{
		DisplayName:    "Test App",
		SignInAudience: "AzureADMyOrg",
	}
	createdApp, err := st.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Mint a client_credentials token (subject = clientID, roles = ["Application"])
	token, err := auth.MintToken("test-tenant", createdApp.AppID, createdApp.AppID, []string{"User.Read"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	// Call GET /v1.0/me with the client_credentials token
	req, err := http.NewRequest("GET", server.URL+"/v1.0/me", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var spResp map[string]interface{}
	err = json.Unmarshal(body, &spResp)
	require.NoError(t, err)

	// Should have @odata.type = servicePrincipal
	assert.Equal(t, "#microsoft.graph.servicePrincipal", spResp["@odata.type"])

	// The "id" field must be the SP's directory object ID, NOT the client/app ID
	sp, err := st.GetServicePrincipalByAppID(ctx, createdApp.AppID)
	require.NoError(t, err)
	assert.Equal(t, sp.ID, spResp["id"])
	assert.Equal(t, createdApp.AppID, spResp["appId"])

	// Should be a full SP object, not just 3 fields
	assert.Contains(t, spResp, "displayName")
}

func TestMeEndpointWithClientCredentialsNoSP(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Register a client so the token endpoint works, but do NOT create an application.
	// This still auto-creates an SP for "test-client", so we must mint the token with
	// a different clientID that has no SP.
	err := st.RegisterClient(ctx, "test-client", "test-secret", "test-tenant")
	require.NoError(t, err)

	// Mint a token with Application role using a clientID that has NO SP
	token, err := auth.MintToken("test-tenant", "orphan-client", "orphan-client", []string{"User.Read"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// Verify no SP exists for this client
	_, err = st.GetServicePrincipalByAppID(ctx, "orphan-client")
	assert.ErrorIs(t, err, store.ErrServicePrincipalNotFound)

	server := httptest.NewServer(router)
	defer server.Close()

	// Call GET /v1.0/me with the client_credentials token
	req, err := http.NewRequest("GET", server.URL+"/v1.0/me", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var errResp map[string]interface{}
	err = json.Unmarshal(body, &errResp)
	require.NoError(t, err)

	inner, ok := errResp["error"].(map[string]interface{})
	require.True(t, ok, "response should have .error object")
	assert.Equal(t, "ResourceNotFound", inner["code"])
}
