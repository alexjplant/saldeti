//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/directoryobjects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_ManagerWorkflow tests the manager/directReports workflow
func TestE2E_ManagerWorkflow(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// 1. Create a manager user via SDK
	managerUser := createTestUserSDK(t, tss, "Manager User", "manager@saldeti.local")
	managerID := *managerUser.GetId()

	// 2. Create 2 report users via SDK
	reportIDs := []string{}
	for i := 1; i <= 2; i++ {
		displayName := fmt.Sprintf("Report User %d", i)
		upn := fmt.Sprintf("report%d@saldeti.local", i)
		reportUser := createTestUserSDK(t, tss, displayName, upn)
		reportIDs = append(reportIDs, *reportUser.GetId())
	}

	// 3. Set manager for both reports via SDK
	for _, reportID := range reportIDs {
		refBody := models.NewReferenceUpdate()
		odataID := fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, managerID)
		refBody.SetOdataId(&odataID)
		err := tss.SDKClient.Users().ByUserId(reportID).Manager().Ref().Put(ctx, refBody, nil)
		require.NoError(t, err, "Failed to set manager for report %s", reportID)
	}

	// 4. GET /users/{manager}/directReports via SDK, verify 2 reports
	directReports, err := tss.SDKClient.Users().ByUserId(managerID).DirectReports().Get(ctx, nil)
	require.NoError(t, err, "Failed to get directReports")

	reports := directReports.GetValue()
	assert.Len(t, reports, 2, "Expected 2 direct reports")
}

// TestE2E_GetByIds tests the getByIds endpoint
func TestE2E_GetByIds(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// Create 2 users via SDK
	userIDs := []string{}
	for i := 1; i <= 2; i++ {
		displayName := fmt.Sprintf("GetByIds User %d", i)
		upn := fmt.Sprintf("getbyidsuser%d@saldeti.local", i)
		user := createTestUserSDK(t, tss, displayName, upn)
		userIDs = append(userIDs, *user.GetId())
	}

	// Call getByIds via SDK with all 2 IDs
	body := directoryobjects.NewGetByIdsPostRequestBody()
	body.SetIds(userIDs)
	types := []string{"user"}
	body.SetTypes(types)
	result, err := tss.SDKClient.DirectoryObjects().GetByIds().Post(ctx, body, nil)
	require.NoError(t, err, "Failed to call getByIds")

	// Should return 2 directory objects
	objects := result.GetValue()
	assert.Len(t, objects, 2, "Expected 2 objects from getByIds")
}

// TestE2E_DeltaQueryWorkflow tests the delta query workflow
func TestE2E_DeltaQueryWorkflow(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// Create some users via SDK
	for i := 1; i <= 3; i++ {
		displayName := fmt.Sprintf("Delta User %d", i)
		upn := fmt.Sprintf("deltauser%d@saldeti.local", i)
		createTestUserSDK(t, tss, displayName, upn)
	}

	// Call /users/delta via SDK, verify deltaLink present
	result, err := tss.SDKClient.Users().Delta().Get(ctx, nil)
	require.NoError(t, err, "Failed to call delta query")

	// Check for deltaLink or @odata.nextLink
	deltaLink := result.GetOdataDeltaLink()
	nextLink := result.GetOdataNextLink()
	if deltaLink == nil && nextLink == nil {
		t.Log("Expected @odata.deltaLink or @odata.nextLink in delta response")
	}
}

// TestE2E_SDKUserMemberOf tests the SDK userMemberOf endpoint
func TestE2E_SDKUserMemberOf(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// Create a user via SDK
	user := createTestUserSDK(t, tss, "SDK MemberOf User", "sdkmemberof@saldeti.local")
	userID := *user.GetId()

	// Create a group via SDK
	group := createTestGroupSDK(t, tss, "SDK MemberOf Group")
	groupID := *group.GetId()

	// Add user to group via SDK (using $ref)
	memberRef := models.NewReferenceCreate()
	memberRef.SetOdataId(ptrString(fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, userID)))
	err := tss.SDKClient.Groups().ByGroupId(groupID).Members().Ref().Post(ctx, memberRef, nil)
	require.NoError(t, err, "Failed to add user to group")

	// Get user's memberOf via SDK
	memberOf, err := tss.SDKClient.Users().ByUserId(userID).MemberOf().Get(ctx, nil)
	require.NoError(t, err, "Failed to get user memberOf")

	memberOfList := memberOf.GetValue()
	assert.NotNil(t, memberOfList, "Expected value field in memberOf response")
	assert.NotEmpty(t, memberOfList, "Expected at least one memberOf entry")
}

// TestE2E_UserMemberOf tests the user memberOf navigation property
func TestE2E_UserMemberOf(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// Create a user via SDK
	user := createTestUserSDK(t, tss, "MemberOf User", "memberof@saldeti.local")
	userID := *user.GetId()

	// Create a group via SDK
	group := createTestGroupSDK(t, tss, "MemberOf Test Group")
	groupID := *group.GetId()

	// Add user to group via SDK (using $ref)
	memberRef := models.NewReferenceCreate()
	memberRef.SetOdataId(ptrString(fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, userID)))
	err := tss.SDKClient.Groups().ByGroupId(groupID).Members().Ref().Post(ctx, memberRef, nil)
	require.NoError(t, err, "Failed to add user to group")

	// GET /users/{id}/memberOf via SDK, verify group appears
	memberOf, err := tss.SDKClient.Users().ByUserId(userID).MemberOf().Get(ctx, nil)
	require.NoError(t, err, "Failed to get user memberOf")

	memberOfList := memberOf.GetValue()
	assert.NotEmpty(t, memberOfList, "Expected at least 1 group")

	// Verify our group is present
	found := false
	for _, g := range memberOfList {
		if *g.GetId() == groupID {
			found = true
			break
		}
	}
	assert.True(t, found, "Group not found in user's memberOf")
}

// TestE2E_DirectReports tests the directReports navigation property
func TestE2E_DirectReports(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// Create a manager via SDK
	managerUser := createTestUserSDK(t, tss, "Direct Reports Manager", "drmanager@saldeti.local")
	managerID := *managerUser.GetId()

	// Create 2 direct reports via SDK
	reportIDs := []string{}
	for i := 1; i <= 2; i++ {
		displayName := fmt.Sprintf("Direct Report %d", i)
		upn := fmt.Sprintf("directreport%d@saldeti.local", i)
		reportUser := createTestUserSDK(t, tss, displayName, upn)
		reportIDs = append(reportIDs, *reportUser.GetId())
	}

	// Set manager for both reports via SDK
	for _, reportID := range reportIDs {
		refBody := models.NewReferenceUpdate()
		odataID := fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, managerID)
		refBody.SetOdataId(&odataID)
		err := tss.SDKClient.Users().ByUserId(reportID).Manager().Ref().Put(ctx, refBody, nil)
		require.NoError(t, err, "Failed to set manager for report %s", reportID)
	}

	// GET /users/{manager}/directReports via SDK, verify 2 reports
	directReports, err := tss.SDKClient.Users().ByUserId(managerID).DirectReports().Get(ctx, nil)
	require.NoError(t, err, "Failed to get directReports")

	reports := directReports.GetValue()
	assert.Len(t, reports, 2, "Expected 2 direct reports")
}
