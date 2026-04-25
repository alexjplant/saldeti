package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

func TestCreateGroup(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Group.ReadWrite.All"}, []string{"Group"}, time.Hour, "", "")
	require.NoError(t, err)

	// Create group
	groupJSON := `{
		"displayName": "Test Group",
		"description": "Test Description",
		"mailNickname": "testgroup",
		"mail": "testgroup@example.com",
		"securityEnabled": true,
		"mailEnabled": false,
		"groupTypes": ["Unified"],
		"visibility": "Public"
	}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/groups", strings.NewReader(groupJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Location"), "/v1.0/groups/")

	var groupResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &groupResp)
	require.NoError(t, err)

	assert.Equal(t, "https://graph.microsoft.com/v1.0/$metadata#groups/$entity", groupResp["@odata.context"])
	assert.Equal(t, "Test Group", groupResp["displayName"])
	assert.Equal(t, "Test Description", groupResp["description"])
	assert.Equal(t, "testgroup", groupResp["mailNickname"])
	assert.Equal(t, "testgroup@example.com", groupResp["mail"])
	assert.Equal(t, true, groupResp["securityEnabled"])
	assert.Equal(t, false, groupResp["mailEnabled"])
	assert.Contains(t, groupResp, "id")
	assert.NotEmpty(t, groupResp["id"])
}

func TestCreateGroupMissingDisplayName(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Group.ReadWrite.All"}, []string{"Group"}, time.Hour, "", "")
	require.NoError(t, err)

	// Try to create group without displayName
	groupJSON := `{
		"description": "Test Description",
		"securityEnabled": true
	}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/groups", strings.NewReader(groupJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetGroup(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	securityEnabled := true
	mailEnabled := false
	group := model.Group{
		DisplayName:     "Test Group",
		Description:     "Test Description",
		MailNickname:    "testgroup",
		Mail:            "testgroup@example.com",
		SecurityEnabled: &securityEnabled,
		MailEnabled:     &mailEnabled,
		Visibility:      "Public",
	}
	createdGroup, err := store.CreateGroup(ctx, group)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Group.Read.All"}, []string{"Group"}, time.Hour, "", "")
	require.NoError(t, err)

	// Get group by ID
	req, err := http.NewRequest("GET", server.URL+"/v1.0/groups/"+createdGroup.ID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var groupResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &groupResp)
	require.NoError(t, err)

	assert.Equal(t, "https://graph.microsoft.com/v1.0/$metadata#groups/$entity", groupResp["@odata.context"])
	assert.Equal(t, createdGroup.ID, groupResp["id"])
	assert.Equal(t, "Test Group", groupResp["displayName"])
	assert.Equal(t, "Test Description", groupResp["description"])
	assert.Equal(t, true, groupResp["securityEnabled"])
}

func TestGetGroupNotFound(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Group.Read.All"}, []string{"Group"}, time.Hour, "", "")
	require.NoError(t, err)

	// Get non-existent group
	req, err := http.NewRequest("GET", server.URL+"/v1.0/groups/nonexistent", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestListGroups(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create 3 test groups
	for i := 1; i <= 3; i++ {
		securityEnabled := true
		mailEnabled := false
		group := model.Group{
			DisplayName:     fmt.Sprintf("Group %d", i),
			Description:     fmt.Sprintf("Description %d", i),
			MailNickname:    fmt.Sprintf("group%d", i),
			Mail:            fmt.Sprintf("group%d@example.com", i),
			SecurityEnabled: &securityEnabled,
			MailEnabled:     &mailEnabled,
		}
		_, err := store.CreateGroup(ctx, group)
		require.NoError(t, err)
	}

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Group.Read.All"}, []string{"Group"}, time.Hour, "", "")
	require.NoError(t, err)

	// Test listing all groups
	req, err := http.NewRequest("GET", server.URL+"/v1.0/groups", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	assert.Equal(t, "https://graph.microsoft.com/v1.0/$metadata#groups", listResp["@odata.context"])
	groups := listResp["value"].([]interface{})
	assert.Len(t, groups, 3)
}

func TestListGroupsFilter(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create groups with different securityEnabled status
	securityEnabledTrue := true
	securityEnabledFalse := false
	mailEnabled := false

	groups := []model.Group{
		{
			DisplayName:     "Security Group 1",
			MailNickname:    "secgroup1",
			SecurityEnabled: &securityEnabledTrue,
			MailEnabled:     &mailEnabled,
		},
		{
			DisplayName:     "Mail Group",
			MailNickname:    "mailgroup",
			SecurityEnabled: &securityEnabledFalse,
			MailEnabled:     &mailEnabled,
		},
		{
			DisplayName:     "Security Group 2",
			MailNickname:    "secgroup2",
			SecurityEnabled: &securityEnabledTrue,
			MailEnabled:     &mailEnabled,
		},
	}

	for _, group := range groups {
		_, err := store.CreateGroup(ctx, group)
		require.NoError(t, err)
	}

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Group.Read.All"}, []string{"Group"}, time.Hour, "", "")
	require.NoError(t, err)

	// Filter by securityEnabled eq true
	req, err := http.NewRequest("GET", server.URL+"/v1.0/groups?$filter=securityEnabled%20eq%20true", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	groupsResp := listResp["value"].([]interface{})
	assert.Len(t, groupsResp, 2) // Should have 2 security groups

	// Verify only security groups returned
	for _, g := range groupsResp {
		groupMap := g.(map[string]interface{})
		assert.Equal(t, true, groupMap["securityEnabled"])
		assert.Contains(t, groupMap["displayName"], "Security Group")
	}
}

func TestUpdateGroup(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	securityEnabled := true
	mailEnabled := false
	group := model.Group{
		DisplayName:     "Original Group",
		Description:     "Original Description",
		MailNickname:    "original",
		Mail:            "original@example.com",
		SecurityEnabled: &securityEnabled,
		MailEnabled:     &mailEnabled,
	}
	createdGroup, err := store.CreateGroup(ctx, group)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Group.ReadWrite.All"}, []string{"Group"}, time.Hour, "", "")
	require.NoError(t, err)

	// Update group
	patchJSON := `{
		"displayName": "Updated Group",
		"description": "Updated Description",
		"visibility": "Private"
	}`

	req, err := http.NewRequest("PATCH", server.URL+"/v1.0/groups/"+createdGroup.ID, strings.NewReader(patchJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var groupResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &groupResp)
	require.NoError(t, err)

	assert.Equal(t, "Updated Group", groupResp["displayName"])
	assert.Equal(t, "Updated Description", groupResp["description"])
	assert.Equal(t, "Private", groupResp["visibility"])
	assert.Equal(t, "original@example.com", groupResp["mail"]) // Should not change
	assert.Equal(t, true, groupResp["securityEnabled"])        // Should not change
}

func TestDeleteGroup(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	securityEnabled := true
	mailEnabled := false
	group := model.Group{
		DisplayName:     "Group to Delete",
		MailNickname:    "deletegroup",
		Mail:            "delete@example.com",
		SecurityEnabled: &securityEnabled,
		MailEnabled:     &mailEnabled,
	}
	createdGroup, err := store.CreateGroup(ctx, group)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Group.ReadWrite.All"}, []string{"Group"}, time.Hour, "", "")
	require.NoError(t, err)

	// Delete group
	req, err := http.NewRequest("DELETE", server.URL+"/v1.0/groups/"+createdGroup.ID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify group is deleted
	req, err = http.NewRequest("GET", server.URL+"/v1.0/groups/"+createdGroup.ID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCreateGroupWithMembers(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// First create a user to add as member
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

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Group.ReadWrite.All", "User.ReadWrite.All"}, []string{"Group", "User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Create group with inline members using members@odata.bind
	groupJSON := fmt.Sprintf(`{
		"displayName": "Group With Members",
		"mailNickname": "withmembers",
		"securityEnabled": true,
		"mailEnabled": false,
		"members@odata.bind": [
			"https://graph.microsoft.com/v1.0/directoryObjects/%s"
		]
	}`, createdUser.ID)

	req, err := http.NewRequest("POST", server.URL+"/v1.0/groups", strings.NewReader(groupJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var groupResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &groupResp)
	require.NoError(t, err)

	groupID := groupResp["id"].(string)

	// Verify members were added
	req, err = http.NewRequest("GET", server.URL+"/v1.0/groups/"+groupID+"/members", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var membersResp map[string]interface{}
	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &membersResp)
	require.NoError(t, err)

	members := membersResp["value"].([]interface{})
	assert.Len(t, members, 1)
}

func TestAddMember(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create a group
	securityEnabled := true
	mailEnabled := false
	group := model.Group{
		DisplayName:     "Test Group",
		MailNickname:    "testgroup",
		SecurityEnabled: &securityEnabled,
		MailEnabled:     &mailEnabled,
	}
	createdGroup, err := store.CreateGroup(ctx, group)
	require.NoError(t, err)

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

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Group.ReadWrite.All", "User.ReadWrite.All"}, []string{"Group", "User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Add member to group
	memberJSON := fmt.Sprintf(`{
		"@odata.id": "https://graph.microsoft.com/v1.0/directoryObjects/%s"
	}`, createdUser.ID)

	req, err := http.NewRequest("POST", server.URL+"/v1.0/groups/"+createdGroup.ID+"/members/$ref", strings.NewReader(memberJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify member was added
	req, err = http.NewRequest("GET", server.URL+"/v1.0/groups/"+createdGroup.ID+"/members", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var membersResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &membersResp)
	require.NoError(t, err)

	members := membersResp["value"].([]interface{})
	assert.Len(t, members, 1)

	member := members[0].(map[string]interface{})
	assert.Equal(t, createdUser.ID, member["id"])
	assert.Equal(t, "#microsoft.graph.user", member["@odata.type"])
}

func TestRemoveMember(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create a group
	securityEnabled := true
	mailEnabled := false
	group := model.Group{
		DisplayName:     "Test Group",
		MailNickname:    "testgroup",
		SecurityEnabled: &securityEnabled,
		MailEnabled:     &mailEnabled,
	}
	createdGroup, err := store.CreateGroup(ctx, group)
	require.NoError(t, err)

	// Create a user and add as member
	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "testuser@example.com",
		Mail:              "testuser@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	// Add member directly via store
	err = store.AddMember(ctx, createdGroup.ID, createdUser.ID, "user")
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Group.ReadWrite.All", "User.ReadWrite.All"}, []string{"Group", "User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Remove member from group
	req, err := http.NewRequest("DELETE", server.URL+"/v1.0/groups/"+createdGroup.ID+"/members/"+createdUser.ID+"/$ref", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify member was removed
	req, err = http.NewRequest("GET", server.URL+"/v1.0/groups/"+createdGroup.ID+"/members", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var membersResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &membersResp)
	require.NoError(t, err)

	members := membersResp["value"].([]interface{})
	assert.Len(t, members, 0)
}

func TestListMembers(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create a group
	securityEnabled := true
	mailEnabled := false
	group := model.Group{
		DisplayName:     "Test Group",
		MailNickname:    "testgroup",
		SecurityEnabled: &securityEnabled,
		MailEnabled:     &mailEnabled,
	}
	createdGroup, err := store.CreateGroup(ctx, group)
	require.NoError(t, err)

	// Create 3 users and add as members
	for i := 1; i <= 3; i++ {
		accountEnabled := true
		user := model.User{
			DisplayName:       fmt.Sprintf("User %d", i),
			UserPrincipalName: fmt.Sprintf("user%d@example.com", i),
			Mail:              fmt.Sprintf("user%d@example.com", i),
			AccountEnabled:    &accountEnabled,
		}
		createdUser, err := store.CreateUser(ctx, user)
		require.NoError(t, err)

		err = store.AddMember(ctx, createdGroup.ID, createdUser.ID, "user")
		require.NoError(t, err)
	}

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Group.Read.All", "User.Read.All"}, []string{"Group", "User"}, time.Hour, "", "")
	require.NoError(t, err)

	// List members
	req, err := http.NewRequest("GET", server.URL+"/v1.0/groups/"+createdGroup.ID+"/members", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var membersResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &membersResp)
	require.NoError(t, err)

	members := membersResp["value"].([]interface{})
	assert.Len(t, members, 3)

	// Verify each member has @odata.type
	for _, m := range members {
		member := m.(map[string]interface{})
		assert.Contains(t, member, "@odata.type")
		assert.Equal(t, "#microsoft.graph.user", member["@odata.type"])
		assert.Contains(t, member, "id")
		assert.Contains(t, member, "displayName")
	}
}

func TestTransitiveMembership(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create nested groups: A -> B -> C
	securityEnabled := true
	mailEnabled := false

	// Create group C
	groupC := model.Group{
		DisplayName:     "Group C",
		MailNickname:    "groupc",
		SecurityEnabled: &securityEnabled,
		MailEnabled:     &mailEnabled,
	}
	createdGroupC, err := store.CreateGroup(ctx, groupC)
	require.NoError(t, err)

	// Create group B with C as member
	groupB := model.Group{
		DisplayName:     "Group B",
		MailNickname:    "groupb",
		SecurityEnabled: &securityEnabled,
		MailEnabled:     &mailEnabled,
	}
	createdGroupB, err := store.CreateGroup(ctx, groupB)
	require.NoError(t, err)

	// Create group A
	groupA := model.Group{
		DisplayName:     "Group A",
		MailNickname:    "groupa",
		SecurityEnabled: &securityEnabled,
		MailEnabled:     &mailEnabled,
	}
	createdGroupA, err := store.CreateGroup(ctx, groupA)
	require.NoError(t, err)

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

	// Add C to B
	err = store.AddMember(ctx, createdGroupB.ID, createdGroupC.ID, "group")
	require.NoError(t, err)

	// Add B to A
	err = store.AddMember(ctx, createdGroupA.ID, createdGroupB.ID, "group")
	require.NoError(t, err)

	// Add user to C
	err = store.AddMember(ctx, createdGroupC.ID, createdUser.ID, "user")
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Group.Read.All", "User.Read.All"}, []string{"Group", "User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Get transitive members of A (should include user in C)
	req, err := http.NewRequest("GET", server.URL+"/v1.0/groups/"+createdGroupA.ID+"/transitiveMembers", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var transitiveResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &transitiveResp)
	require.NoError(t, err)

	members := transitiveResp["value"].([]interface{})
	
	// Should include: user, group C, group B
	// Note: transitiveMembers includes all nested objects
	foundUser := false
	for _, m := range members {
		member := m.(map[string]interface{})
		if member["id"] == createdUser.ID {
			foundUser = true
			break
		}
	}
	assert.True(t, foundUser, "User should be in transitive members of Group A")
}

func TestAddOwner(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create a group
	securityEnabled := true
	mailEnabled := false
	group := model.Group{
		DisplayName:     "Test Group",
		MailNickname:    "testgroup",
		SecurityEnabled: &securityEnabled,
		MailEnabled:     &mailEnabled,
	}
	createdGroup, err := store.CreateGroup(ctx, group)
	require.NoError(t, err)

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

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Group.ReadWrite.All", "User.ReadWrite.All"}, []string{"Group", "User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Add owner to group
	ownerJSON := fmt.Sprintf(`{
		"@odata.id": "https://graph.microsoft.com/v1.0/directoryObjects/%s"
	}`, createdUser.ID)

	req, err := http.NewRequest("POST", server.URL+"/v1.0/groups/"+createdGroup.ID+"/owners/$ref", strings.NewReader(ownerJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify owner was added
	req, err = http.NewRequest("GET", server.URL+"/v1.0/groups/"+createdGroup.ID+"/owners", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var ownersResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &ownersResp)
	require.NoError(t, err)

	owners := ownersResp["value"].([]interface{})
	assert.Len(t, owners, 1)

	owner := owners[0].(map[string]interface{})
	assert.Equal(t, createdUser.ID, owner["id"])
	assert.Equal(t, "#microsoft.graph.user", owner["@odata.type"])
}

func TestRemoveOwner(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create a group
	securityEnabled := true
	mailEnabled := false
	group := model.Group{
		DisplayName:     "Test Group",
		MailNickname:    "testgroup",
		SecurityEnabled: &securityEnabled,
		MailEnabled:     &mailEnabled,
	}
	createdGroup, err := store.CreateGroup(ctx, group)
	require.NoError(t, err)

	// Create a user and add as owner
	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "testuser@example.com",
		Mail:              "testuser@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	// Add owner directly via store
	err = store.AddOwner(ctx, createdGroup.ID, createdUser.ID, "user")
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Group.ReadWrite.All", "User.ReadWrite.All"}, []string{"Group", "User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Remove owner from group
	req, err := http.NewRequest("DELETE", server.URL+"/v1.0/groups/"+createdGroup.ID+"/owners/"+createdUser.ID+"/$ref", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify owner was removed
	req, err = http.NewRequest("GET", server.URL+"/v1.0/groups/"+createdGroup.ID+"/owners", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var ownersResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &ownersResp)
	require.NoError(t, err)

	owners := ownersResp["value"].([]interface{})
	assert.Len(t, owners, 0)
}

func TestCheckMemberGroups(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create a group to check membership for
	securityEnabled := true
	mailEnabled := false
	groupToCheck := model.Group{
		DisplayName:     "Group to Check",
		MailNickname:    "checkgroup",
		SecurityEnabled: &securityEnabled,
		MailEnabled:     &mailEnabled,
	}
	createdGroupToCheck, err := store.CreateGroup(ctx, groupToCheck)
	require.NoError(t, err)

	// Create 3 other groups
	var groupIDs []string
	for i := 1; i <= 3; i++ {
		group := model.Group{
			DisplayName:     fmt.Sprintf("Group %d", i),
			MailNickname:    fmt.Sprintf("group%d", i),
			SecurityEnabled: &securityEnabled,
			MailEnabled:     &mailEnabled,
		}
		createdGroup, err := store.CreateGroup(ctx, group)
		require.NoError(t, err)
		groupIDs = append(groupIDs, createdGroup.ID)
	}

	// Add group to check as member of G1 and G2
	err = store.AddMember(ctx, groupIDs[0], createdGroupToCheck.ID, "group")
	require.NoError(t, err)
	err = store.AddMember(ctx, groupIDs[1], createdGroupToCheck.ID, "group")
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"GroupMember.Read.All"}, []string{"Group"}, time.Hour, "", "")
	require.NoError(t, err)

	// Check member groups - group is in G1 and G2, not in G3
	checkJSON := fmt.Sprintf(`{
		"groupIds": ["%s", "%s", "%s"]
	}`, groupIDs[0], groupIDs[1], groupIDs[2])

	req, err := http.NewRequest("POST", server.URL+"/v1.0/groups/"+createdGroupToCheck.ID+"/checkMemberGroups", strings.NewReader(checkJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var checkResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &checkResp)
	require.NoError(t, err)

	value := checkResp["value"].([]interface{})
	assert.Len(t, value, 2) // Should return G1 and G2

	// Verify G1 and G2 are in response, G3 is not
	foundG1 := false
	foundG2 := false
	for _, v := range value {
		groupID := v.(string)
		if groupID == groupIDs[0] {
			foundG1 = true
		}
		if groupID == groupIDs[1] {
			foundG2 = true
		}
	}
	assert.True(t, foundG1, "G1 should be in checkMemberGroups response")
	assert.True(t, foundG2, "G2 should be in checkMemberGroups response")
}

func TestGetMemberGroups(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create nested groups: A -> B -> C
	securityEnabled := true
	mailEnabled := false

	// Create group C
	groupC := model.Group{
		DisplayName:     "Group C",
		MailNickname:    "groupc",
		SecurityEnabled: &securityEnabled,
		MailEnabled:     &mailEnabled,
	}
	createdGroupC, err := store.CreateGroup(ctx, groupC)
	require.NoError(t, err)

	// Create group B with C as member
	groupB := model.Group{
		DisplayName:     "Group B",
		MailNickname:    "groupb",
		SecurityEnabled: &securityEnabled,
		MailEnabled:     &mailEnabled,
	}
	createdGroupB, err := store.CreateGroup(ctx, groupB)
	require.NoError(t, err)

	// Create group A with B as member
	groupA := model.Group{
		DisplayName:     "Group A",
		MailNickname:    "groupa",
		SecurityEnabled: &securityEnabled,
		MailEnabled:     &mailEnabled,
	}
	createdGroupA, err := store.CreateGroup(ctx, groupA)
	require.NoError(t, err)

	// Add C to B
	err = store.AddMember(ctx, createdGroupB.ID, createdGroupC.ID, "group")
	require.NoError(t, err)

	// Add B to A
	err = store.AddMember(ctx, createdGroupA.ID, createdGroupB.ID, "group")
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"GroupMember.Read.All"}, []string{"Group"}, time.Hour, "", "")
	require.NoError(t, err)

	// Get member groups for group C (should include both A and B transitively)
	getJSON := `{
		"securityEnabledOnly": false
	}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/groups/"+createdGroupC.ID+"/getMemberGroups", strings.NewReader(getJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var getResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &getResp)
	require.NoError(t, err)

	value := getResp["value"].([]interface{})
	// Should include both A and B (transitive membership)
	assert.Len(t, value, 2)

	// Verify both group IDs are in response
	foundA := false
	foundB := false
	for _, v := range value {
		groupID := v.(string)
		if groupID == createdGroupA.ID {
			foundA = true
		}
		if groupID == createdGroupB.ID {
			foundB = true
		}
	}
	assert.True(t, foundA, "Group A should be in getMemberGroups response")
	assert.True(t, foundB, "Group B should be in getMemberGroups response")
}

func TestGroupMemberOf(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create groups A and B
	securityEnabled := true
	mailEnabled := false

	// Create group B
	groupB := model.Group{
		DisplayName:     "Group B",
		MailNickname:    "groupb",
		SecurityEnabled: &securityEnabled,
		MailEnabled:     &mailEnabled,
	}
	createdGroupB, err := store.CreateGroup(ctx, groupB)
	require.NoError(t, err)

	// Create group A
	groupA := model.Group{
		DisplayName:     "Group A",
		MailNickname:    "groupa",
		SecurityEnabled: &securityEnabled,
		MailEnabled:     &mailEnabled,
	}
	createdGroupA, err := store.CreateGroup(ctx, groupA)
	require.NoError(t, err)

	// Add B to A as member
	err = store.AddMember(ctx, createdGroupA.ID, createdGroupB.ID, "group")
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"Group.Read.All"}, []string{"Group"}, time.Hour, "", "")
	require.NoError(t, err)

	// Get memberOf for group B (should return A)
	req, err := http.NewRequest("GET", server.URL+"/v1.0/groups/"+createdGroupB.ID+"/memberOf", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var memberOfResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &memberOfResp)
	require.NoError(t, err)

	memberOf := memberOfResp["value"].([]interface{})
	assert.Len(t, memberOf, 1)

	// Verify group A is returned
	member := memberOf[0].(map[string]interface{})
	assert.Equal(t, createdGroupA.ID, member["id"])
	assert.Equal(t, "#microsoft.graph.group", member["@odata.type"])
	assert.Equal(t, "Group A", member["displayName"])
}