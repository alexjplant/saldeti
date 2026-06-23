package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleRoleCRUD(t *testing.T) {
	ts := setupGoogleTestServer(t)
	token := getGoogleToken(t, ts)
	base := ts.URL + "/admin/directory/v1/customer/my_customer"

	// Create role
	createJSON := `{"roleName":"TestRole","rolePrivileges":[{"privilegeName":"USERS_MANAGE","serviceId":"serviceAdmin"}]}`
	resp := doGoogleRequest(t, http.MethodPost, base+"/roles", token, createJSON)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := readBody(t, resp)
	assert.Equal(t, "TestRole", body["roleName"])
	roleID, _ := body["roleId"].(string)
	require.NotEmpty(t, roleID)

	// List roles
	resp = doGoogleRequest(t, http.MethodGet, base+"/roles", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	items, ok := body["items"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(items), 1)

	// Get role
	resp = doGoogleRequest(t, http.MethodGet, base+"/roles/"+roleID, token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	assert.Equal(t, "TestRole", body["roleName"])
	assert.Equal(t, roleID, body["roleId"])

	// Delete role
	resp = doGoogleRequest(t, http.MethodDelete, base+"/roles/"+roleID, token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Verify 404
	resp = doGoogleRequest(t, http.MethodGet, base+"/roles/"+roleID, token, "")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}
