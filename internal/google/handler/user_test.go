package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleUserCRUD(t *testing.T) {
	ts := setupGoogleTestServer(t)
	token := getGoogleToken(t, ts)
	base := ts.URL + "/admin/directory/v1"

	// Create user
	createJSON := `{"primaryEmail":"test@example.com","name":{"givenName":"Test","familyName":"User"}}`
	resp := doGoogleRequest(t, http.MethodPost, base+"/users", token, createJSON)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := readBody(t, resp)
	userID, _ := body["id"].(string)
	require.NotEmpty(t, userID)
	assert.Equal(t, "test@example.com", body["primaryEmail"])

	// Get by ID
	resp = doGoogleRequest(t, http.MethodGet, base+"/users/"+userID, token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	assert.Equal(t, "test@example.com", body["primaryEmail"])

	// Get by email
	resp = doGoogleRequest(t, http.MethodGet, base+"/users/test@example.com", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	assert.Equal(t, userID, body["id"])

	// List users
	resp = doGoogleRequest(t, http.MethodGet, base+"/users", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	users, ok := body["users"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(users), 1)

	// Update user
	updateJSON := `{"primaryEmail":"updated@example.com","name":{"givenName":"Updated","familyName":"User"}}`
	resp = doGoogleRequest(t, http.MethodPut, base+"/users/"+userID, token, updateJSON)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	assert.Equal(t, "updated@example.com", body["primaryEmail"])

	// Patch user
	patchJSON := `{"name":{"givenName":"Patched","familyName":"User"}}`
	resp = doGoogleRequest(t, http.MethodPatch, base+"/users/updated@example.com", token, patchJSON)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	nameObj, ok := body["name"].(map[string]interface{})
	require.True(t, ok, "expected name object in response")
	assert.Equal(t, "Patched", nameObj["givenName"])

	// Delete user
	resp = doGoogleRequest(t, http.MethodDelete, base+"/users/"+userID, token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Verify 404 after delete
	resp = doGoogleRequest(t, http.MethodGet, base+"/users/"+userID, token, "")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestGoogleUserAliases(t *testing.T) {
	ts := setupGoogleTestServer(t)
	token := getGoogleToken(t, ts)
	base := ts.URL + "/admin/directory/v1"

	// Create user
	createJSON := `{"primaryEmail":"aliastest@example.com","name":{"givenName":"Alias","familyName":"Test"}}`
	resp := doGoogleRequest(t, http.MethodPost, base+"/users", token, createJSON)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := readBody(t, resp)
	userID, _ := body["id"].(string)
	require.NotEmpty(t, userID)

	// Add alias
	aliasJSON := `{"alias":"myalias@example.com"}`
	resp = doGoogleRequest(t, http.MethodPost, base+"/users/"+userID+"/aliases", token, aliasJSON)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	assert.Equal(t, "myalias@example.com", body["alias"])

	// List aliases
	resp = doGoogleRequest(t, http.MethodGet, base+"/users/"+userID+"/aliases", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	aliases, ok := body["aliases"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(aliases), 1)

	// Remove alias
	resp = doGoogleRequest(t, http.MethodDelete, base+"/users/"+userID+"/aliases/myalias@example.com", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Verify alias removed
	resp = doGoogleRequest(t, http.MethodGet, base+"/users/"+userID+"/aliases", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	aliases, ok = body["aliases"].([]interface{})
	require.True(t, ok)
	assert.Len(t, aliases, 0)
}

func TestGoogleUserMakeAdmin(t *testing.T) {
	ts := setupGoogleTestServer(t)
	token := getGoogleToken(t, ts)
	base := ts.URL + "/admin/directory/v1"

	// Create user
	createJSON := `{"primaryEmail":"admin@example.com","name":{"givenName":"Admin","familyName":"User"}}`
	resp := doGoogleRequest(t, http.MethodPost, base+"/users", token, createJSON)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := readBody(t, resp)
	userID, _ := body["id"].(string)
	require.NotEmpty(t, userID)

	// Make admin
	adminJSON := `{"status":true}`
	resp = doGoogleRequest(t, http.MethodPost, base+"/users/"+userID+"/makeAdmin", token, adminJSON)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Verify isAdmin by getting user
	resp = doGoogleRequest(t, http.MethodGet, base+"/users/"+userID, token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	assert.Equal(t, true, body["isAdmin"])
}
