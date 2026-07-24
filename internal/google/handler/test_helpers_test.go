package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/google/auth"
	"github.com/saldeti/saldeti/internal/google/store"
	"github.com/stretchr/testify/require"
)

func setupGoogleTestServer(t *testing.T) *httptest.Server {
	gin.SetMode(gin.TestMode)
	auth.SetSigningKey([]byte("test-secret-key-that-is-32bytes!"))
	s := store.NewMemoryStore()
	s.RegisterClient(context.Background(), "test-client", "test-secret")
	r := gin.New()
	RegisterRoutes(r, s)
	ts := httptest.NewServer(r)
	t.Cleanup(ts.Close)
	return ts
}

func getGoogleToken(t *testing.T, ts *httptest.Server) string {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", "test-client")
	form.Set("client_secret", "test-secret")
	resp, err := http.Post(ts.URL+"/oauth2/v1/token", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "token failed: %s", string(body))
	var result map[string]any
	json.Unmarshal(body, &result)
	token, _ := result["access_token"].(string)
	require.NotEmpty(t, token)
	return token
}

func doGoogleRequest(t *testing.T, method, url, token, body string) *http.Response {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func readBody(t *testing.T, resp *http.Response) map[string]any {
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()
	var result map[string]any
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)
	return result
}
