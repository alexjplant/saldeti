package store

import (
	"context"
	"testing"
	"time"

	"github.com/saldeti/saldeti/internal/entra/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStore_UserOperations(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	// Test CreateUser
	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "test@example.com",
		Mail:              "test@example.com",
		AccountEnabled:    &accountEnabled,
	}

	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)
	assert.NotEmpty(t, createdUser.ID)
	assert.NotNil(t, createdUser.CreatedDateTime)
	assert.WithinDuration(t, time.Now(), *createdUser.CreatedDateTime, time.Second)

	// Test GetUser
	retrievedUser, err := store.GetUser(ctx, createdUser.ID)
	require.NoError(t, err)
	assert.Equal(t, createdUser.ID, retrievedUser.ID)
	assert.Equal(t, "test@example.com", retrievedUser.UserPrincipalName)
	assert.Equal(t, "Test User", retrievedUser.DisplayName)

	// Test GetUserByUPN
	retrievedByUPN, err := store.GetUserByUPN(ctx, "test@example.com")
	require.NoError(t, err)
	assert.Equal(t, createdUser.ID, retrievedByUPN.ID)

	// Test duplicate user
	_, err = store.CreateUser(ctx, user)
	assert.ErrorIs(t, err, ErrDuplicateUPN)

	// Test GetUser with non-existent ID
	_, err = store.GetUser(ctx, "non-existent-id")
	assert.ErrorIs(t, err, ErrUserNotFound)

	// Test GetUserByUPN with non-existent UPN
	_, err = store.GetUserByUPN(ctx, "nonexistent@example.com")
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestMemoryStore_ClientOperations(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	// Test RegisterClient
	err := store.RegisterClient(ctx, "client1", "secret1", "tenant1")
	require.NoError(t, err)

	// Test GetClient
	clientID, secret, tenantID, err := store.GetClient(ctx, "client1")
	require.NoError(t, err)
	assert.Equal(t, "client1", clientID)
	assert.Equal(t, "secret1", secret)
	assert.Equal(t, "tenant1", tenantID)

	// Test duplicate client registration
	err = store.RegisterClient(ctx, "client1", "secret2", "tenant2")
	assert.ErrorIs(t, err, ErrDuplicateClient)

	// Test GetClient with non-existent client
	_, _, _, err = store.GetClient(ctx, "non-existent-client")
	assert.ErrorIs(t, err, ErrClientNotFound)
}

func TestRegisterClient(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	err := store.RegisterClient(ctx, "test-client-id", "test-secret", "00000000-0000-0000-0000-000000000001")
	require.NoError(t, err)

	// Verify the auto-created SP has AppOwnerOrganizationId populated with the valid UUID
	sp, err := store.GetServicePrincipalByAppID(ctx, "test-client-id")
	require.NoError(t, err)
	assert.NotEmpty(t, sp.AppOwnerOrganizationID, "AppOwnerOrganizationId should be set on the auto-created service principal")
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", sp.AppOwnerOrganizationID, "AppOwnerOrganizationId should match the tenantID passed to RegisterClient")

	// Verify that a non-UUID tenantID results in an empty AppOwnerOrganizationID
	err = store.RegisterClient(ctx, "test-client-id-2", "test-secret-2", "not-a-uuid")
	require.NoError(t, err)
	sp2, err := store.GetServicePrincipalByAppID(ctx, "test-client-id-2")
	require.NoError(t, err)
	assert.Empty(t, sp2.AppOwnerOrganizationID, "AppOwnerOrganizationId should be empty when tenantID is not a valid UUID")
}

func TestDeleteUser_CascadesToAppOwners(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore().(*memoryStore)

	// Create a user
	accountEnabled := true
	user, err := store.CreateUser(ctx, model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "test@example.com",
		AccountEnabled:    &accountEnabled,
	})
	require.NoError(t, err)

	// Create an application
	app, err := store.CreateApplication(ctx, model.Application{
		DisplayName: "Test App",
	})
	require.NoError(t, err)

	// Add user as app owner
	err = store.AddApplicationOwner(ctx, app.ID, user.ID, "user")
	require.NoError(t, err)

	// Verify user is an app owner
	owners, _, err := store.ListApplicationOwners(ctx, app.ID, model.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, owners, 1)
	assert.Equal(t, user.ID, owners[0].ID)

	// Delete user
	err = store.DeleteUser(ctx, user.ID)
	require.NoError(t, err)

	// Verify user is gone from app owners
	owners, _, err = store.ListApplicationOwners(ctx, app.ID, model.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, owners, 0)

	// Also verify directly on the internal map
	assert.NotContains(t, store.appOwners[app.ID], user.ID)
}

func TestDeleteUser_CascadesToSPOwners(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore().(*memoryStore)

	// Create a user
	accountEnabled := true
	user, err := store.CreateUser(ctx, model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "test@example.com",
		AccountEnabled:    &accountEnabled,
	})
	require.NoError(t, err)

	// Create an application (auto-creates SP)
	app, err := store.CreateApplication(ctx, model.Application{
		DisplayName: "Test App",
	})
	require.NoError(t, err)

	// Find the auto-created SP
	sp, err := store.GetServicePrincipalByAppID(ctx, app.AppID)
	require.NoError(t, err)

	// Add user as SP owner
	err = store.AddSPOwner(ctx, sp.ID, user.ID, "user")
	require.NoError(t, err)

	// Verify user is an SP owner
	owners, _, err := store.ListSPOwners(ctx, sp.ID, model.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, owners, 1)

	// Delete user
	err = store.DeleteUser(ctx, user.ID)
	require.NoError(t, err)

	// Verify user is gone from SP owners
	owners, _, err = store.ListSPOwners(ctx, sp.ID, model.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, owners, 0)

	// Also verify directly on the internal map
	assert.NotContains(t, store.spOwners[sp.ID], user.ID)
}

func TestDeleteUser_CascadesToAppRoleAssignments(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore().(*memoryStore)

	// Create a user
	accountEnabled := true
	user, err := store.CreateUser(ctx, model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "test@example.com",
		AccountEnabled:    &accountEnabled,
	})
	require.NoError(t, err)

	// Create an application with an app role
	roleEnabled := true
	roleID := "00000000-0000-0000-0000-000000000001"
	app, err := store.CreateApplication(ctx, model.Application{
		DisplayName: "Test App",
		AppRoles: []model.AppRole{
			{
				ID:                 roleID,
				AllowedMemberTypes: []string{"User"},
				DisplayName:        "Test Role",
				IsEnabled:          &roleEnabled,
			},
		},
	})
	require.NoError(t, err)

	// Get the auto-created SP
	sp, err := store.GetServicePrincipalByAppID(ctx, app.AppID)
	require.NoError(t, err)

	// Create an app role assignment for the user
	assignment, err := store.CreateAppRoleAssignment(ctx, sp.ID, user.ID, roleID)
	require.NoError(t, err)

	// Verify assignment exists
	assignments, _, err := store.ListAppRoleAssignments(ctx, user.ID, model.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, assignments, 1)
	assert.Equal(t, assignment.ID, assignments[0].ID)

	// Delete user
	err = store.DeleteUser(ctx, user.ID)
	require.NoError(t, err)

	// Verify assignment is gone from internal map
	assert.NotContains(t, store.appRoleAssignments, assignment.ID)
}

func TestDeleteUser_CascadesToOAuth2PermissionGrants(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore().(*memoryStore)

	// Create a user
	accountEnabled := true
	user, err := store.CreateUser(ctx, model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "test@example.com",
		AccountEnabled:    &accountEnabled,
	})
	require.NoError(t, err)

	// Create an application (auto-creates SP)
	app, err := store.CreateApplication(ctx, model.Application{
		DisplayName: "Test App",
	})
	require.NoError(t, err)

	sp, err := store.GetServicePrincipalByAppID(ctx, app.AppID)
	require.NoError(t, err)

	// Create an OAuth2 permission grant with user as principal
	grant, err := store.CreateOAuth2PermissionGrant(ctx, model.OAuth2PermissionGrant{
		ClientID:    sp.ID,
		ConsentType: "Principal",
		PrincipalID: user.ID,
		ResourceID:  sp.ID,
		Scope:       "User.Read",
	})
	require.NoError(t, err)

	// Verify grant exists
	_, err = store.GetOAuth2PermissionGrant(ctx, grant.ID)
	require.NoError(t, err)

	// Delete user
	err = store.DeleteUser(ctx, user.ID)
	require.NoError(t, err)

	// Verify grant is gone
	_, err = store.GetOAuth2PermissionGrant(ctx, grant.ID)
	assert.ErrorIs(t, err, ErrGrantNotFound)

	// Also verify on internal map
	assert.NotContains(t, store.oauth2PermissionGrants, grant.ID)
}

func TestDeleteGroup_CascadesToAppOwners(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore().(*memoryStore)

	// Create a group
	group, err := store.CreateGroup(ctx, model.Group{
		DisplayName: "Test Group",
	})
	require.NoError(t, err)

	// Create an application
	app, err := store.CreateApplication(ctx, model.Application{
		DisplayName: "Test App",
	})
	require.NoError(t, err)

	// Add group as app owner
	err = store.AddApplicationOwner(ctx, app.ID, group.ID, "group")
	require.NoError(t, err)

	// Verify group is an app owner
	owners, _, err := store.ListApplicationOwners(ctx, app.ID, model.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, owners, 1)

	// Delete group
	err = store.DeleteGroup(ctx, group.ID)
	require.NoError(t, err)

	// Verify group is gone from app owners
	owners, _, err = store.ListApplicationOwners(ctx, app.ID, model.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, owners, 0)

	// Also verify directly on the internal map
	assert.NotContains(t, store.appOwners[app.ID], group.ID)
}

func TestDeleteGroup_CascadesToSPOwners(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore().(*memoryStore)

	// Create a group
	group, err := store.CreateGroup(ctx, model.Group{
		DisplayName: "Test Group",
	})
	require.NoError(t, err)

	// Create an application (auto-creates SP)
	app, err := store.CreateApplication(ctx, model.Application{
		DisplayName: "Test App",
	})
	require.NoError(t, err)

	sp, err := store.GetServicePrincipalByAppID(ctx, app.AppID)
	require.NoError(t, err)

	// Add group as SP owner
	err = store.AddSPOwner(ctx, sp.ID, group.ID, "group")
	require.NoError(t, err)

	// Delete group
	err = store.DeleteGroup(ctx, group.ID)
	require.NoError(t, err)

	// Verify group is gone from SP owners
	owners, _, err := store.ListSPOwners(ctx, sp.ID, model.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, owners, 0)

	assert.NotContains(t, store.spOwners[sp.ID], group.ID)
}

func TestDeleteGroup_CascadesToAppRoleAssignments(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore().(*memoryStore)

	// Create a group
	group, err := store.CreateGroup(ctx, model.Group{
		DisplayName: "Test Group",
	})
	require.NoError(t, err)

	// Create an application with an app role
	roleEnabled := true
	roleID := "00000000-0000-0000-0000-000000000002"
	app, err := store.CreateApplication(ctx, model.Application{
		DisplayName: "Test App",
		AppRoles: []model.AppRole{
			{
				ID:                 roleID,
				AllowedMemberTypes: []string{"User", "Group"},
				DisplayName:        "Test Role",
				IsEnabled:          &roleEnabled,
			},
		},
	})
	require.NoError(t, err)

	sp, err := store.GetServicePrincipalByAppID(ctx, app.AppID)
	require.NoError(t, err)

	// Create an app role assignment for the group
	assignment, err := store.CreateAppRoleAssignment(ctx, sp.ID, group.ID, roleID)
	require.NoError(t, err)

	// Delete group
	err = store.DeleteGroup(ctx, group.ID)
	require.NoError(t, err)

	// Verify assignment is gone
	assert.NotContains(t, store.appRoleAssignments, assignment.ID)
}

func TestDeleteGroup_CascadesToOAuth2PermissionGrants(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore().(*memoryStore)

	// Create a group
	group, err := store.CreateGroup(ctx, model.Group{
		DisplayName: "Test Group",
	})
	require.NoError(t, err)

	// Create an application (auto-creates SP)
	app, err := store.CreateApplication(ctx, model.Application{
		DisplayName: "Test App",
	})
	require.NoError(t, err)

	sp, err := store.GetServicePrincipalByAppID(ctx, app.AppID)
	require.NoError(t, err)

	// Create an OAuth2 permission grant with group as principal
	grant, err := store.CreateOAuth2PermissionGrant(ctx, model.OAuth2PermissionGrant{
		ClientID:    sp.ID,
		ConsentType: "Principal",
		PrincipalID: group.ID,
		ResourceID:  sp.ID,
		Scope:       "User.Read",
	})
	require.NoError(t, err)

	// Delete group
	err = store.DeleteGroup(ctx, group.ID)
	require.NoError(t, err)

	// Verify grant is gone
	_, err = store.GetOAuth2PermissionGrant(ctx, grant.ID)
	assert.ErrorIs(t, err, ErrGrantNotFound)

	assert.NotContains(t, store.oauth2PermissionGrants, grant.ID)
}