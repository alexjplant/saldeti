//go:build e2e

package e2e

import (
	"context"
	"strings"
	"testing"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_ODataFilterScenarios tests various OData filter scenarios
func TestE2E_ODataFilterScenarios(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// Create 3 users with different departments via SDK
	testUsers := []struct {
		name       string
		department string
		enabled    bool
	}{
		{"Alice", "Engineering", true},
		{"Bob", "Marketing", true},
		{"Charlie", "HR", false},
	}

	for _, u := range testUsers {
		requestBody := models.NewUser()
		requestBody.SetDisplayName(&u.name)
		upn := strings.ToLower(u.name) + "@saldeti.local"
		requestBody.SetUserPrincipalName(&upn)
		requestBody.SetMail(&upn)
		requestBody.SetDepartment(&u.department)
		requestBody.SetAccountEnabled(&u.enabled)

		passwordProfile := models.NewPasswordProfile()
		password := "Test1234!"
		passwordProfile.SetPassword(&password)
		requestBody.SetPasswordProfile(passwordProfile)

		userType := "Member"
		requestBody.SetUserType(&userType)

		_, err := tss.SDKClient.Users().Post(ctx, requestBody, nil)
		require.NoError(t, err, "Failed to create user %s", u.name)
	}

	// Filter by `accountEnabled eq true` via SDK
	filter := "accountEnabled eq true"
	result, err := tss.SDKClient.Users().Get(ctx, &users.UsersRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
			Filter: &filter,
		},
	})
	require.NoError(t, err, "Failed to filter by accountEnabled")

	enabledUsers := result.GetValue()
	assert.Len(t, enabledUsers, 2, "Expected 2 enabled users (Alice and Bob)")

	// Use `$count=true` via SDK
	count := true
	countResult, err := tss.SDKClient.Users().Get(ctx, &users.UsersRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
			Count: &count,
		},
	})
	require.NoError(t, err, "Failed to get count")

	// Check if @odata.count is available via OData additional data
	// The SDK response should have GetOdataCount() or similar method
	countUsers := countResult.GetValue()
	// Verify we have the expected number of users
	assert.Len(t, countUsers, 3, "Expected 3 users total")

	// Try to access OData count from additional data
	additionalData := countResult.GetAdditionalData()
	if odataCount, ok := additionalData["@odata.count"].(float64); ok {
		count := int(odataCount)
		assert.Equal(t, 3, count, "Expected @odata.count to be 3")
	}
}

// TestE2E_ODataOrderBy tests the $orderby query option
func TestE2E_ODataOrderBy(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// Create 3 users with specific displayNames via SDK
	userNames := []string{"Charlie OrderBy", "Bob OrderBy", "Alice OrderBy"}

	for _, name := range userNames {
		requestBody := models.NewUser()
		requestBody.SetDisplayName(&name)
		upn := strings.ToLower(strings.ReplaceAll(name, " ", "")) + "@saldeti.local"
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

		_, err := tss.SDKClient.Users().Post(ctx, requestBody, nil)
		require.NoError(t, err, "Failed to create user %s", name)
	}

	// GET /users?$orderby=displayName via SDK
	filter := "contains(displayName,'OrderBy')"
	orderby := []string{"displayName"}
	result, err := tss.SDKClient.Users().Get(ctx, &users.UsersRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
			Filter:  &filter,
			Orderby: orderby,
		},
	})
	require.NoError(t, err, "Failed to get users with orderby")

	usersList := result.GetValue()
	assert.Len(t, usersList, 3, "Expected 3 users")

	// Verify alphabetical order
	expectedNames := []string{"Alice OrderBy", "Bob OrderBy", "Charlie OrderBy"}
	for i, expectedName := range expectedNames {
		displayName := usersList[i].GetDisplayName()
		assert.NotNil(t, displayName, "User at position %d missing displayName", i)
		assert.Equal(t, expectedName, *displayName, "Expected displayName '%s' at position %d", expectedName, i)
	}
}

// TestE2E_ODataSearch tests the $search query option
func TestE2E_ODataSearch(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// Create users with department info via SDK
	testUsers := []struct {
		name       string
		department string
	}{
		{"Alice Engineer", "Engineering"},
		{"Bob Sales", "Sales"},
		{"Charlie Marketing", "Marketing"},
	}

	for _, u := range testUsers {
		requestBody := models.NewUser()
		requestBody.SetDisplayName(&u.name)
		upn := strings.ToLower(strings.ReplaceAll(u.name, " ", "")) + "@saldeti.local"
		requestBody.SetUserPrincipalName(&upn)
		requestBody.SetMail(&upn)
		requestBody.SetDepartment(&u.department)

		accountEnabled := true
		requestBody.SetAccountEnabled(&accountEnabled)

		passwordProfile := models.NewPasswordProfile()
		password := "Test1234!"
		passwordProfile.SetPassword(&password)
		requestBody.SetPasswordProfile(passwordProfile)

		userType := "Member"
		requestBody.SetUserType(&userType)

		_, err := tss.SDKClient.Users().Post(ctx, requestBody, nil)
		require.NoError(t, err, "Failed to create user %s", u.name)
	}

	// Test 1: Search for "Ali" without quotes via SDK
	search := "Ali"
	result, err := tss.SDKClient.Users().Get(ctx, &users.UsersRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
			Search: &search,
		},
	})
	require.NoError(t, err, "Failed to search users (without quotes)")

	// Should find "Alice Engineer"
	usersList := result.GetValue()
	assert.NotEmpty(t, usersList, "Search returned 0 users, expected at least 1 (Alice Engineer)")

	// Verify that Alice is in the results
	foundAlice := false
	for _, user := range usersList {
		displayName := user.GetDisplayName()
		if displayName != nil && *displayName == "Alice Engineer" {
			foundAlice = true
			break
		}
	}
	assert.True(t, foundAlice, "Expected to find 'Alice Engineer' in search results")

	// Test 2: Search for "Ali" with quotes (OData format) via SDK
	searchWithQuotes := "\"Ali\""
	result2, err := tss.SDKClient.Users().Get(ctx, &users.UsersRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
			Search: &searchWithQuotes,
		},
	})
	require.NoError(t, err, "Failed to search users (with quotes)")

	// Should also find "Alice Engineer" when using quotes
	usersList2 := result2.GetValue()
	assert.NotEmpty(t, usersList2, "Search with quotes returned 0 users, expected at least 1 (Alice Engineer)")

	// Verify that Alice is in the results
	foundAlice2 := false
	for _, user := range usersList2 {
		displayName := user.GetDisplayName()
		if displayName != nil && *displayName == "Alice Engineer" {
			foundAlice2 = true
			break
		}
	}
	assert.True(t, foundAlice2, "Expected to find 'Alice Engineer' in search results with quotes")
}
