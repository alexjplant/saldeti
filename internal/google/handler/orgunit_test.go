package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleOrgUnitCRUD(t *testing.T) {
	ts := setupGoogleTestServer(t)
	token := getGoogleToken(t, ts)
	base := ts.URL + "/admin/directory/v1/customer/my_customer/orgunits"

	// Create OU
	createJSON := `{"name":"Engineering","orgUnitPath":"engineering"}`
	resp := doGoogleRequest(t, http.MethodPost, base, token, createJSON)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := readBody(t, resp)
	assert.Equal(t, "Engineering", body["name"])
	assert.Equal(t, "engineering", body["orgUnitPath"])
	ouID, _ := body["orgUnitId"].(string)
	require.NotEmpty(t, ouID)

	// List OUs (trailing slash triggers list via wildcard empty path)
	resp = doGoogleRequest(t, http.MethodGet, base+"/", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	orgUnits, ok := body["organizationUnits"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(orgUnits), 1)

	// Get OU by path
	resp = doGoogleRequest(t, http.MethodGet, base+"/engineering", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	assert.Equal(t, "Engineering", body["name"])

	// Update OU
	updateJSON := `{"name":"Engineering Updated","orgUnitPath":"engineering"}`
	resp = doGoogleRequest(t, http.MethodPut, base+"/engineering", token, updateJSON)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	assert.Equal(t, "Engineering Updated", body["name"])

	// Delete OU
	resp = doGoogleRequest(t, http.MethodDelete, base+"/engineering", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Verify 404
	resp = doGoogleRequest(t, http.MethodGet, base+"/engineering", token, "")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}