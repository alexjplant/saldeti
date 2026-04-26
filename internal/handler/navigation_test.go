package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/saldeti/saldeti/internal/auth"
	"github.com/saldeti/saldeti/internal/model"
	"github.com/saldeti/saldeti/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListUserMemberOf(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create a user
	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "testuser@example.com",
		Mail:              "testuser@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	// Create 2 groups
	var groupIDs []string
	for i := 1; i <= 2; i++ {
		securityEnabled := true
		group := model.Group{
			DisplayName:     fmt.Sprintf("Group %d", i),
			MailNickname:    fmt.Sprintf("group%d", i),
			SecurityEnabled: &securityEnabled,
		}
		createdGroup, err := store.CreateGroup(ctx, group)
		require.NoError(t, err)
		groupIDs = append(groupIDs, createdGroup.ID)

		// Add user to group
		err = store.AddMember(ctx, createdGroup.ID, createdUser.ID, "user")
		require.NoError(t, err)
	}

	server := httptest.NewServer(router)
	defer server.Close()

	// Mint a token
	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Test listing user memberOf
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users/"+createdUser.ID+"/memberOf", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	value, ok := result["value"].([]interface{})
	require.True(t, ok)
	assert.Len(t, value, 2)

	// Verify each group is returned as a directory object with correct @odata.type
	for _, item := range value {
		obj, ok := item.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "#microsoft.graph.group", obj["@odata.type"])
		assert.Contains(t, []string{"Group 1", "Group 2"}, obj["displayName"])
	}
}

func TestListUserTransitiveMemberOf(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create a user
	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "testuser@example.com",
		Mail:              "testuser@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	// Create nested groups: A -> B -> C, add user to C
	var groups []model.Group
	for i := 1; i <= 3; i++ {
		securityEnabled := true
		group := model.Group{
			DisplayName:     fmt.Sprintf("Group %c", 'A'+i-1),
			MailNickname:    fmt.Sprintf("group%c", 'a'+i-1),
			SecurityEnabled: &securityEnabled,
		}
		createdGroup, err := store.CreateGroup(ctx, group)
		require.NoError(t, err)
		groups = append(groups, createdGroup)
	}

	// Create hierarchy: A contains B, B contains C, user is in C
	err = store.AddMember(ctx, groups[0].ID, groups[1].ID, "group") // A contains B
	require.NoError(t, err)
	err = store.AddMember(ctx, groups[1].ID, groups[2].ID, "group") // B contains C
	require.NoError(t, err)
	err = store.AddMember(ctx, groups[2].ID, createdUser.ID, "user") // C contains user
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	// Mint a token
	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Test listing user transitiveMemberOf
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users/"+createdUser.ID+"/transitiveMemberOf", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	value, ok := result["value"].([]interface{})
	require.True(t, ok)
	assert.Len(t, value, 3) // User should be transitive member of A, B, and C

	// Verify all groups are returned
	groupNames := make([]string, 0, 3)
	for _, item := range value {
		obj, ok := item.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "#microsoft.graph.group", obj["@odata.type"])
		groupNames = append(groupNames, obj["displayName"].(string))
	}

	assert.Contains(t, groupNames, "Group A")
	assert.Contains(t, groupNames, "Group B")
	assert.Contains(t, groupNames, "Group C")
}

func TestGetManager(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create manager user
	accountEnabled := true
	manager := model.User{
		DisplayName:       "Manager User",
		UserPrincipalName: "manager@example.com",
		Mail:              "manager@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdManager, err := store.CreateUser(ctx, manager)
	require.NoError(t, err)

	// Create regular user
	user := model.User{
		DisplayName:       "Regular User",
		UserPrincipalName: "regular@example.com",
		Mail:              "regular@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	// Set manager
	err = store.SetManager(ctx, createdUser.ID, createdManager.ID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	// Mint a token
	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Test getting manager
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users/"+createdUser.ID+"/manager", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "#microsoft.graph.user", result["@odata.type"])
	assert.Equal(t, createdManager.ID, result["id"])
	assert.Equal(t, "Manager User", result["displayName"])
}

func TestGetManagerNotSet(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create user
	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "testuser@example.com",
		Mail:              "testuser@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	// Mint a token
	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Test getting manager when not set
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users/"+createdUser.ID+"/manager", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "error")
	errorObj := result["error"].(map[string]interface{})
	assert.Equal(t, "Request_ResourceNotFound", errorObj["code"])
}

func TestSetManager(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create manager user
	accountEnabled := true
	manager := model.User{
		DisplayName:       "Manager User",
		UserPrincipalName: "manager@example.com",
		Mail:              "manager@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdManager, err := store.CreateUser(ctx, manager)
	require.NoError(t, err)

	// Create regular user
	user := model.User{
		DisplayName:       "Regular User",
		UserPrincipalName: "regular@example.com",
		Mail:              "regular@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	// Mint a token
	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Test setting manager
	requestBody := fmt.Sprintf(`{"@odata.id": "https://graph.microsoft.com/v1.0/users/%s"}`, createdManager.ID)
	req, err := http.NewRequest("PUT", server.URL+"/v1.0/users/"+createdUser.ID+"/manager/$ref", strings.NewReader(requestBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify manager was set by getting it
	req2, err := http.NewRequest("GET", server.URL+"/v1.0/users/"+createdUser.ID+"/manager", nil)
	require.NoError(t, err)
	req2.Header.Set("Authorization", "Bearer "+token)

	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp2.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, createdManager.ID, result["id"])
	assert.Equal(t, "Manager User", result["displayName"])
}

func TestRemoveManager(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create manager user
	accountEnabled := true
	manager := model.User{
		DisplayName:       "Manager User",
		UserPrincipalName: "manager@example.com",
		Mail:              "manager@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdManager, err := store.CreateUser(ctx, manager)
	require.NoError(t, err)

	// Create regular user
	user := model.User{
		DisplayName:       "Regular User",
		UserPrincipalName: "regular@example.com",
		Mail:              "regular@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	// Set manager
	err = store.SetManager(ctx, createdUser.ID, createdManager.ID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	// Mint a token
	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Test removing manager
	req, err := http.NewRequest("DELETE", server.URL+"/v1.0/users/"+createdUser.ID+"/manager/$ref", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify manager was removed by trying to get it
	req2, err := http.NewRequest("GET", server.URL+"/v1.0/users/"+createdUser.ID+"/manager", nil)
	require.NoError(t, err)
	req2.Header.Set("Authorization", "Bearer "+token)

	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp2.StatusCode)
}

func TestDirectReports(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create manager user
	accountEnabled := true
	manager := model.User{
		DisplayName:       "Manager User",
		UserPrincipalName: "manager@example.com",
		Mail:              "manager@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdManager, err := store.CreateUser(ctx, manager)
	require.NoError(t, err)

	// Create 3 direct reports
	var reportIDs []string
	for i := 1; i <= 3; i++ {
		user := model.User{
			DisplayName:       fmt.Sprintf("Report %d", i),
			UserPrincipalName: fmt.Sprintf("report%d@example.com", i),
			Mail:              fmt.Sprintf("report%d@example.com", i),
			AccountEnabled:    &accountEnabled,
		}
		createdUser, err := store.CreateUser(ctx, user)
		require.NoError(t, err)
		reportIDs = append(reportIDs, createdUser.ID)

		// Set manager
		err = store.SetManager(ctx, createdUser.ID, createdManager.ID)
		require.NoError(t, err)
	}

	server := httptest.NewServer(router)
	defer server.Close()

	// Mint a token
	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Test listing direct reports
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users/"+createdManager.ID+"/directReports", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	value, ok := result["value"].([]interface{})
	require.True(t, ok)
	assert.Len(t, value, 3)

	// Verify all direct reports are returned
	reportNames := make([]string, 0, 3)
	for _, item := range value {
		obj, ok := item.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "#microsoft.graph.user", obj["@odata.type"])
		reportNames = append(reportNames, obj["displayName"].(string))
	}

	assert.Contains(t, reportNames, "Report 1")
	assert.Contains(t, reportNames, "Report 2")
	assert.Contains(t, reportNames, "Report 3")
}

func TestGetByIds(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create 2 users
	var userIDs []string
	for i := 1; i <= 2; i++ {
		accountEnabled := true
		user := model.User{
			DisplayName:       fmt.Sprintf("User %d", i),
			UserPrincipalName: fmt.Sprintf("user%d@example.com", i),
			Mail:              fmt.Sprintf("user%d@example.com", i),
			AccountEnabled:    &accountEnabled,
		}
		createdUser, err := store.CreateUser(ctx, user)
		require.NoError(t, err)
		userIDs = append(userIDs, createdUser.ID)
	}

	// Create 1 group
	securityEnabled := true
	group := model.Group{
		DisplayName:     "Test Group",
		MailNickname:    "testgroup",
		SecurityEnabled: &securityEnabled,
	}
	createdGroup, err := store.CreateGroup(ctx, group)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	// Mint a token
	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Directory.Read.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// Test getByIds with all IDs
	requestBody := fmt.Sprintf(`{"ids": ["%s", "%s", "%s"]}`, userIDs[0], userIDs[1], createdGroup.ID)
	req, err := http.NewRequest("POST", server.URL+"/v1.0/directoryObjects/getByIds", strings.NewReader(requestBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	value, ok := result["value"].([]interface{})
	require.True(t, ok)
	assert.Len(t, value, 3)

	// Verify all objects are returned
	userCount := 0
	groupCount := 0
	for _, item := range value {
		obj, ok := item.(map[string]interface{})
		require.True(t, ok)
		if obj["@odata.type"] == "#microsoft.graph.user" {
			userCount++
		} else if obj["@odata.type"] == "#microsoft.graph.group" {
			groupCount++
			assert.Equal(t, "Test Group", obj["displayName"])
		}
	}
	assert.Equal(t, 2, userCount)
	assert.Equal(t, 1, groupCount)
}

func TestGetByIdsWithTypeFilter(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create 2 users
	var userIDs []string
	for i := 1; i <= 2; i++ {
		accountEnabled := true
		user := model.User{
			DisplayName:       fmt.Sprintf("User %d", i),
			UserPrincipalName: fmt.Sprintf("user%d@example.com", i),
			Mail:              fmt.Sprintf("user%d@example.com", i),
			AccountEnabled:    &accountEnabled,
		}
		createdUser, err := store.CreateUser(ctx, user)
		require.NoError(t, err)
		userIDs = append(userIDs, createdUser.ID)
	}

	// Create 1 group
	securityEnabled := true
	group := model.Group{
		DisplayName:     "Test Group",
		MailNickname:    "testgroup",
		SecurityEnabled: &securityEnabled,
	}
	createdGroup, err := store.CreateGroup(ctx, group)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	// Mint a token
	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Directory.Read.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// Test getByIds with type filter (only users)
	requestBody := fmt.Sprintf(`{"ids": ["%s", "%s", "%s"], "types": ["user"]}`, userIDs[0], userIDs[1], createdGroup.ID)
	req, err := http.NewRequest("POST", server.URL+"/v1.0/directoryObjects/getByIds", strings.NewReader(requestBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	value, ok := result["value"].([]interface{})
	require.True(t, ok)
	assert.Len(t, value, 2) // Only users should be returned

	// Verify only users are returned
	for _, item := range value {
		obj, ok := item.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "#microsoft.graph.user", obj["@odata.type"])
	}
}

func TestUserCheckMemberGroups(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create a user
	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "testuser@example.com",
		Mail:              "testuser@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	// Create 3 groups
	var groupIDs []string
	for i := 1; i <= 3; i++ {
		securityEnabled := true
		group := model.Group{
			DisplayName:     fmt.Sprintf("Group %d", i),
			MailNickname:    fmt.Sprintf("group%d", i),
			SecurityEnabled: &securityEnabled,
		}
		createdGroup, err := store.CreateGroup(ctx, group)
		require.NoError(t, err)
		groupIDs = append(groupIDs, createdGroup.ID)

		// Add user to groups 1 and 2 only
		if i <= 2 {
			err = store.AddMember(ctx, createdGroup.ID, createdUser.ID, "user")
			require.NoError(t, err)
		}
	}

	server := httptest.NewServer(router)
	defer server.Close()

	// Mint a token
	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Test checkMemberGroups
	requestBody := fmt.Sprintf(`{"groupIds": ["%s", "%s", "%s"]}`, groupIDs[0], groupIDs[1], groupIDs[2])
	req, err := http.NewRequest("POST", server.URL+"/v1.0/users/"+createdUser.ID+"/checkMemberGroups", strings.NewReader(requestBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	value, ok := result["value"].([]interface{})
	require.True(t, ok)
	assert.Len(t, value, 2) // User is only member of groups 1 and 2

	// Verify correct groups are returned
	returnedGroupIDs := make([]string, 0, 2)
	for _, item := range value {
		returnedGroupIDs = append(returnedGroupIDs, item.(string))
	}

	assert.Contains(t, returnedGroupIDs, groupIDs[0])
	assert.Contains(t, returnedGroupIDs, groupIDs[1])
	assert.NotContains(t, returnedGroupIDs, groupIDs[2])
}

func TestUserGetMemberGroups(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create a user
	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "testuser@example.com",
		Mail:              "testuser@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	// Create nested groups: A -> B, user is in B
	var groups []model.Group
	for i := 1; i <= 2; i++ {
		securityEnabled := true
		group := model.Group{
			DisplayName:     fmt.Sprintf("Group %c", 'A'+i-1),
			MailNickname:    fmt.Sprintf("group%c", 'a'+i-1),
			SecurityEnabled: &securityEnabled,
		}
		createdGroup, err := store.CreateGroup(ctx, group)
		require.NoError(t, err)
		groups = append(groups, createdGroup)
	}

	// Create hierarchy: A contains B, user is in B
	err = store.AddMember(ctx, groups[0].ID, groups[1].ID, "group") // A contains B
	require.NoError(t, err)
	err = store.AddMember(ctx, groups[1].ID, createdUser.ID, "user") // B contains user
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	// Mint a token
	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Test getMemberGroups
	requestBody := `{"securityEnabledOnly": false}`
	req, err := http.NewRequest("POST", server.URL+"/v1.0/users/"+createdUser.ID+"/getMemberGroups", strings.NewReader(requestBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	value, ok := result["value"].([]interface{})
	require.True(t, ok)
	assert.Len(t, value, 2) // User should be transitive member of both A and B

	// Verify both groups are returned
	returnedGroupIDs := make([]string, 0, 2)
	for _, item := range value {
		returnedGroupIDs = append(returnedGroupIDs, item.(string))
	}

	assert.Contains(t, returnedGroupIDs, groups[0].ID) // Group A
	assert.Contains(t, returnedGroupIDs, groups[1].ID) // Group B
}

func TestUsersDelta(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create 3 test users
	for i := 1; i <= 3; i++ {
		accountEnabled := true
		user := model.User{
			DisplayName:       fmt.Sprintf("User %d", i),
			UserPrincipalName: fmt.Sprintf("user%d@example.com", i),
			Mail:              fmt.Sprintf("user%d@example.com", i),
			AccountEnabled:    &accountEnabled,
		}
		_, err := store.CreateUser(ctx, user)
		require.NoError(t, err)
	}

	server := httptest.NewServer(router)
	defer server.Close()

	// Mint a token
	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Test users delta
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users/delta", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, result, "@odata.context")
	assert.Contains(t, result, "@odata.deltaLink")
	assert.Contains(t, result, "value")

	value, ok := result["value"].([]interface{})
	require.True(t, ok)
	assert.Len(t, value, 3)

	// Verify delta link is present
	deltaLink, ok := result["@odata.deltaLink"].(string)
	require.True(t, ok)
	assert.Contains(t, deltaLink, "$deltatoken=")
}

func TestGroupsDelta(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create 3 test groups
	for i := 1; i <= 3; i++ {
		securityEnabled := true
		group := model.Group{
			DisplayName:     fmt.Sprintf("Group %d", i),
			MailNickname:    fmt.Sprintf("group%d", i),
			SecurityEnabled: &securityEnabled,
		}
		_, err := store.CreateGroup(ctx, group)
		require.NoError(t, err)
	}

	server := httptest.NewServer(router)
	defer server.Close()

	// Mint a token
	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Group.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Test groups delta
	req, err := http.NewRequest("GET", server.URL+"/v1.0/groups/delta", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, result, "@odata.context")
	assert.Contains(t, result, "@odata.deltaLink")
	assert.Contains(t, result, "value")

	value, ok := result["value"].([]interface{})
	require.True(t, ok)
	assert.Len(t, value, 3)

	// Verify delta link is present
	deltaLink, ok := result["@odata.deltaLink"].(string)
	require.True(t, ok)
	assert.Contains(t, deltaLink, "$deltatoken=")
}