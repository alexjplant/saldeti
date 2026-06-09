//go:build google

package google_e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleGroup_CRUD(t *testing.T) {
	ts, _ := setupGoogleServer(t)
	client := makeClient()
	token := getGoogleToken(t, client, ts.URL)
	base := ts.URL + "/admin/directory/v1"

	// Create group
	resp := googleRequest(t, client, http.MethodPost, base+"/groups", token, `{"email":"e2e-group@example.com","name":"E2E Group"}`)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := readBody(t, resp)
	groupID, _ := body["id"].(string)
	require.NotEmpty(t, groupID)
	assert.Equal(t, "e2e-group@example.com", body["email"])

	// Get group
	resp = googleRequest(t, client, http.MethodGet, base+"/groups/"+groupID, token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	assert.Equal(t, "e2e-group@example.com", body["email"])

	// List groups
	resp = googleRequest(t, client, http.MethodGet, base+"/groups", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	groups, ok := body["groups"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(groups), 1)

	// Update group
	resp = googleRequest(t, client, http.MethodPut, base+"/groups/"+groupID, token, `{"email":"updated-group@example.com","name":"Updated Group"}`)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	assert.Equal(t, "updated-group@example.com", body["email"])
	assert.Equal(t, "Updated Group", body["name"])

	// Delete group
	resp = googleRequest(t, client, http.MethodDelete, base+"/groups/"+groupID, token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Verify 404
	resp = googleRequest(t, client, http.MethodGet, base+"/groups/"+groupID, token, "")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestGoogleGroup_Members(t *testing.T) {
	ts, _ := setupGoogleServer(t)
	client := makeClient()
	token := getGoogleToken(t, client, ts.URL)
	base := ts.URL + "/admin/directory/v1"

	// Create group
	resp := googleRequest(t, client, http.MethodPost, base+"/groups", token, `{"email":"members-group@example.com","name":"Members Group"}`)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := readBody(t, resp)
	groupID, _ := body["id"].(string)
	require.NotEmpty(t, groupID)

	// Add member
	resp = googleRequest(t, client, http.MethodPost, base+"/groups/"+groupID+"/members", token, `{"email":"member@example.com","role":"MEMBER"}`)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	assert.Equal(t, "member@example.com", body["email"])
	assert.Equal(t, "MEMBER", body["role"])

	// List members
	resp = googleRequest(t, client, http.MethodGet, base+"/groups/"+groupID+"/members", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	members, ok := body["members"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(members), 1)

	// Remove member
	resp = googleRequest(t, client, http.MethodDelete, base+"/groups/"+groupID+"/members/member@example.com", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}
