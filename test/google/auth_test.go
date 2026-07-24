//go:build e2e && google

package google_e2e

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleToken_ClientCredentials(t *testing.T) {
	ts, _ := setupGoogleServer(t)
	client := makeClient()

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", "test-client")
	form.Set("client_secret", "test-secret")
	resp, err := client.PostForm(ts.URL+"/oauth2/v1/token", form)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	var result map[string]any
	require.NoError(t, json.Unmarshal(body, &result))
	assert.NotEmpty(t, result["access_token"])
	assert.Equal(t, "Bearer", result["token_type"])
	assert.NotZero(t, result["expires_in"])
}

func TestGoogleToken_InvalidCredentials(t *testing.T) {
	ts, _ := setupGoogleServer(t)
	client := makeClient()

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", "wrong-client")
	form.Set("client_secret", "wrong-secret")
	resp, err := client.PostForm(ts.URL+"/oauth2/v1/token", form)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var result map[string]any
	require.NoError(t, json.Unmarshal(body, &result))
	errField, ok := result["error"].(string)
	require.True(t, ok)
	assert.Equal(t, "invalid_client", errField)
}

func TestGoogleAuth_MiddlewareRejection(t *testing.T) {
	ts, _ := setupGoogleServer(t)
	client := makeClient()

	// Request without Authorization header should be rejected
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/admin/directory/v1/users", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
