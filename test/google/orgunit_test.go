//go:build google

package google_e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleOrgUnit_CreateAndList(t *testing.T) {
	ts, _ := setupGoogleServer(t)
	client := makeClient()
	token := getGoogleToken(t, client, ts.URL)
	base := ts.URL + "/admin/directory/v1/customer/my_customer/orgunits"

	// Create org unit
	resp := googleRequest(t, client, http.MethodPost, base, token, `{"name":"Engineering","orgUnitPath":"engineering"}`)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := readBody(t, resp)
	assert.Equal(t, "Engineering", body["name"])
	assert.Equal(t, "engineering", body["orgUnitPath"])

	// List org units
	resp = googleRequest(t, client, http.MethodGet, base+"/", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = readBody(t, resp)
	orgUnits, ok := body["organizationUnits"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(orgUnits), 1)
}

func TestGoogleOrgUnit_GetByPath(t *testing.T) {
	ts, _ := setupGoogleServer(t)
	client := makeClient()
	token := getGoogleToken(t, client, ts.URL)
	base := ts.URL + "/admin/directory/v1/customer/my_customer/orgunits"

	// Create org unit
	resp := googleRequest(t, client, http.MethodPost, base, token, `{"name":"Sales","orgUnitPath":"sales"}`)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	readBody(t, resp)

	// Get by path
	resp = googleRequest(t, client, http.MethodGet, base+"/sales", token, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := readBody(t, resp)
	assert.Equal(t, "Sales", body["name"])
}
