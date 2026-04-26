//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/stretchr/testify/require"
)

// credCache caches one *azidentity.ClientSecretCredential per *TestServer so
// the credential is created once per test process instead of once per HTTP call.
var credCache sync.Map // map[*TestServer]*azidentity.ClientSecretCredential

// ptrString returns a pointer to the given string
func ptrString(s string) *string {
	return &s
}

// createTestUserSDK creates a user via the SDK and returns it.
func createTestUserSDK(t *testing.T, ts *TestServer, displayName, upn string) models.Userable {
	t.Helper()

	requestBody := models.NewUser()
	requestBody.SetDisplayName(&displayName)
	requestBody.SetUserPrincipalName(&upn)
	requestBody.SetMail(&upn)

	accountEnabled := true
	requestBody.SetAccountEnabled(&accountEnabled)

	passwordProfile := models.NewPasswordProfile()
	password := "Test1234!"
	passwordProfile.SetPassword(&password)
	requestBody.SetPasswordProfile(passwordProfile)

	userType := "Member"
	requestBody.SetUserType(&userType)

	result, err := ts.SDKClient.Users().Post(context.Background(), requestBody, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	return result
}

// createTestGroupSDK creates a group via the SDK and returns it.
func createTestGroupSDK(t *testing.T, ts *TestServer, displayName string) models.Groupable {
	t.Helper()

	requestBody := models.NewGroup()
	requestBody.SetDisplayName(&displayName)

	mailNickname := displayName // or derive from display name
	requestBody.SetMailNickname(&mailNickname)

	securityEnabled := true
	requestBody.SetSecurityEnabled(&securityEnabled)

	mailEnabled := false
	requestBody.SetMailEnabled(&mailEnabled)

	visibility := "Public"
	requestBody.SetVisibility(&visibility)

	result, err := ts.SDKClient.Groups().Post(context.Background(), requestBody, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	return result
}

// createTestApplicationSDK creates an application via the SDK and returns it.
func createTestApplicationSDK(t *testing.T, ts *TestServer, displayName string) models.Applicationable {
	t.Helper()

	requestBody := models.NewApplication()
	requestBody.SetDisplayName(&displayName)

	result, err := ts.SDKClient.Applications().Post(context.Background(), requestBody, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	return result
}

// createTestServicePrincipalSDK creates a service principal from an existing application via the SDK and returns it.
// It takes the application's appId as parameter.
func createTestServicePrincipalSDK(t *testing.T, ts *TestServer, appId string) models.ServicePrincipalable {
	t.Helper()

	requestBody := models.NewServicePrincipal()
	requestBody.SetAppId(&appId)

	result, err := ts.SDKClient.ServicePrincipals().Post(context.Background(), requestBody, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	return result
}

// getTestToken obtains an auth token for raw HTTP requests to the test server.
// The underlying ClientSecretCredential is cached per TestServer so it is only
// created once per test process instead of once per HTTP call.
func getTestToken(t *testing.T, ts *TestServer) string {
	t.Helper()

	cred, ok := credCache.Load(ts)
	if !ok {
		var err error
		var newCred *azidentity.ClientSecretCredential
		newCred, err = azidentity.NewClientSecretCredential(
			"sim-tenant-id", "sim-client-id", "sim-client-secret",
			&azidentity.ClientSecretCredentialOptions{
				ClientOptions: azcore.ClientOptions{
					Cloud: cloud.Configuration{
						ActiveDirectoryAuthorityHost: ts.BaseURL,
					},
					Transport: &httpTransport{client: ts.Server.Client()},
				},
				DisableInstanceDiscovery: true,
			},
		)
		require.NoError(t, err, "Failed to create credential")

		// Store the credential; use LoadOrStore to handle the race where
		// another goroutine may have stored one between our Load and here.
		actual, loaded := credCache.LoadOrStore(ts, newCred)
		if loaded {
			cred = actual
		} else {
			cred = newCred
		}
	}

	token, err := cred.(*azidentity.ClientSecretCredential).GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{"https://graph.microsoft.com/.default"},
	})
	require.NoError(t, err, "Failed to get token")

	return token.Token
}

// authedGet performs an authenticated GET request
func authedGet(t *testing.T, ts *TestServer, url string) *http.Response {
	t.Helper()
	token := getTestToken(t, ts)
	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := ts.Server.Client().Do(req)
	require.NoError(t, err)
	return resp
}

// authedPost performs an authenticated POST request with JSON body
func authedPost(t *testing.T, ts *TestServer, url string, body []byte) *http.Response {
	t.Helper()
	token := getTestToken(t, ts)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := ts.Server.Client().Do(req)
	require.NoError(t, err)
	return resp
}

// authedPatch performs an authenticated PATCH request with JSON body
func authedPatch(t *testing.T, ts *TestServer, url string, body []byte) *http.Response {
	t.Helper()
	token := getTestToken(t, ts)
	req, err := http.NewRequest("PATCH", url, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := ts.Server.Client().Do(req)
	require.NoError(t, err)
	return resp
}

// authedDelete performs an authenticated DELETE request
func authedDelete(t *testing.T, ts *TestServer, url string) *http.Response {
	t.Helper()
	token := getTestToken(t, ts)
	req, err := http.NewRequest("DELETE", url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := ts.Server.Client().Do(req)
	require.NoError(t, err)
	return resp
}

// readJSON reads the response body and unmarshals it as map[string]interface{}
func readJSON(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)
	return result
}
