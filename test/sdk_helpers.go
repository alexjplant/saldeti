//go:build e2e

package e2e

import (
	"context"
	"testing"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/stretchr/testify/require"
)

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
