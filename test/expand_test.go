//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// E2E Tests: $expand
// ============================================================================

// TestExpandUserManager tests expanding the manager navigation property on a user
func TestExpandUserManager(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	token := getToken(t, tss)

	// Create two users
	user1 := createTestUserSDK(t, tss, "User With Manager", "user.with.manager@saldeti.local")
	user1ID := *user1.GetId()

	user2 := createTestUserSDK(t, tss, "The Manager", "the.manager@saldeti.local")
	user2ID := *user2.GetId()

	// Set user2 as manager of user1 via SDK
	refBody := models.NewReferenceUpdate()
	odataID := fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, user2ID)
	refBody.SetOdataId(&odataID)
	err := tss.SDKClient.Users().ByUserId(user1ID).Manager().Ref().Put(ctx, refBody, nil)
	require.NoError(t, err, "Failed to set manager")

	// GET /users/{user1ID}?$expand=manager
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1.0/users/%s?$expand=manager", tss.BaseURL, user1ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))

	// Assert response contains "manager" field
	manager, ok := result["manager"].(map[string]interface{})
	require.True(t, ok, "Expected manager field to be present")
	require.NotNil(t, manager, "Expected manager to not be null")

	// Verify manager is user2
	managerID, ok := manager["id"].(string)
	require.True(t, ok, "Expected manager.id to be a string")
	assert.Equal(t, user2ID, managerID, "Expected manager ID to match user2")
	managerDisplayName, ok := manager["displayName"].(string)
	require.True(t, ok, "Expected manager.displayName to be a string")
	assert.Equal(t, "The Manager", managerDisplayName, "Expected manager display name to match")

	t.Logf("Successfully expanded manager for user %s", user1ID)
}

// TestExpandUserManagerNull tests expanding manager when it's null
func TestExpandUserManagerNull(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	token := getToken(t, tss)

	// Create a user without setting a manager
	user1 := createTestUserSDK(t, tss, "User No Manager", "user.no.manager@saldeti.local")
	user1ID := *user1.GetId()

	// GET /users/{user1ID}?$expand=manager
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1.0/users/%s?$expand=manager", tss.BaseURL, user1ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))

	// Assert response contains "manager" field with null
	manager, ok := result["manager"]
	require.True(t, ok, "Expected manager field to be present")
	assert.Nil(t, manager, "Expected manager to be null")

	t.Logf("Successfully verified null manager expansion for user %s", user1ID)
}

// TestExpandUserDirectReports tests expanding directReports on a user
func TestExpandUserDirectReports(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	token := getToken(t, tss)

	// Create a manager user
	managerUser := createTestUserSDK(t, tss, "Manager For Reports", "mgr.reports@saldeti.local")
	managerID := *managerUser.GetId()

	// Create a report user
	reportUser := createTestUserSDK(t, tss, "Direct Report", "direct.report@saldeti.local")
	reportID := *reportUser.GetId()

	// Set the manager for the report user
	refBody := models.NewReferenceUpdate()
	odataID := fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, managerID)
	refBody.SetOdataId(&odataID)
	err := tss.SDKClient.Users().ByUserId(reportID).Manager().Ref().Put(ctx, refBody, nil)
	require.NoError(t, err, "Failed to set manager")

	// GET /users/{managerID}?$expand=directReports
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1.0/users/%s?$expand=directReports", tss.BaseURL, managerID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))

	// Assert response contains "directReports" array
	directReports, ok := result["directReports"].([]interface{})
	require.True(t, ok, "Expected directReports field to be an array")
	assert.NotEmpty(t, directReports, "Expected at least one direct report")

	// Verify the report user is in the list
	found := false
	for _, dr := range directReports {
		drMap, ok := dr.(map[string]interface{})
		require.True(t, ok, "Expected direct report to be an object")
		drID, ok := drMap["id"].(string)
		if ok && drID == reportID {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected to find report user in directReports")

	t.Logf("Successfully expanded directReports for manager %s", managerID)
}

// TestExpandUserMemberOf tests expanding memberOf on a user
func TestExpandUserMemberOf(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	token := getToken(t, tss)

	// Create a user
	user := createTestUserSDK(t, tss, "User MemberOf", "user.memberof@saldeti.local")
	userID := *user.GetId()

	// Create a group
	group := createTestGroupSDK(t, tss, "MemberOf Test Group")
	groupID := *group.GetId()

	// Add user to group via SDK
	memberRef := models.NewReferenceCreate()
	memberRef.SetOdataId(ptrString(fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, userID)))
	err := tss.SDKClient.Groups().ByGroupId(groupID).Members().Ref().Post(ctx, memberRef, nil)
	require.NoError(t, err, "Failed to add user to group")

	// GET /users/{userID}?$expand=memberOf
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1.0/users/%s?$expand=memberOf", tss.BaseURL, userID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))

	// Assert response contains "memberOf" array
	memberOf, ok := result["memberOf"].([]interface{})
	require.True(t, ok, "Expected memberOf field to be an array")
	assert.NotEmpty(t, memberOf, "Expected at least one memberOf entry")

	// Verify the group is in the list
	found := false
	for _, mo := range memberOf {
		moMap, ok := mo.(map[string]interface{})
		require.True(t, ok, "Expected memberOf entry to be an object")
		moID, ok := moMap["id"].(string)
		if ok && moID == groupID {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected to find group in user's memberOf")

	t.Logf("Successfully expanded memberOf for user %s", userID)
}

// TestExpandGroupMembers tests expanding members on a group
func TestExpandGroupMembers(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	token := getToken(t, tss)

	// Create a group
	group := createTestGroupSDK(t, tss, "Group Members Test")
	groupID := *group.GetId()

	// Create a user
	user := createTestUserSDK(t, tss, "Group Member User", "group.member@saldeti.local")
	userID := *user.GetId()

	// Add user to group via SDK
	memberRef := models.NewReferenceCreate()
	memberRef.SetOdataId(ptrString(fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, userID)))
	err := tss.SDKClient.Groups().ByGroupId(groupID).Members().Ref().Post(ctx, memberRef, nil)
	require.NoError(t, err, "Failed to add user to group")

	// GET /groups/{groupID}?$expand=members
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1.0/groups/%s?$expand=members", tss.BaseURL, groupID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))

	// Assert response contains "members" array
	members, ok := result["members"].([]interface{})
	require.True(t, ok, "Expected members field to be an array")
	assert.NotEmpty(t, members, "Expected at least one member")

	// Verify the user is in the list
	found := false
	for _, m := range members {
		mMap, ok := m.(map[string]interface{})
		require.True(t, ok, "Expected member to be an object")
		mID, ok := mMap["id"].(string)
		if ok && mID == userID {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected to find user in group members")

	t.Logf("Successfully expanded members for group %s", groupID)
}

// TestExpandGroupOwners tests expanding owners on a group
func TestExpandGroupOwners(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	token := getToken(t, tss)

	// Create a group
	group := createTestGroupSDK(t, tss, "Group Owners Test")
	groupID := *group.GetId()

	// Create a user
	user := createTestUserSDK(t, tss, "Group Owner User", "group.owner@saldeti.local")
	userID := *user.GetId()

	// Add user as owner to group via SDK
	ownerRef := models.NewReferenceCreate()
	ownerRef.SetOdataId(ptrString(fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, userID)))
	err := tss.SDKClient.Groups().ByGroupId(groupID).Owners().Ref().Post(ctx, ownerRef, nil)
	require.NoError(t, err, "Failed to add user as owner to group")

	// GET /groups/{groupID}?$expand=owners
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1.0/groups/%s?$expand=owners", tss.BaseURL, groupID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))

	// Assert response contains "owners" array
	owners, ok := result["owners"].([]interface{})
	require.True(t, ok, "Expected owners field to be an array")
	assert.NotEmpty(t, owners, "Expected at least one owner")

	// Verify the user is in the list
	found := false
	for _, o := range owners {
		oMap, ok := o.(map[string]interface{})
		require.True(t, ok, "Expected owner to be an object")
		oID, ok := oMap["id"].(string)
		if ok && oID == userID {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected to find user in group owners")

	t.Logf("Successfully expanded owners for group %s", groupID)
}

// TestExpandGroupMemberOf tests expanding memberOf on a group (nested groups)
func TestExpandGroupMemberOf(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	token := getToken(t, tss)

	// Create a child group
	childGroup := createTestGroupSDK(t, tss, "Child Group")
	childGroupID := *childGroup.GetId()

	// Create a parent group
	parentGroup := createTestGroupSDK(t, tss, "Parent Group")
	parentGroupID := *parentGroup.GetId()

	// Add child group as member of parent group
	memberRef := models.NewReferenceCreate()
	memberRef.SetOdataId(ptrString(fmt.Sprintf("%s/v1.0/groups/%s", tss.BaseURL, childGroupID)))
	err := tss.SDKClient.Groups().ByGroupId(parentGroupID).Members().Ref().Post(ctx, memberRef, nil)
	require.NoError(t, err, "Failed to add child group to parent group")

	// GET /groups/{childGroupID}?$expand=memberOf
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1.0/groups/%s?$expand=memberOf", tss.BaseURL, childGroupID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))

	// Assert response contains "memberOf" array
	memberOf, ok := result["memberOf"].([]interface{})
	require.True(t, ok, "Expected memberOf field to be an array")
	assert.NotEmpty(t, memberOf, "Expected at least one memberOf entry")

	// Verify the parent group is in the list
	found := false
	for _, mo := range memberOf {
		moMap, ok := mo.(map[string]interface{})
		require.True(t, ok, "Expected memberOf entry to be an object")
		moID, ok := moMap["id"].(string)
		if ok && moID == parentGroupID {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected to find parent group in child group's memberOf")

	t.Logf("Successfully expanded memberOf for group %s", childGroupID)
}

// TestExpandMultipleProps tests expanding multiple properties at once
func TestExpandMultipleProps(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	token := getToken(t, tss)

	// Create two users
	user1 := createTestUserSDK(t, tss, "User Multiple Props", "user.multiple@saldeti.local")
	user1ID := *user1.GetId()

	user2 := createTestUserSDK(t, tss, "Manager Multiple", "manager.multiple@saldeti.local")
	user2ID := *user2.GetId()

	// Set user2 as manager of user1
	refBody := models.NewReferenceUpdate()
	odataID := fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, user2ID)
	refBody.SetOdataId(&odataID)
	err := tss.SDKClient.Users().ByUserId(user1ID).Manager().Ref().Put(ctx, refBody, nil)
	require.NoError(t, err, "Failed to set manager")

	// Create a group and add user1 to it
	group := createTestGroupSDK(t, tss, "Multiple Props Group")
	groupID := *group.GetId()

	memberRef := models.NewReferenceCreate()
	memberRef.SetOdataId(ptrString(fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, user1ID)))
	err = tss.SDKClient.Groups().ByGroupId(groupID).Members().Ref().Post(ctx, memberRef, nil)
	require.NoError(t, err, "Failed to add user to group")

	// GET /users/{user1ID}?$expand=manager,memberOf
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1.0/users/%s?$expand=manager,memberOf", tss.BaseURL, user1ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))

	// Assert both "manager" and "memberOf" are present
	manager, ok := result["manager"].(map[string]interface{})
	require.True(t, ok, "Expected manager field to be present")
	require.NotNil(t, manager, "Expected manager to not be null")
	assert.Equal(t, user2ID, manager["id"], "Expected manager ID to match")

	memberOf, ok := result["memberOf"].([]interface{})
	require.True(t, ok, "Expected memberOf field to be an array")
	assert.NotEmpty(t, memberOf, "Expected at least one memberOf entry")

	// Verify the group is in memberOf
	found := false
	for _, mo := range memberOf {
		moMap, ok := mo.(map[string]interface{})
		if ok {
			moID, ok := moMap["id"].(string)
			if ok && moID == groupID {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "Expected to find group in user's memberOf")

	t.Logf("Successfully expanded multiple properties for user %s", user1ID)
}

// TestExpandUnknownProperty tests expanding a property that doesn't exist
func TestExpandUnknownProperty(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	token := getToken(t, tss)

	// Create a user
	user := createTestUserSDK(t, tss, "User Unknown Prop", "user.unknown@saldeti.local")
	userID := *user.GetId()

	// GET /users/{userID}?$expand=nonexistent
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1.0/users/%s?$expand=nonexistent", tss.BaseURL, userID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should succeed without error (unknown property is ignored)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK for unknown expand property")

	t.Logf("Successfully handled unknown expand property for user %s", userID)
}

// TestExpandListUsersManager tests expanding manager on a list of users
func TestExpandListUsersManager(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	token := getToken(t, tss)

	// Create a manager
	manager := createTestUserSDK(t, tss, "List Manager", "list.manager@saldeti.local")
	managerID := *manager.GetId()

	// Create two users and set the same manager for both
	userIDs := []string{}
	for i := 1; i <= 2; i++ {
		displayName := fmt.Sprintf("List User %d", i)
		upn := fmt.Sprintf("listuser%d@saldeti.local", i)
		user := createTestUserSDK(t, tss, displayName, upn)
		userID := *user.GetId()
		userIDs = append(userIDs, userID)

		// Set manager
		refBody := models.NewReferenceUpdate()
		odataID := fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, managerID)
		refBody.SetOdataId(&odataID)
		err := tss.SDKClient.Users().ByUserId(userID).Manager().Ref().Put(ctx, refBody, nil)
		require.NoError(t, err, "Failed to set manager for user %d", i)
	}

	// GET /users?$expand=manager
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1.0/users?$expand=manager", tss.BaseURL), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))

	// Assert each user has a "manager" field
	users, ok := result["value"].([]interface{})
	require.True(t, ok, "Expected value to be an array")
	assert.NotEmpty(t, users, "Expected at least one user")

	// Check that our created users have the manager expanded
	foundCount := 0
	for _, u := range users {
		userMap, ok := u.(map[string]interface{})
		require.True(t, ok, "Expected user to be an object")

		// Check for manager field presence
		managerField, hasManager := userMap["manager"]
		if hasManager {
			// If it's one of our users, verify the manager
			userID, ok := userMap["id"].(string)
			if ok {
				for _, createdID := range userIDs {
					if userID == createdID {
						foundCount++
						managerMap, ok := managerField.(map[string]interface{})
						require.True(t, ok, "Expected manager to be an object")
						assert.Equal(t, managerID, managerMap["id"], "Expected manager ID to match")
						break
					}
				}
			}
		}
	}

	assert.GreaterOrEqual(t, foundCount, 2, "Expected to find at least 2 users with expanded manager")

	t.Logf("Successfully expanded manager for list of users")
}

// TestExpandListGroupsMembers tests expanding members on a list of groups
func TestExpandListGroupsMembers(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	token := getToken(t, tss)

	// Create two groups and add a member to each
	groupIDs := []string{}
	for i := 1; i <= 2; i++ {
		displayName := fmt.Sprintf("List Group %d", i)
		group := createTestGroupSDK(t, tss, displayName)
		groupID := *group.GetId()
		groupIDs = append(groupIDs, groupID)

		// Create a user and add to group
		displayName = fmt.Sprintf("List Group User %d", i)
		upn := fmt.Sprintf("listgroupuser%d@saldeti.local", i)
		user := createTestUserSDK(t, tss, displayName, upn)
		userID := *user.GetId()

		memberRef := models.NewReferenceCreate()
		memberRef.SetOdataId(ptrString(fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, userID)))
		err := tss.SDKClient.Groups().ByGroupId(groupID).Members().Ref().Post(ctx, memberRef, nil)
		require.NoError(t, err, "Failed to add user to group %d", i)
	}

	// GET /groups?$expand=members
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1.0/groups?$expand=members", tss.BaseURL), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))

	// Assert each group has a "members" field
	groups, ok := result["value"].([]interface{})
	require.True(t, ok, "Expected value to be an array")
	assert.NotEmpty(t, groups, "Expected at least one group")

	// Check that our created groups have members expanded
	foundCount := 0
	for _, g := range groups {
		groupMap, ok := g.(map[string]interface{})
		require.True(t, ok, "Expected group to be an object")

		// Check for members field presence
		membersField, hasMembers := groupMap["members"]
		if hasMembers {
			// If it's one of our groups, verify the members
			groupID, ok := groupMap["id"].(string)
			if ok {
				for _, createdID := range groupIDs {
					if groupID == createdID {
						foundCount++
						membersArray, ok := membersField.([]interface{})
						require.True(t, ok, "Expected members to be an array")
						assert.NotEmpty(t, membersArray, "Expected at least one member")
						break
					}
				}
			}
		}
	}

	assert.GreaterOrEqual(t, foundCount, 2, "Expected to find at least 2 groups with expanded members")

	t.Logf("Successfully expanded members for list of groups")
}
