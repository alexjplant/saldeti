package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleGroupCRUD(t *testing.T) {
	ts := setupGoogleTestServer(t)
	token := getGoogleToken(t, ts)
	base := ts.URL + "/admin/directory/v1"

	// Create group
	createJSON := `{"email":"group@example.com","name":"Test Group"}`
	resp := doGoogleRequest(t, http.MethodPost, base+"/groups", token, createJSON)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := readBody(t, resp)
	groupID, _ := body["id"].(string)
	require.NotEmpty(t, groupID)
	assert.Equal(t, "group@example.com", body["email"])

	// Get group
	resp = doGoogleRequest(t, http.MethodGet, base+"/groups/"+groupID, token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	assert.Equal(t, "group@example.com", body["email"])

	// List groups
	resp = doGoogleRequest(t, http.MethodGet, base+"/groups", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	groups, ok := body["groups"].([]any)
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(groups), 1)

	// Update group
	updateJSON := `{"email":"updated-group@example.com","name":"Updated Group"}`
	resp = doGoogleRequest(t, http.MethodPut, base+"/groups/"+groupID, token, updateJSON)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	assert.Equal(t, "updated-group@example.com", body["email"])
	assert.Equal(t, "Updated Group", body["name"])

	// Patch group
	patchJSON := `{"description":"A patched description"}`
	resp = doGoogleRequest(t, http.MethodPatch, base+"/groups/updated-group@example.com", token, patchJSON)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	assert.Equal(t, "A patched description", body["description"])

	// Delete group
	resp = doGoogleRequest(t, http.MethodDelete, base+"/groups/"+groupID, token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Verify 404 after delete
	resp = doGoogleRequest(t, http.MethodGet, base+"/groups/"+groupID, token, "")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}
