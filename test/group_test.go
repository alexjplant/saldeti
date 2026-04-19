//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_CreateGroupWithMembers tests creating a group and adding members
func TestE2E_CreateGroupWithMembers(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// 1. Create 2 users via SDK
	userIDs := []string{}
	for i := 1; i <= 2; i++ {
		displayName := fmt.Sprintf("Test User %d", i)
		upn := fmt.Sprintf("test%d@saldeti.local", i)
		user := createTestUserSDK(t, tss, displayName, upn)
		userIDs = append(userIDs, *user.GetId())
	}

	// 2. Create a group via SDK
	newGroup := models.NewGroup()
	displayName := "Test Group"
	mailNickname := "testgroup"
	secEnabled := true
	mailEnabled := false
	newGroup.SetDisplayName(&displayName)
	newGroup.SetMailNickname(&mailNickname)
	newGroup.SetSecurityEnabled(&secEnabled)
	newGroup.SetMailEnabled(&mailEnabled)

	created, err := tss.SDKClient.Groups().Post(ctx, newGroup, nil)
	require.NoError(t, err, "Failed to create group")
	require.NotNil(t, created)

	groupID := *created.GetId()

	// 3. Add both users as members via SDK (using $ref)
	for _, userID := range userIDs {
		refBody := models.NewReferenceCreate()
		refBody.SetOdataId(ptrString(fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, userID)))
		err := tss.SDKClient.Groups().ByGroupId(groupID).Members().Ref().Post(ctx, refBody, nil)
		require.NoError(t, err, "Failed to add member %s", userID)
	}

	// 4. List members via SDK, verify 2 members
	members, err := tss.SDKClient.Groups().ByGroupId(groupID).Members().Get(ctx, nil)
	require.NoError(t, err, "Failed to get group members")

	memberList := members.GetValue()
	assert.Len(t, memberList, 2, "Expected 2 members")
}

// TestE2E_GroupCRUDLifecycle tests creating, reading, updating, and deleting a group
func TestE2E_GroupCRUDLifecycle(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// 1. Create a group via SDK
	newGroup := models.NewGroup()
	displayName := "Test Group CRUD"
	mailNickname := "testgroupcrud"
	secEnabled := true
	mailEnabled := false
	newGroup.SetDisplayName(&displayName)
	newGroup.SetMailNickname(&mailNickname)
	newGroup.SetSecurityEnabled(&secEnabled)
	newGroup.SetMailEnabled(&mailEnabled)

	created, err := tss.SDKClient.Groups().Post(ctx, newGroup, nil)
	require.NoError(t, err, "Failed to create group")
	require.NotNil(t, created)

	groupID := *created.GetId()

	// 2. GET the group by ID via SDK
	group, err := tss.SDKClient.Groups().ByGroupId(groupID).Get(ctx, nil)
	require.NoError(t, err, "Failed to get group")
	require.NotNil(t, group.GetDisplayName(), "Group should have displayName")
	assert.Equal(t, "Test Group CRUD", *group.GetDisplayName())

	// 3. PATCH the group (change displayName) via SDK
	patch := models.NewGroup()
	updatedName := "Updated Group"
	patch.SetDisplayName(&updatedName)

	_, err = tss.SDKClient.Groups().ByGroupId(groupID).Patch(ctx, patch, nil)
	require.NoError(t, err, "Failed to patch group")

	// 4. GET again via SDK to verify update
	verifyGroup, err := tss.SDKClient.Groups().ByGroupId(groupID).Get(ctx, nil)
	require.NoError(t, err, "Failed to get updated group")
	assert.Equal(t, "Updated Group", *verifyGroup.GetDisplayName())

	// 5. DELETE the group via SDK
	err = tss.SDKClient.Groups().ByGroupId(groupID).Delete(ctx, nil)
	require.NoError(t, err, "Failed to delete group")

	// 6. GET again, expect error (404)
	_, err = tss.SDKClient.Groups().ByGroupId(groupID).Get(ctx, nil)
	require.Error(t, err, "Expected error getting deleted group")
}

// TestE2E_GroupOwnerManagement tests adding, listing, and removing group owners
func TestE2E_GroupOwnerManagement(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// 1. Create a user via SDK
	user := createTestUserSDK(t, tss, "Owner User", "owner@saldeti.local")
	userID := *user.GetId()

	// 2. Create a group via SDK
	group := createTestGroupSDK(t, tss, "Owner Test Group")
	groupID := *group.GetId()

	// 3. Add owner via SDK (using $ref)
	ownerRef := models.NewReferenceCreate()
	ownerRef.SetOdataId(ptrString(fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, userID)))
	err := tss.SDKClient.Groups().ByGroupId(groupID).Owners().Ref().Post(ctx, ownerRef, nil)
	require.NoError(t, err, "Failed to add owner")

	// 4. List owners via SDK
	owners, err := tss.SDKClient.Groups().ByGroupId(groupID).Owners().Get(ctx, nil)
	require.NoError(t, err, "Failed to list owners")

	ownerList := owners.GetValue()
	assert.Len(t, ownerList, 1, "Expected 1 owner")

	// 5. Remove owner via SDK
	err = tss.SDKClient.Groups().ByGroupId(groupID).Owners().ByDirectoryObjectId(userID).Ref().Delete(ctx, nil)
	require.NoError(t, err, "Failed to remove owner")

	// 6. List owners again via SDK, verify none
	updatedOwners, err := tss.SDKClient.Groups().ByGroupId(groupID).Owners().Get(ctx, nil)
	require.NoError(t, err, "Failed to list updated owners")

	updatedOwnerList := updatedOwners.GetValue()
	assert.Len(t, updatedOwnerList, 0, "Expected 0 owners after removal")
}

// TestE2E_SDKListGroupsWithMembers tests listing groups and their members
func TestE2E_SDKListGroupsWithMembers(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// Create a user via SDK
	user := createTestUserSDK(t, tss, "SDK Test User", "sdktest@saldeti.local")
	userID := *user.GetId()

	// Create a group via SDK
	group := createTestGroupSDK(t, tss, "SDK Test Group")
	groupID := *group.GetId()

	// Add user to group via SDK (using $ref)
	memberRef := models.NewReferenceCreate()
	memberRef.SetOdataId(ptrString(fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, userID)))
	err := tss.SDKClient.Groups().ByGroupId(groupID).Members().Ref().Post(ctx, memberRef, nil)
	require.NoError(t, err, "Failed to add member")

	// List groups via SDK
	groupsResult, err := tss.SDKClient.Groups().Get(ctx, nil)
	require.NoError(t, err, "Failed to list groups")

	groupList := groupsResult.GetValue()
	assert.NotEmpty(t, groupList, "Expected at least 1 group")

	// Verify group structure
	firstGroup := groupList[0]
	assert.NotNil(t, firstGroup.GetId(), "Group missing id field")
	assert.NotNil(t, firstGroup.GetDisplayName(), "Group missing displayName field")
	assert.NotNil(t, firstGroup.GetSecurityEnabled(), "Group missing securityEnabled field")
	assert.NotNil(t, firstGroup.GetMailEnabled(), "Group missing mailEnabled field")

	// Get members of the group via SDK
	members, err := tss.SDKClient.Groups().ByGroupId(groupID).Members().Get(ctx, nil)
	require.NoError(t, err, "Failed to get group members")

	memberList := members.GetValue()
	assert.NotEmpty(t, memberList, "Expected at least 1 member")

	// Verify @odata.type is present in members
	firstMember := memberList[0]
	assert.NotNil(t, firstMember.GetOdataType(), "Member missing @odata.type field")
}
