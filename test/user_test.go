//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_FullUserLifecycle tests creating, reading, updating, and deleting a user
func TestE2E_FullUserLifecycle(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// 1. Create user via SDK
	newUser := models.NewUser()
	displayName := "Test User"
	upn := "test.user@saldeti.local"
	mail := upn
	dept := "Engineering"
	jobTitle := "Developer"
	enabled := true
	newUser.SetDisplayName(&displayName)
	newUser.SetUserPrincipalName(&upn)
	newUser.SetMail(&mail)
	newUser.SetAccountEnabled(&enabled)
	newUser.SetDepartment(&dept)
	newUser.SetJobTitle(&jobTitle)

	passwordProfile := models.NewPasswordProfile()
	password := "Test1234!"
	passwordProfile.SetPassword(&password)
	newUser.SetPasswordProfile(passwordProfile)

	userType := "Member"
	newUser.SetUserType(&userType)

	createdUser, err := tss.SDKClient.Users().Post(ctx, newUser, nil)
	require.NoError(t, err)
	require.NotNil(t, createdUser)

	userID := *createdUser.GetId()

	// 2. GET user by ID
	fetchedUser, err := tss.SDKClient.Users().ByUserId(userID).Get(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, "Test User", *fetchedUser.GetDisplayName())
	assert.Equal(t, "Engineering", *fetchedUser.GetDepartment())

	// 3. PATCH user (change department)
	patchUser := models.NewUser()
	newDept := "Marketing"
	patchUser.SetDepartment(&newDept)

	updatedUser, err := tss.SDKClient.Users().ByUserId(userID).Patch(ctx, patchUser, nil)
	require.NoError(t, err)
	assert.Equal(t, "Marketing", *updatedUser.GetDepartment())

	// 4. GET again to verify
	verifyUser, err := tss.SDKClient.Users().ByUserId(userID).Get(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, "Marketing", *verifyUser.GetDepartment())

	// 5. DELETE
	err = tss.SDKClient.Users().ByUserId(userID).Delete(ctx, nil)
	require.NoError(t, err)

	// 6. GET again, expect error (404)
	_, err = tss.SDKClient.Users().ByUserId(userID).Get(ctx, nil)
	require.Error(t, err, "Expected error getting deleted user")
}

// TestE2E_GetUserByUPN tests getting a user by user principal name
func TestE2E_GetUserByUPN(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// Create user via SDK helper
	upn := "testupnunique@saldeti.local"
	createdUser := createTestUserSDK(t, tss, "Test User UPN", upn)
	userID := *createdUser.GetId()

	// List users with filter
	filter := fmt.Sprintf("userPrincipalName eq '%s'", upn)
	result, err := tss.SDKClient.Users().Get(ctx, &users.UsersRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
			Filter: &filter,
		},
	})
	require.NoError(t, err)

	userList := result.GetValue()
	require.Len(t, userList, 1)
	assert.Equal(t, upn, *userList[0].GetUserPrincipalName())
	assert.Equal(t, userID, *userList[0].GetId())
}

// TestE2E_CreateUserValidation tests user creation validation
func TestE2E_CreateUserValidation(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// 1. Try to create user without displayName via SDK
	user := models.NewUser()
	upn := "test@saldeti.local"
	user.SetUserPrincipalName(&upn)
	user.SetMail(&upn)
	enabled := true
	user.SetAccountEnabled(&enabled)
	_, err := tss.SDKClient.Users().Post(ctx, user, nil)
	require.Error(t, err, "Expected error creating user without displayName")

	// 2. Create a valid user via SDK
	user.SetDisplayName(ptrString("Test User"))
	created, err := tss.SDKClient.Users().Post(ctx, user, nil)
	require.NoError(t, err, "Expected success creating valid user")
	require.NotNil(t, created)

	// 3. Try to create user with duplicate UPN via SDK
	dup := models.NewUser()
	dup.SetDisplayName(ptrString("Test User"))
	dup.SetUserPrincipalName(&upn)
	dup.SetMail(&upn)
	dup.SetAccountEnabled(&enabled)
	_, err = tss.SDKClient.Users().Post(ctx, dup, nil)
	require.Error(t, err, "Expected error creating user with duplicate UPN")
}

// TestE2E_SDKCredentialGetToken tests the SimulatorCredential
func TestE2E_SDKCredentialGetToken(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	cred, err := azidentity.NewClientSecretCredential(
		"sim-tenant-id", "sim-client-id", "sim-client-secret",
		&azidentity.ClientSecretCredentialOptions{
			ClientOptions: azcore.ClientOptions{
				Cloud: cloud.Configuration{
					ActiveDirectoryAuthorityHost: tss.BaseURL,
				},
				Transport: &httpTransport{client: tss.Server.Client()},
			},
			DisableInstanceDiscovery: true,
		},
	)
	if err != nil {
		t.Fatalf("Failed to create credential: %v", err)
	}

	ctx := context.Background()
	token, err := cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://graph.microsoft.com/.default"},
	})
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	if token.Token == "" {
		t.Error("Token is empty")
	}
	if token.ExpiresOn.IsZero() {
		t.Error("ExpiresOn is zero")
	}
	if token.ExpiresOn.Before(time.Now()) {
		t.Error("ExpiresOn is in the past")
	}

	// Verify the token is valid by making an authenticated request
	req, _ := http.NewRequest("GET", tss.BaseURL+"/v1.0/users", nil)
	req.Header.Set("Authorization", "Bearer "+token.Token)

	resp, err := tss.Server.Client().Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 with token, got %d", resp.StatusCode)
	}
}

// TestE2E_PaginationWorkflow tests pagination with $top
func TestE2E_PaginationWorkflow(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// Create 15 users via SDK
	for i := 1; i <= 15; i++ {
		upn := fmt.Sprintf("user%d@saldeti.local", i)
		displayName := fmt.Sprintf("User %d", i)
		createTestUserSDK(t, tss, displayName, upn)
	}

	// Get first page with $top=5
	top := int32(5)
	result, err := tss.SDKClient.Users().Get(ctx, &users.UsersRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
			Top: &top,
		},
	})
	require.NoError(t, err)

	page1Users := result.GetValue()
	assert.Len(t, page1Users, 5)

	// Verify next link exists
	nextLink := result.GetOdataNextLink()
	require.NotNil(t, nextLink, "Expected @odata.nextLink in response")
	require.NotEmpty(t, *nextLink, "Expected non-empty next link")

	// Verify total count by getting all users with a large top
	allTop := int32(999)
	allResult, err := tss.SDKClient.Users().Get(ctx, &users.UsersRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
			Top: &allTop,
		},
	})
	require.NoError(t, err)
	assert.Len(t, allResult.GetValue(), 15)
}
