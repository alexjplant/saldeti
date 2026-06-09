//go:build google

package google_e2e

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	gauth "github.com/saldeti/saldeti/internal/google/auth"
	"github.com/saldeti/saldeti/internal/google/handler"
	"github.com/saldeti/saldeti/internal/google/store"
)

func TestMain(m *testing.M) {
	gauth.SetSigningKey([]byte("test-secret-key-that-is-32bytes!"))
	os.Exit(m.Run())
}

func setupGoogleServer(t *testing.T) (*httptest.Server, store.Store) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	s := store.NewMemoryStore()
	s.RegisterClient(context.Background(), "test-client", "test-secret")
	r := gin.New()
	handler.RegisterRoutes(r, s)
	ts := httptest.NewTLSServer(r)
	t.Cleanup(ts.Close)
	return ts, s
}

func makeClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}

func getGoogleToken(t *testing.T, client *http.Client, baseURL string) string {
	t.Helper()
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", "test-client")
	form.Set("client_secret", "test-secret")
	resp, err := client.PostForm(baseURL+"/oauth2/v1/token", form)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "token request failed: %s", string(body))
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))
	token, _ := result["access_token"].(string)
	require.NotEmpty(t, token, "access_token was empty")
	return token
}

func googleRequest(t *testing.T, client *http.Client, method, url, token, body string) *http.Response {
	t.Helper()
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
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func readBody(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))
	return result
}
