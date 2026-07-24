//go:build e2e && google

package google_e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleUser_CreateGetDelete(t *testing.T) {
	ts, _ := setupGoogleServer(t)
	client := makeClient()
	token := getGoogleToken(t, client, ts.URL)
	base := ts.URL + "/admin/directory/v1"

	// Create user
	resp := googleRequest(t, client, http.MethodPost, base+"/users", token, `{"primaryEmail":"e2e-user@example.com","name":{"givenName":"E2E","familyName":"User"}}`)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := readBody(t, resp)
	userID, _ := body["id"].(string)
	require.NotEmpty(t, userID)
	assert.Equal(t, "e2e-user@example.com", body["primaryEmail"])

	// Get by ID
	resp = googleRequest(t, client, http.MethodGet, base+"/users/"+userID, token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	assert.Equal(t, "e2e-user@example.com", body["primaryEmail"])

	// Get by email
	resp = googleRequest(t, client, http.MethodGet, base+"/users/e2e-user@example.com", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	assert.Equal(t, userID, body["id"])

	// Delete user
	resp = googleRequest(t, client, http.MethodDelete, base+"/users/"+userID, token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Verify 404
	resp = googleRequest(t, client, http.MethodGet, base+"/users/"+userID, token, "")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestGoogleUser_List(t *testing.T) {
	ts, _ := setupGoogleServer(t)
	client := makeClient()
	token := getGoogleToken(t, client, ts.URL)
	base := ts.URL + "/admin/directory/v1"

	// Create user
	resp := googleRequest(t, client, http.MethodPost, base+"/users", token, `{"primaryEmail":"list-user@example.com","name":{"givenName":"List","familyName":"User"}}`)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	readBody(t, resp)

	// List users
	resp = googleRequest(t, client, http.MethodGet, base+"/users", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := readBody(t, resp)
	users, ok := body["users"].([]any)
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(users), 1)
}

func TestGoogleUser_Update(t *testing.T) {
	ts, _ := setupGoogleServer(t)
	client := makeClient()
	token := getGoogleToken(t, client, ts.URL)
	base := ts.URL + "/admin/directory/v1"

	// Create user
	resp := googleRequest(t, client, http.MethodPost, base+"/users", token, `{"primaryEmail":"update-user@example.com","name":{"givenName":"Update","familyName":"User"}}`)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := readBody(t, resp)
	userID, _ := body["id"].(string)
	require.NotEmpty(t, userID)

	// Update user
	resp = googleRequest(t, client, http.MethodPut, base+"/users/"+userID, token, `{"primaryEmail":"updated-user@example.com","name":{"givenName":"Updated","familyName":"User"}}`)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	assert.Equal(t, "updated-user@example.com", body["primaryEmail"])
}

func TestGoogleUser_Patch(t *testing.T) {
	ts, _ := setupGoogleServer(t)
	client := makeClient()
	token := getGoogleToken(t, client, ts.URL)
	base := ts.URL + "/admin/directory/v1"

	// Create user
	resp := googleRequest(t, client, http.MethodPost, base+"/users", token, `{"primaryEmail":"patch-user@example.com","name":{"givenName":"Patch","familyName":"User"}}`)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := readBody(t, resp)
	userID, _ := body["id"].(string)
	require.NotEmpty(t, userID)

	// Patch user
	resp = googleRequest(t, client, http.MethodPatch, base+"/users/"+userID, token, `{"name":{"givenName":"Patched","familyName":"User"}}`)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	nameObj, ok := body["name"].(map[string]any)
	require.True(t, ok, "expected name object in response")
	assert.Equal(t, "Patched", nameObj["givenName"])
}
