package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleMemberCRUD(t *testing.T) {
	ts := setupGoogleTestServer(t)
	token := getGoogleToken(t, ts)
	base := ts.URL + "/admin/directory/v1"

	// Create group first
	groupJSON := `{"email":"membergroup@example.com","name":"Member Test Group"}`
	resp := doGoogleRequest(t, http.MethodPost, base+"/groups", token, groupJSON)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := readBody(t, resp)
	groupID, ok := body["id"].(string)
	require.True(t, ok, "expected group id to be a string")
	require.NotEmpty(t, groupID)
	assert.Equal(t, "membergroup@example.com", body["email"])

	// Add member
	memberJSON := `{"email":"member@example.com","role":"MEMBER"}`
	resp = doGoogleRequest(t, http.MethodPost, base+"/groups/"+groupID+"/members", token, memberJSON)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	assert.Equal(t, "member@example.com", body["email"])
	assert.Equal(t, "MEMBER", body["role"])

	// List members
	resp = doGoogleRequest(t, http.MethodGet, base+"/groups/"+groupID+"/members", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	members, ok := body["members"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(members), 1)

	// Get member (memberKey is the email in the store)
	resp = doGoogleRequest(t, http.MethodGet, base+"/groups/"+groupID+"/members/member@example.com", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	assert.Equal(t, "member@example.com", body["email"])

	// Has member
	resp = doGoogleRequest(t, http.MethodGet, base+"/groups/"+groupID+"/hasMember/member@example.com", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	assert.Equal(t, true, body["isMember"])

	// Remove member
	resp = doGoogleRequest(t, http.MethodDelete, base+"/groups/"+groupID+"/members/member@example.com", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Verify member removed
	resp = doGoogleRequest(t, http.MethodGet, base+"/groups/"+groupID+"/hasMember/member@example.com", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	assert.Equal(t, false, body["isMember"])
}
