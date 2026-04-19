//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/microsoftgraph/msgraph-sdk-go/directoryobjects"
)

// TestSDKCompatibility tests every Microsoft Graph SDK operation systematically
// to identify what works and what breaks against the simulator API.
func TestSDKCompatibility(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	var userID, groupID string
	var testUserIDs []string
	var testGroupIDs []string

	t.Run("UserOperations", func(t *testing.T) {
		t.Run("Users_Post_CreateUser", func(t *testing.T) {
			t.Log("Attempting to create user via SDK.Users().Post()")

			newUser := models.NewUser()
			displayName := "SDK Compat User"
			newUser.SetDisplayName(&displayName)
			upn := "sdkcompat@saldeti.local"
			newUser.SetUserPrincipalName(&upn)
			newUser.SetMail(&upn)

			accountEnabled := true
			newUser.SetAccountEnabled(&accountEnabled)

			department := "Test Department"
			newUser.SetDepartment(&department)

			passwordProfile := models.NewPasswordProfile()
			password := "Test1234!"
			passwordProfile.SetPassword(&password)
			newUser.SetPasswordProfile(passwordProfile)

			userType := "Member"
			newUser.SetUserType(&userType)

			result, err := tss.SDKClient.Users().Post(ctx, newUser, nil)
			if err != nil {
				t.Fatalf("Users.Post failed: %v (type: %T)", err, err)
			}
			if result == nil {
				t.Fatal("Users.Post returned nil result (no error)")
			}
			if result.GetId() == nil || *result.GetId() == "" {
				t.Fatal("Users.Post returned user with nil/empty ID")
			}
			userID = *result.GetId()
			testUserIDs = append(testUserIDs, userID)
			t.Logf("Created user: id=%s, displayName=%s, upn=%s", *result.GetId(), *result.GetDisplayName(), *result.GetUserPrincipalName())
		})

		t.Run("Users_ById_Get", func(t *testing.T) {
			if userID == "" {
				t.Skip("Skipping: user ID not available")
			}
			t.Logf("Attempting to get user via SDK.Users().ByUserId(%s).Get()", userID)

			result, err := tss.SDKClient.Users().ByUserId(userID).Get(ctx, nil)
			if err != nil {
				t.Fatalf("Users.ByUserId(%s).Get failed: %v (type: %T)", userID, err, err)
			}
			if result == nil {
				t.Fatal("Users.ByUserId().Get returned nil result (no error)")
			}
			t.Logf("Got user: id=%s, displayName=%s, department=%s", *result.GetId(), *result.GetDisplayName(), *result.GetDepartment())
		})

		t.Run("Users_ById_Patch", func(t *testing.T) {
			if userID == "" {
				t.Skip("Skipping: user ID not available")
			}
			t.Logf("Attempting to update user via SDK.Users().ByUserId(%s).Patch()", userID)

			patchUser := models.NewUser()
			newDept := "Updated Department"
			patchUser.SetDepartment(&newDept)

			result, err := tss.SDKClient.Users().ByUserId(userID).Patch(ctx, patchUser, nil)
			if err != nil {
				t.Fatalf("Users.ByUserId(%s).Patch failed: %v (type: %T)", userID, err, err)
			}
			if result == nil {
				t.Fatal("Users.ByUserId().Patch returned nil result (no error)")
			}
			t.Logf("Updated user: id=%s, newDepartment=%s", *result.GetId(), *result.GetDepartment())
		})

		t.Run("Users_ById_Delete", func(t *testing.T) {
			if userID == "" {
				t.Skip("Skipping: user ID not available")
			}
			t.Logf("Attempting to delete user via SDK.Users().ByUserId(%s).Delete()", userID)

			err := tss.SDKClient.Users().ByUserId(userID).Delete(ctx, nil)
			if err != nil {
				t.Fatalf("Users.ByUserId(%s).Delete failed: %v (type: %T)", userID, err, err)
			}
			t.Logf("Successfully deleted user: id=%s", userID)

			// Verify deletion by trying to get the user
			_, err = tss.SDKClient.Users().ByUserId(userID).Get(ctx, nil)
			if err == nil {
				t.Fatalf("User was deleted but can still be retrieved")
			} else {
				t.Logf("Confirmed deletion: Get returned error as expected: %v", err)
			}
		})

		t.Run("Users_Get_List", func(t *testing.T) {
			t.Log("Attempting to list users via SDK.Users().Get()")

			result, err := tss.SDKClient.Users().Get(ctx, nil)
			if err != nil {
				t.Fatalf("Users.Get failed: %v (type: %T)", err, err)
			}
			if result == nil {
				t.Fatal("Users.Get returned nil result (no error)")
			}
			users := result.GetValue()
			if users == nil {
				t.Fatal("Users.Get returned nil users list (no error)")
			}
			t.Logf("Listed %d users", len(users))
			if len(users) > 0 {
				for i, u := range users {
					t.Logf("  User %d: id=%s, displayName=%s", i+1, *u.GetId(), *u.GetDisplayName())
				}
			}
		})
	})

	t.Run("GroupOperations", func(t *testing.T) {
		t.Run("Groups_Post_CreateGroup", func(t *testing.T) {
			t.Log("Attempting to create group via SDK.Groups().Post()")

			newGroup := models.NewGroup()
			displayName := "SDK Compat Test Group"
			newGroup.SetDisplayName(&displayName)
			mailNickname := "sdkcompgroup"
			newGroup.SetMailNickname(&mailNickname)

			securityEnabled := true
			newGroup.SetSecurityEnabled(&securityEnabled)
			mailEnabled := false
			newGroup.SetMailEnabled(&mailEnabled)

			result, err := tss.SDKClient.Groups().Post(ctx, newGroup, nil)
			if err != nil {
				t.Fatalf("Groups.Post failed: %v (type: %T)", err, err)
			}
			if result == nil {
				t.Fatal("Groups.Post returned nil result (no error)")
			}
			if result.GetId() == nil || *result.GetId() == "" {
				t.Fatal("Groups.Post returned group with nil/empty ID")
			}
			groupID = *result.GetId()
			testGroupIDs = append(testGroupIDs, groupID)
			t.Logf("Created group: id=%s, displayName=%s", *result.GetId(), *result.GetDisplayName())
		})

		t.Run("Groups_ById_Get", func(t *testing.T) {
			if groupID == "" {
				t.Skip("Skipping: group ID not available")
			}
			t.Logf("Attempting to get group via SDK.Groups().ByGroupId(%s).Get()", groupID)

			result, err := tss.SDKClient.Groups().ByGroupId(groupID).Get(ctx, nil)
			if err != nil {
				t.Fatalf("Groups.ByGroupId(%s).Get failed: %v (type: %T)", groupID, err, err)
			}
			if result == nil {
				t.Fatal("Groups.ByGroupId().Get returned nil result (no error)")
			}
			t.Logf("Got group: id=%s, displayName=%s, securityEnabled=%v", *result.GetId(), *result.GetDisplayName(), *result.GetSecurityEnabled())
		})

		t.Run("Groups_ById_Patch", func(t *testing.T) {
			if groupID == "" {
				t.Skip("Skipping: group ID not available")
			}
			t.Logf("Attempting to update group via SDK.Groups().ByGroupId(%s).Patch()", groupID)

			patchGroup := models.NewGroup()
			newDesc := "Updated description"
			patchGroup.SetDescription(&newDesc)

			result, err := tss.SDKClient.Groups().ByGroupId(groupID).Patch(ctx, patchGroup, nil)
			if err != nil {
				t.Fatalf("Groups.ByGroupId(%s).Patch failed: %v (type: %T)", groupID, err, err)
			}
			if result == nil {
				t.Fatal("Groups.ByGroupId().Patch returned nil result (no error)")
			}
			t.Logf("Updated group: id=%s", *result.GetId())
		})

		t.Run("Groups_ById_Delete", func(t *testing.T) {
			if groupID == "" {
				t.Skip("Skipping: group ID not available")
			}
			t.Logf("Attempting to delete group via SDK.Groups().ByGroupId(%s).Delete()", groupID)

			err := tss.SDKClient.Groups().ByGroupId(groupID).Delete(ctx, nil)
			if err != nil {
				t.Fatalf("Groups.ByGroupId(%s).Delete failed: %v (type: %T)", groupID, err, err)
			}
			t.Logf("Successfully deleted group: id=%s", groupID)

			// Verify deletion by trying to get the group
			_, err = tss.SDKClient.Groups().ByGroupId(groupID).Get(ctx, nil)
			if err == nil {
				t.Fatalf("Group was deleted but can still be retrieved")
			} else {
				t.Logf("Confirmed deletion: Get returned error as expected: %v", err)
			}
		})

		t.Run("Groups_Get_List", func(t *testing.T) {
			t.Log("Attempting to list groups via SDK.Groups().Get()")

			result, err := tss.SDKClient.Groups().Get(ctx, nil)
			if err != nil {
				t.Fatalf("Groups.Get failed: %v (type: %T)", err, err)
			}
			if result == nil {
				t.Fatal("Groups.Get returned nil result (no error)")
			}
			groups := result.GetValue()
			if groups == nil {
				t.Fatal("Groups.Get returned nil groups list (no error)")
			}
			t.Logf("Listed %d groups", len(groups))
			if len(groups) > 0 {
				for i, g := range groups {
					t.Logf("  Group %d: id=%s, displayName=%s", i+1, *g.GetId(), *g.GetDisplayName())
				}
			}
		})
	})

	t.Run("MembershipOperations", func(t *testing.T) {
		// Create test users and groups for membership operations
		var memberUser1ID, memberUser2ID, memberGroupID string

		t.Run("Setup_CreateUsersAndGroup", func(t *testing.T) {
			t.Log("Setting up: Creating test users and group for membership tests")

			// Create users
			for i := 1; i <= 2; i++ {
				newUser := models.NewUser()
				displayName := fmt.Sprintf("Member User %d", i)
				newUser.SetDisplayName(&displayName)
				upn := fmt.Sprintf("memberuser%d@saldeti.local", i)
				newUser.SetUserPrincipalName(&upn)
				newUser.SetMail(&upn)

				accountEnabled := true
				newUser.SetAccountEnabled(&accountEnabled)

				passwordProfile := models.NewPasswordProfile()
				password := "Test1234!"
				passwordProfile.SetPassword(&password)
				newUser.SetPasswordProfile(passwordProfile)

				userType := "Member"
				newUser.SetUserType(&userType)

				result, err := tss.SDKClient.Users().Post(ctx, newUser, nil)
				if err != nil {
					t.Fatalf("Setup: Failed to create user %d: %v (type: %T)", i, err, err)
				}
				if result == nil || result.GetId() == nil {
					t.Fatalf("Setup: User %d created but ID is nil", i)
				}

				if i == 1 {
					memberUser1ID = *result.GetId()
				} else {
					memberUser2ID = *result.GetId()
				}
				t.Logf("Setup: Created user %d: id=%s", i, *result.GetId())
			}

			// Create group
			newGroup := models.NewGroup()
			displayName := "Membership Test Group"
			newGroup.SetDisplayName(&displayName)
			mailNickname := "membershiptestgroup"
			newGroup.SetMailNickname(&mailNickname)

			securityEnabled := true
			newGroup.SetSecurityEnabled(&securityEnabled)
			mailEnabled := false
			newGroup.SetMailEnabled(&mailEnabled)

			result, err := tss.SDKClient.Groups().Post(ctx, newGroup, nil)
			if err != nil {
				t.Fatalf("Setup: Failed to create group: %v (type: %T)", err, err)
			}
			if result == nil || result.GetId() == nil {
				t.Fatal("Setup: Group created but ID is nil")
			}
			memberGroupID = *result.GetId()
			t.Logf("Setup: Created group: id=%s", memberGroupID)
		})

		t.Run("Groups_ById_Members_Ref_Post_AddMember", func(t *testing.T) {
			if memberUser1ID == "" || memberGroupID == "" {
				t.Skip("Skipping: user ID or group ID not available")
			}
			t.Logf("Attempting to add member via SDK.Groups().ByGroupId(%s).Members().Ref().Post()", memberGroupID)

			refBody := models.NewReferenceCreate()
			odataID := fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, memberUser1ID)
			refBody.SetOdataId(&odataID)

			err := tss.SDKClient.Groups().ByGroupId(memberGroupID).Members().Ref().Post(ctx, refBody, nil)
			if err != nil {
				t.Fatalf("Groups.ByGroupId(%s).Members().Ref().Post failed: %v (type: %T)", memberGroupID, err, err)
			}
			t.Logf("Successfully added member: userId=%s to groupId=%s", memberUser1ID, memberGroupID)
		})

		t.Run("Groups_ById_Members_Get_ListMembers", func(t *testing.T) {
			if memberGroupID == "" {
				t.Skip("Skipping: group ID not available")
			}
			t.Logf("Attempting to list members via SDK.Groups().ByGroupId(%s).Members().Get()", memberGroupID)

			result, err := tss.SDKClient.Groups().ByGroupId(memberGroupID).Members().Get(ctx, nil)
			if err != nil {
				t.Fatalf("Groups.ByGroupId(%s).Members().Get failed: %v (type: %T)", memberGroupID, err, err)
			}
			if result == nil {
				t.Fatal("Groups.ByGroupId().Members().Get returned nil result (no error)")
			}
			members := result.GetValue()
			if members == nil {
				t.Fatal("Groups.ByGroupId().Members().Get returned nil members list (no error)")
			}
			t.Logf("Listed %d members", len(members))
			if len(members) > 0 {
				for i, m := range members {
					t.Logf("  Member %d: id=%s, @odata.type=%s", i+1, *m.GetId(), *m.GetOdataType())
				}
			}
		})

		t.Run("Groups_ById_Members_ById_Ref_Delete_RemoveMember", func(t *testing.T) {
			if memberUser1ID == "" || memberGroupID == "" {
				t.Skip("Skipping: user ID or group ID not available")
			}
			t.Logf("Attempting to remove member via SDK.Groups().ByGroupId(%s).Members().ByDirectoryObjectId(%s).Ref().Delete()", memberGroupID, memberUser1ID)

			err := tss.SDKClient.Groups().ByGroupId(memberGroupID).Members().ByDirectoryObjectId(memberUser1ID).Ref().Delete(ctx, nil)
			if err != nil {
				t.Fatalf("Groups.ByGroupId(%s).Members().ByDirectoryObjectId(%s).Ref().Delete failed: %v (type: %T)", memberGroupID, memberUser1ID, err, err)
			}
			t.Logf("Successfully removed member: userId=%s from groupId=%s", memberUser1ID, memberGroupID)
		})

		t.Run("Groups_ById_Owners_Ref_Post_AddOwner", func(t *testing.T) {
			if memberUser2ID == "" || memberGroupID == "" {
				t.Skip("Skipping: user ID or group ID not available")
			}
			t.Logf("Attempting to add owner via SDK.Groups().ByGroupId(%s).Owners().Ref().Post()", memberGroupID)

			refBody := models.NewReferenceCreate()
			odataID := fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, memberUser2ID)
			refBody.SetOdataId(&odataID)

			err := tss.SDKClient.Groups().ByGroupId(memberGroupID).Owners().Ref().Post(ctx, refBody, nil)
			if err != nil {
				t.Fatalf("Groups.ByGroupId(%s).Owners().Ref().Post failed: %v (type: %T)", memberGroupID, err, err)
			}
			t.Logf("Successfully added owner: userId=%s to groupId=%s", memberUser2ID, memberGroupID)
		})

		t.Run("Groups_ById_Owners_Get_ListOwners", func(t *testing.T) {
			if memberGroupID == "" {
				t.Skip("Skipping: group ID not available")
			}
			t.Logf("Attempting to list owners via SDK.Groups().ByGroupId(%s).Owners().Get()", memberGroupID)

			result, err := tss.SDKClient.Groups().ByGroupId(memberGroupID).Owners().Get(ctx, nil)
			if err != nil {
				t.Fatalf("Groups.ByGroupId(%s).Owners().Get failed: %v (type: %T)", memberGroupID, err, err)
			}
			if result == nil {
				t.Fatal("Groups.ByGroupId().Owners().Get returned nil result (no error)")
			}
			owners := result.GetValue()
			if owners == nil {
				t.Fatal("Groups.ByGroupId().Owners().Get returned nil owners list (no error)")
			}
			t.Logf("Listed %d owners", len(owners))
			if len(owners) > 0 {
				for i, o := range owners {
					t.Logf("  Owner %d: id=%s, @odata.type=%s", i+1, *o.GetId(), *o.GetOdataType())
				}
			}
		})

		t.Run("Groups_ById_Owners_ById_Ref_Delete_RemoveOwner", func(t *testing.T) {
			if memberUser2ID == "" || memberGroupID == "" {
				t.Skip("Skipping: user ID or group ID not available")
			}
			t.Logf("Attempting to remove owner via SDK.Groups().ByGroupId(%s).Owners().ByDirectoryObjectId(%s).Ref().Delete()", memberGroupID, memberUser2ID)

			err := tss.SDKClient.Groups().ByGroupId(memberGroupID).Owners().ByDirectoryObjectId(memberUser2ID).Ref().Delete(ctx, nil)
			if err != nil {
				t.Fatalf("Groups.ByGroupId(%s).Owners().ByDirectoryObjectId(%s).Ref().Delete failed: %v (type: %T)", memberGroupID, memberUser2ID, err, err)
			}
			t.Logf("Successfully removed owner: userId=%s from groupId=%s", memberUser2ID, memberGroupID)
		})
	})

	t.Run("NavigationOperations", func(t *testing.T) {
		var navUserID, navGroupID string

		t.Run("Setup_CreateUserAndGroup", func(t *testing.T) {
			t.Log("Setup: Creating user and group for navigation tests")

			// Create user
			newUser := models.NewUser()
			displayName := "Nav Test User"
			newUser.SetDisplayName(&displayName)
			upn := "navtestuser@saldeti.local"
			newUser.SetUserPrincipalName(&upn)
			newUser.SetMail(&upn)

			accountEnabled := true
			newUser.SetAccountEnabled(&accountEnabled)

			passwordProfile := models.NewPasswordProfile()
			password := "Test1234!"
			passwordProfile.SetPassword(&password)
			newUser.SetPasswordProfile(passwordProfile)

			userType := "Member"
			newUser.SetUserType(&userType)

			result, err := tss.SDKClient.Users().Post(ctx, newUser, nil)
			if err != nil {
				t.Fatalf("Setup: Failed to create user: %v (type: %T)", err, err)
			}
			if result == nil || result.GetId() == nil {
				t.Fatal("Setup: User created but ID is nil")
			}
			navUserID = *result.GetId()
			t.Logf("Setup: Created user: id=%s", navUserID)

			// Create group
			newGroup := models.NewGroup()
			displayName = "Nav Test Group"
			newGroup.SetDisplayName(&displayName)
			mailNickname := "navtestgroup"
			newGroup.SetMailNickname(&mailNickname)

			securityEnabled := true
			newGroup.SetSecurityEnabled(&securityEnabled)
			mailEnabled := false
			newGroup.SetMailEnabled(&mailEnabled)

			groupResult, err := tss.SDKClient.Groups().Post(ctx, newGroup, nil)
			if err != nil {
				t.Fatalf("Setup: Failed to create group: %v (type: %T)", err, err)
			}
			if groupResult == nil || groupResult.GetId() == nil {
				t.Fatal("Setup: Group created but ID is nil")
			}
			navGroupID = *groupResult.GetId()
			t.Logf("Setup: Created group: id=%s", navGroupID)

			// Add user to group
			refBody := models.NewReferenceCreate()
			odataID := fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, navUserID)
			refBody.SetOdataId(&odataID)

			err = tss.SDKClient.Groups().ByGroupId(navGroupID).Members().Ref().Post(ctx, refBody, nil)
			if err != nil {
				t.Fatalf("Setup: Failed to add user to group: %v (type: %T)", err, err)
			}
			t.Logf("Setup: Added user to group")
		})

		t.Run("Users_ById_MemberOf_Get", func(t *testing.T) {
			if navUserID == "" {
				t.Skip("Skipping: user ID not available")
			}
			t.Logf("Attempting to get user memberOf via SDK.Users().ByUserId(%s).MemberOf().Get()", navUserID)

			result, err := tss.SDKClient.Users().ByUserId(navUserID).MemberOf().Get(ctx, nil)
			if err != nil {
				t.Fatalf("Users.ByUserId(%s).MemberOf().Get failed: %v (type: %T)", navUserID, err, err)
			}
			if result == nil {
				t.Fatal("Users.ByUserId().MemberOf().Get returned nil result (no error)")
			}
			memberOf := result.GetValue()
			if memberOf == nil {
				t.Fatal("Users.ByUserId().MemberOf().Get returned nil memberOf list (no error)")
			}
			t.Logf("User is member of %d groups", len(memberOf))
			if len(memberOf) > 0 {
				for i, g := range memberOf {
					// Try to cast to Groupable to get displayName
					if grp, ok := g.(models.Groupable); ok && grp.GetDisplayName() != nil {
						t.Logf("  Group %d: id=%s, displayName=%s", i+1, *g.GetId(), *grp.GetDisplayName())
					} else {
						t.Logf("  Group %d: id=%s, @odata.type=%s", i+1, *g.GetId(), *g.GetOdataType())
					}
				}
			}
		})

		t.Run("Users_ById_TransitiveMemberOf_Get", func(t *testing.T) {
			if navUserID == "" {
				t.Skip("Skipping: user ID not available")
			}
			t.Logf("Attempting to get user transitiveMemberOf via SDK.Users().ByUserId(%s).TransitiveMemberOf().Get()", navUserID)

			result, err := tss.SDKClient.Users().ByUserId(navUserID).TransitiveMemberOf().Get(ctx, nil)
			if err != nil {
				t.Fatalf("Users.ByUserId(%s).TransitiveMemberOf().Get failed: %v (type: %T)", navUserID, err, err)
			}
			if result == nil {
				t.Fatal("Users.ByUserId().TransitiveMemberOf().Get returned nil result (no error)")
			}
			transitiveMemberOf := result.GetValue()
			if transitiveMemberOf == nil {
				t.Fatal("Users.ByUserId().TransitiveMemberOf().Get returned nil transitiveMemberOf list (no error)")
			}
			t.Logf("User is transitively member of %d groups", len(transitiveMemberOf))
			if len(transitiveMemberOf) > 0 {
				for i, g := range transitiveMemberOf {
					// Try to cast to Groupable to get displayName
					if grp, ok := g.(models.Groupable); ok && grp.GetDisplayName() != nil {
						t.Logf("  Group %d: id=%s, displayName=%s", i+1, *g.GetId(), *grp.GetDisplayName())
					} else {
						t.Logf("  Group %d: id=%s, @odata.type=%s", i+1, *g.GetId(), *g.GetOdataType())
					}
				}
			}
		})

		t.Run("Users_ById_DirectReports_Get", func(t *testing.T) {
			if navUserID == "" {
				t.Skip("Skipping: user ID not available")
			}
			t.Logf("Attempting to get user directReports via SDK.Users().ByUserId(%s).DirectReports().Get()", navUserID)

			result, err := tss.SDKClient.Users().ByUserId(navUserID).DirectReports().Get(ctx, nil)
			if err != nil {
				t.Fatalf("Users.ByUserId(%s).DirectReports().Get failed: %v (type: %T)", navUserID, err, err)
			}
			if result == nil {
				t.Fatal("Users.ByUserId().DirectReports().Get returned nil result (no error)")
			}
			directReports := result.GetValue()
			if directReports == nil {
				t.Fatal("Users.ByUserId().DirectReports().Get returned nil directReports list (no error)")
			}
			t.Logf("User has %d direct reports", len(directReports))
			if len(directReports) > 0 {
				for i, r := range directReports {
					// Try to cast to Userable to get displayName
					if usr, ok := r.(models.Userable); ok && usr.GetDisplayName() != nil {
						t.Logf("  Direct Report %d: id=%s, displayName=%s", i+1, *r.GetId(), *usr.GetDisplayName())
					} else {
						t.Logf("  Direct Report %d: id=%s, @odata.type=%s", i+1, *r.GetId(), *r.GetOdataType())
					}
				}
			}
		})

		t.Run("Groups_ById_TransitiveMembers_Get", func(t *testing.T) {
			if navGroupID == "" {
				t.Skip("Skipping: group ID not available")
			}
			t.Logf("Attempting to get group transitiveMembers via SDK.Groups().ByGroupId(%s).TransitiveMembers().Get()", navGroupID)

			result, err := tss.SDKClient.Groups().ByGroupId(navGroupID).TransitiveMembers().Get(ctx, nil)
			if err != nil {
				t.Fatalf("Groups.ByGroupId(%s).TransitiveMembers().Get failed: %v (type: %T)", navGroupID, err, err)
			}
			if result == nil {
				t.Fatal("Groups.ByGroupId().TransitiveMembers().Get returned nil result (no error)")
			}
			transitiveMembers := result.GetValue()
			if transitiveMembers == nil {
				t.Fatal("Groups.ByGroupId().TransitiveMembers().Get returned nil transitiveMembers list (no error)")
			}
			t.Logf("Group has %d transitive members", len(transitiveMembers))
			if len(transitiveMembers) > 0 {
				for i, m := range transitiveMembers {
					t.Logf("  Transitive Member %d: id=%s, @odata.type=%s", i+1, *m.GetId(), *m.GetOdataType())
				}
			}
		})
	})

	t.Run("ManagerOperations", func(t *testing.T) {
		var managerUserID, reportUserID string

		t.Run("Setup_CreateManagerAndReport", func(t *testing.T) {
			t.Log("Setup: Creating manager and report users for manager tests")

			// Create manager
			newUser := models.NewUser()
			displayName := "Manager User"
			newUser.SetDisplayName(&displayName)
			upn := "manageruser@saldeti.local"
			newUser.SetUserPrincipalName(&upn)
			newUser.SetMail(&upn)

			accountEnabled := true
			newUser.SetAccountEnabled(&accountEnabled)

			passwordProfile := models.NewPasswordProfile()
			password := "Test1234!"
			passwordProfile.SetPassword(&password)
			newUser.SetPasswordProfile(passwordProfile)

			userType := "Member"
			newUser.SetUserType(&userType)

			result, err := tss.SDKClient.Users().Post(ctx, newUser, nil)
			if err != nil {
				t.Fatalf("Setup: Failed to create manager user: %v (type: %T)", err, err)
			}
			if result == nil || result.GetId() == nil {
				t.Fatal("Setup: Manager user created but ID is nil")
			}
			managerUserID = *result.GetId()
			t.Logf("Setup: Created manager user: id=%s", managerUserID)

			// Create report
			newUser = models.NewUser()
			displayName = "Report User"
			newUser.SetDisplayName(&displayName)
			upn = "reportuser@saldeti.local"
			newUser.SetUserPrincipalName(&upn)
			newUser.SetMail(&upn)

			newUser.SetAccountEnabled(&accountEnabled)

			passwordProfile = models.NewPasswordProfile()
			passwordProfile.SetPassword(&password)
			newUser.SetPasswordProfile(passwordProfile)

			newUser.SetUserType(&userType)

			result, err = tss.SDKClient.Users().Post(ctx, newUser, nil)
			if err != nil {
				t.Fatalf("Setup: Failed to create report user: %v (type: %T)", err, err)
			}
			if result == nil || result.GetId() == nil {
				t.Fatal("Setup: Report user created but ID is nil")
			}
			reportUserID = *result.GetId()
			t.Logf("Setup: Created report user: id=%s", reportUserID)
		})

		t.Run("Users_ById_Manager_Ref_Put_SetManager", func(t *testing.T) {
			if managerUserID == "" || reportUserID == "" {
				t.Skip("Skipping: manager ID or report ID not available")
			}
			t.Logf("Attempting to set manager via SDK.Users().ByUserId(%s).Manager().Ref().Put()", reportUserID)

			refBody := models.NewReferenceUpdate()
			odataID := fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, managerUserID)
			refBody.SetOdataId(&odataID)

			err := tss.SDKClient.Users().ByUserId(reportUserID).Manager().Ref().Put(ctx, refBody, nil)
			if err != nil {
				t.Fatalf("Users.ByUserId(%s).Manager().Ref().Put failed: %v (type: %T)", reportUserID, err, err)
			}
			t.Logf("Successfully set manager: managerId=%s for userId=%s", managerUserID, reportUserID)
		})

		t.Run("Users_ById_Manager_Get", func(t *testing.T) {
			if reportUserID == "" {
				t.Skip("Skipping: report ID not available")
			}
			t.Logf("Attempting to get manager via SDK.Users().ByUserId(%s).Manager().Get()", reportUserID)

			result, err := tss.SDKClient.Users().ByUserId(reportUserID).Manager().Get(ctx, nil)
			if err != nil {
				t.Fatalf("Users.ByUserId(%s).Manager().Get failed: %v (type: %T)", reportUserID, err, err)
			}
			if result == nil {
				t.Fatal("Users.ByUserId().Manager().Get returned nil result (no error)")
			}
			if result.GetId() == nil {
				t.Fatal("Users.ByUserId().Manager().Get returned result with nil ID")
			}
			// Try to cast to Userable to get displayName
			if usr, ok := result.(models.Userable); ok && usr.GetDisplayName() != nil {
				t.Logf("Got manager: id=%s, displayName=%s", *result.GetId(), *usr.GetDisplayName())
			} else {
				t.Logf("Got manager: id=%s, @odata.type=%s", *result.GetId(), *result.GetOdataType())
			}
		})

		t.Run("Users_ById_Manager_Ref_Delete_RemoveManager", func(t *testing.T) {
			if reportUserID == "" {
				t.Skip("Skipping: report ID not available")
			}
			t.Logf("Attempting to remove manager via SDK.Users().ByUserId(%s).Manager().Ref().Delete()", reportUserID)

			err := tss.SDKClient.Users().ByUserId(reportUserID).Manager().Ref().Delete(ctx, nil)
			if err != nil {
				t.Fatalf("Users.ByUserId(%s).Manager().Ref().Delete failed: %v (type: %T)", reportUserID, err, err)
			}
			t.Logf("Successfully removed manager from userId=%s", reportUserID)
		})
	})

	t.Run("ODataQueryOperations", func(t *testing.T) {
		t.Run("Users_Get_WithFilter", func(t *testing.T) {
			t.Log("Attempting to list users with $filter query parameter")

			filter := "accountEnabled eq true"
			result, err := tss.SDKClient.Users().Get(ctx, &users.UsersRequestBuilderGetRequestConfiguration{
				QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
					Filter: &filter,
				},
			})
			if err != nil {
				t.Fatalf("Users.Get with $filter failed: %v (type: %T)", err, err)
			}
			if result == nil {
				t.Fatal("Users.Get with $filter returned nil result (no error)")
			}
			users := result.GetValue()
			if users == nil {
				t.Fatal("Users.Get with $filter returned nil users list (no error)")
			}
			t.Logf("Listed %d users with filter=%s", len(users), filter)
			if len(users) > 0 {
				for i, u := range users {
					t.Logf("  User %d: id=%s, displayName=%s, accountEnabled=%v", i+1, *u.GetId(), *u.GetDisplayName(), *u.GetAccountEnabled())
				}
			}
		})

		t.Run("Users_Get_WithOrderby", func(t *testing.T) {
			t.Log("Attempting to list users with $orderby query parameter")

			orderby := []string{"displayName"}
			result, err := tss.SDKClient.Users().Get(ctx, &users.UsersRequestBuilderGetRequestConfiguration{
				QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
					Orderby: orderby,
				},
			})
			if err != nil {
				t.Fatalf("Users.Get with $orderby failed: %v (type: %T)", err, err)
			}
			if result == nil {
				t.Fatal("Users.Get with $orderby returned nil result (no error)")
			}
			users := result.GetValue()
			if users == nil {
				t.Fatal("Users.Get with $orderby returned nil users list (no error)")
			}
			t.Logf("Listed %d users with orderby=%v", len(users), orderby)
			if len(users) > 0 {
				for i, u := range users {
					t.Logf("  User %d: displayName=%s", i+1, *u.GetDisplayName())
				}
			}
		})

		t.Run("Users_Get_WithSearch", func(t *testing.T) {
			t.Log("Attempting to list users with $search query parameter")

			search := "Test"
			result, err := tss.SDKClient.Users().Get(ctx, &users.UsersRequestBuilderGetRequestConfiguration{
				QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
					Search: &search,
				},
			})
			if err != nil {
				t.Fatalf("Users.Get with $search failed: %v (type: %T)", err, err)
			}
			if result == nil {
				t.Fatal("Users.Get with $search returned nil result (no error)")
			}
			users := result.GetValue()
			if users == nil {
				t.Fatal("Users.Get with $search returned nil users list (no error)")
			}
			t.Logf("Listed %d users with search=%s", len(users), search)
			if len(users) > 0 {
				for i, u := range users {
					t.Logf("  User %d: id=%s, displayName=%s", i+1, *u.GetId(), *u.GetDisplayName())
				}
			}
		})

		t.Run("Users_Get_WithTop", func(t *testing.T) {
			t.Log("Attempting to list users with $top query parameter")

			top := int32(5)
			result, err := tss.SDKClient.Users().Get(ctx, &users.UsersRequestBuilderGetRequestConfiguration{
				QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
					Top: &top,
				},
			})
			if err != nil {
				t.Fatalf("Users.Get with $top failed: %v (type: %T)", err, err)
			}
			if result == nil {
				t.Fatal("Users.Get with $top returned nil result (no error)")
			}
			users := result.GetValue()
			if users == nil {
				t.Fatal("Users.Get with $top returned nil users list (no error)")
			}
			t.Logf("Listed %d users with top=%d", len(users), top)
			if len(users) > 0 {
				for i, u := range users {
					t.Logf("  User %d: id=%s, displayName=%s", i+1, *u.GetId(), *u.GetDisplayName())
				}
			}
		})

		t.Run("Users_Get_WithCount", func(t *testing.T) {
			t.Log("Attempting to list users with $count query parameter")

			count := true
			result, err := tss.SDKClient.Users().Get(ctx, &users.UsersRequestBuilderGetRequestConfiguration{
				QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
					Count: &count,
				},
			})
			if err != nil {
				t.Fatalf("Users.Get with $count failed: %v (type: %T)", err, err)
			}
			if result == nil {
				t.Fatal("Users.Get with $count returned nil result (no error)")
			}
			users := result.GetValue()
			if users == nil {
				t.Fatal("Users.Get with $count returned nil users list (no error)")
			}
			t.Logf("Listed %d users with count=true", len(users))

			// Check for @odata.count in additional data
			additionalData := result.GetAdditionalData()
			if odataCount, ok := additionalData["@odata.count"]; ok {
				t.Logf("OData count: %v", odataCount)
			} else {
				t.Log("Note: @odata.count not found in additional data")
			}
		})
	})

	t.Run("ActionFunctionOperations", func(t *testing.T) {
		var actionUserID, actionGroupID string

		t.Run("Setup_CreateUserAndGroup", func(t *testing.T) {
			t.Log("Setup: Creating user and group for action/function tests")

			// Create user
			newUser := models.NewUser()
			displayName := "Action User"
			newUser.SetDisplayName(&displayName)
			upn := "actionuser@saldeti.local"
			newUser.SetUserPrincipalName(&upn)
			newUser.SetMail(&upn)

			accountEnabled := true
			newUser.SetAccountEnabled(&accountEnabled)

			passwordProfile := models.NewPasswordProfile()
			password := "Test1234!"
			passwordProfile.SetPassword(&password)
			newUser.SetPasswordProfile(passwordProfile)

			userType := "Member"
			newUser.SetUserType(&userType)

			result, err := tss.SDKClient.Users().Post(ctx, newUser, nil)
			if err != nil {
				t.Fatalf("Setup: Failed to create user: %v (type: %T)", err, err)
			}
			if result == nil || result.GetId() == nil {
				t.Fatal("Setup: User created but ID is nil")
			}
			actionUserID = *result.GetId()
			t.Logf("Setup: Created user: id=%s", actionUserID)

			// Create group
			newGroup := models.NewGroup()
			displayName = "Action Group"
			newGroup.SetDisplayName(&displayName)
			mailNickname := "actiongroup"
			newGroup.SetMailNickname(&mailNickname)

			securityEnabled := true
			newGroup.SetSecurityEnabled(&securityEnabled)
			mailEnabled := false
			newGroup.SetMailEnabled(&mailEnabled)

			groupResult, err := tss.SDKClient.Groups().Post(ctx, newGroup, nil)
			if err != nil {
				t.Fatalf("Setup: Failed to create group: %v (type: %T)", err, err)
			}
			if groupResult == nil || groupResult.GetId() == nil {
				t.Fatal("Setup: Group created but ID is nil")
			}
			actionGroupID = *groupResult.GetId()
			t.Logf("Setup: Created group: id=%s", actionGroupID)

			// Add user to group
			refBody := models.NewReferenceCreate()
			odataID := fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, actionUserID)
			refBody.SetOdataId(&odataID)

			err = tss.SDKClient.Groups().ByGroupId(actionGroupID).Members().Ref().Post(ctx, refBody, nil)
			if err != nil {
				t.Fatalf("Setup: Failed to add user to group: %v (type: %T)", err, err)
			}
			t.Logf("Setup: Added user to group")
		})

		t.Run("Users_ById_CheckMemberGroups", func(t *testing.T) {
			if actionUserID == "" || actionGroupID == "" {
				t.Skip("Skipping: user ID or group ID not available")
			}
			t.Logf("Attempting to call checkMemberGroups via SDK.Users().ByUserId(%s).CheckMemberGroups()", actionUserID)

			checkGroupsBody := users.NewItemCheckMemberGroupsPostRequestBody()
			groupIDs := []string{actionGroupID}
			checkGroupsBody.SetGroupIds(groupIDs)

			result, err := tss.SDKClient.Users().ByUserId(actionUserID).CheckMemberGroups().Post(ctx, checkGroupsBody, nil)
			if err != nil {
				t.Fatalf("Users.ByUserId(%s).CheckMemberGroups().Post failed: %v (type: %T)", actionUserID, err, err)
			}
			if result == nil {
				t.Fatal("Users.ByUserId().CheckMemberGroups().Post returned nil result (no error)")
			}
			value := result.GetValue()
			if value == nil {
				t.Fatal("Users.ByUserId().CheckMemberGroups().Post returned nil value (no error)")
			}
			t.Logf("checkMemberGroups returned %d matching group IDs", len(value))
			for i, id := range value {
				t.Logf("  Matching Group ID %d: %s", i+1, id)
			}
		})

		t.Run("Users_ById_GetMemberGroups", func(t *testing.T) {
			if actionUserID == "" {
				t.Skip("Skipping: user ID not available")
			}
			t.Logf("Attempting to call getMemberGroups via SDK.Users().ByUserId(%s).GetMemberGroups()", actionUserID)

			getGroupsBody := users.NewItemGetMemberGroupsPostRequestBody()
			securityEnabledOnly := true
			getGroupsBody.SetSecurityEnabledOnly(&securityEnabledOnly)

			result, err := tss.SDKClient.Users().ByUserId(actionUserID).GetMemberGroups().Post(ctx, getGroupsBody, nil)
			if err != nil {
				t.Fatalf("Users.ByUserId(%s).GetMemberGroups().Post failed: %v (type: %T)", actionUserID, err, err)
			}
			if result == nil {
				t.Fatal("Users.ByUserId().GetMemberGroups().Post returned nil result (no error)")
			}
			value := result.GetValue()
			if value == nil {
				t.Fatal("Users.ByUserId().GetMemberGroups().Post returned nil value (no error)")
			}
			t.Logf("getMemberGroups returned %d group IDs", len(value))
			for i, id := range value {
				t.Logf("  Group ID %d: %s", i+1, id)
			}
		})

		t.Run("DirectoryObjects_GetByIds", func(t *testing.T) {
			if actionUserID == "" {
				t.Skip("Skipping: user ID not available")
			}
			t.Log("Attempting to call getByIds via SDK.DirectoryObjects().GetByIds()")

			getByIdsBody := directoryobjects.NewGetByIdsPostRequestBody()
			ids := []string{actionUserID}
			getByIdsBody.SetIds(ids)
			types := []string{"microsoft.graph.user"}
			getByIdsBody.SetTypes(types)

			result, err := tss.SDKClient.DirectoryObjects().GetByIds().Post(ctx, getByIdsBody, nil)
			if err != nil {
				t.Fatalf("DirectoryObjects().GetByIds().Post failed: %v (type: %T)", err, err)
			}
			if result == nil {
				t.Fatal("DirectoryObjects().GetByIds().Post returned nil result (no error)")
			}
			value := result.GetValue()
			if value == nil {
				t.Fatal("DirectoryObjects().GetByIds().Post returned nil value (no error)")
			}
			t.Logf("getByIds returned %d directory objects", len(value))
			for i, obj := range value {
				t.Logf("  Object %d: id=%s, @odata.type=%s", i+1, *obj.GetId(), *obj.GetOdataType())
			}
		})

		t.Run("Users_Delta_Get", func(t *testing.T) {
			t.Log("Attempting to call delta query via SDK.Users().Delta().Get()")

			result, err := tss.SDKClient.Users().Delta().Get(ctx, nil)
			if err != nil {
				t.Fatalf("Users().Delta().Get failed: %v (type: %T)", err, err)
			}
			if result == nil {
				t.Fatal("Users().Delta().Get returned nil result (no error)")
			}
			users := result.GetValue()
			if users == nil {
				t.Fatal("Users().Delta().Get returned nil users list (no error)")
			}
			t.Logf("Delta query returned %d users", len(users))

			// Check for deltaLink or nextLink
			deltaLink := result.GetOdataDeltaLink()
			nextLink := result.GetOdataNextLink()
			if deltaLink != nil {
				t.Logf("Delta link: %s", *deltaLink)
			}
			if nextLink != nil {
				t.Logf("Next link: %s", *nextLink)
			}
		})
	})
}
