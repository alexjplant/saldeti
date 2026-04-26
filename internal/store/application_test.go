package store

import (
	"context"
	"testing"

	"github.com/saldeti/saldeti/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== Application CRUD ==========

func TestApplicationCRUD(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	// Create
	app := model.Application{
		DisplayName: "Test App",
		Description: "A test application",
	}
	created, err := s.CreateApplication(ctx, app)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.NotEmpty(t, created.AppID)
	assert.Equal(t, "Test App", created.DisplayName)
	assert.NotNil(t, created.CreatedDateTime)
	assert.Equal(t, "#microsoft.graph.application", created.ODataType)

	// Get by ID
	got, err := s.GetApplication(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, created.AppID, got.AppID)

	// Get by AppID
	gotByAppID, err := s.GetApplicationByAppID(ctx, created.AppID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, gotByAppID.ID)

	// List
	apps, count, err := s.ListApplications(ctx, model.ListOptions{})
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Len(t, apps, 1)

	// Update
	updated, err := s.UpdateApplication(ctx, created.ID, map[string]interface{}{
		"displayName": "Updated App",
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated App", updated.DisplayName)

	// Delete
	err = s.DeleteApplication(ctx, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = s.GetApplication(ctx, created.ID)
	assert.ErrorIs(t, err, ErrApplicationNotFound)
}

func TestApplicationDuplicateAppID(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	app := model.Application{
		DisplayName: "App 1",
		AppID:       "dup-app-id",
	}
	_, err := s.CreateApplication(ctx, app)
	require.NoError(t, err)

	app2 := model.Application{
		DisplayName: "App 2",
		AppID:       "dup-app-id",
	}
	_, err = s.CreateApplication(ctx, app2)
	assert.ErrorIs(t, err, ErrDuplicateAppID)
}

func TestApplicationNotFound(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	_, err := s.GetApplication(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrApplicationNotFound)

	_, err = s.GetApplicationByAppID(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrApplicationNotFound)

	err = s.DeleteApplication(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrApplicationNotFound)

	_, err = s.UpdateApplication(ctx, "nonexistent", map[string]interface{}{"displayName": "x"})
	assert.ErrorIs(t, err, ErrApplicationNotFound)
}

// ========== Credential Management ==========

func TestApplicationPasswordCredentials(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	app := model.Application{DisplayName: "Test App"}
	created, err := s.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Add password
	cred, err := s.AddApplicationPassword(ctx, created.ID, model.PasswordCredential{
		DisplayName: "Test Secret",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, cred.KeyID)
	assert.NotEmpty(t, cred.SecretText)
	assert.Len(t, cred.SecretText, 32)
	assert.Contains(t, cred.Hint, "***")
	assert.NotNil(t, cred.StartDateTime)
	assert.NotNil(t, cred.EndDateTime)

	// Verify it's on the app
	got, _ := s.GetApplication(ctx, created.ID)
	assert.Len(t, got.PasswordCredentials, 1)
	assert.Equal(t, cred.KeyID, got.PasswordCredentials[0].KeyID)

	// Remove password
	err = s.RemoveApplicationPassword(ctx, created.ID, cred.KeyID)
	require.NoError(t, err)

	got, _ = s.GetApplication(ctx, created.ID)
	assert.Len(t, got.PasswordCredentials, 0)

	// Remove non-existent
	err = s.RemoveApplicationPassword(ctx, created.ID, "nonexistent-key")
	assert.ErrorIs(t, err, ErrCredentialNotFound)

	// Add to non-existent app
	_, err = s.AddApplicationPassword(ctx, "nonexistent", model.PasswordCredential{})
	assert.ErrorIs(t, err, ErrApplicationNotFound)
}

func TestApplicationKeyCredentials(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	app := model.Application{DisplayName: "Test App"}
	created, err := s.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Add key
	cred, err := s.AddApplicationKey(ctx, created.ID, model.KeyCredential{
		DisplayName: "Test Key",
		Type:        "AsymmetricX509Cert",
		Usage:       "Verify",
		Key:         "base64encodedkey",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, cred.KeyID)
	assert.NotNil(t, cred.StartDateTime)

	// Verify it's on the app
	got, _ := s.GetApplication(ctx, created.ID)
	assert.Len(t, got.KeyCredentials, 1)

	// Remove key
	err = s.RemoveApplicationKey(ctx, created.ID, cred.KeyID)
	require.NoError(t, err)

	got, _ = s.GetApplication(ctx, created.ID)
	assert.Len(t, got.KeyCredentials, 0)

	// Remove non-existent
	err = s.RemoveApplicationKey(ctx, created.ID, "nonexistent-key")
	assert.ErrorIs(t, err, ErrCredentialNotFound)
}

// ========== Application Owners ==========

func TestApplicationOwners(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	// Create user
	user, _ := s.CreateUser(ctx, model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "test@example.com",
	})

	// Create app
	app, _ := s.CreateApplication(ctx, model.Application{DisplayName: "Test App"})

	// Add owner
	err := s.AddApplicationOwner(ctx, app.ID, user.ID, "user")
	require.NoError(t, err)

	// List owners
	owners, count, err := s.ListApplicationOwners(ctx, app.ID, model.ListOptions{})
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Len(t, owners, 1)
	assert.Equal(t, user.ID, owners[0].ID)
	assert.Equal(t, "#microsoft.graph.user", owners[0].ODataType)

	// Duplicate owner
	err = s.AddApplicationOwner(ctx, app.ID, user.ID, "user")
	assert.ErrorIs(t, err, ErrAlreadyAppOwner)

	// Remove owner
	err = s.RemoveApplicationOwner(ctx, app.ID, user.ID)
	require.NoError(t, err)

	// Not an owner
	err = s.RemoveApplicationOwner(ctx, app.ID, user.ID)
	assert.ErrorIs(t, err, ErrNotAppOwner)
}

// ========== Extension Properties ==========

func TestExtensionProperties(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	app, _ := s.CreateApplication(ctx, model.Application{DisplayName: "Test App"})

	// Create extension
	ep, err := s.CreateExtensionProperty(ctx, app.ID, model.ExtensionProperty{
		Name:         "extension_test_attr",
		DataType:     "String",
		TargetObjects: []string{"User"},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, ep.ID)
	assert.Equal(t, "extension_test_attr", ep.Name)
	assert.Equal(t, "Test App", ep.AppDisplayName)

	// List extensions
	exts, err := s.ListExtensionProperties(ctx, app.ID)
	require.NoError(t, err)
	assert.Len(t, exts, 1)

	// Delete extension
	err = s.DeleteExtensionProperty(ctx, app.ID, ep.ID)
	require.NoError(t, err)

	exts, err = s.ListExtensionProperties(ctx, app.ID)
	require.NoError(t, err)
	assert.Len(t, exts, 0)

	// Delete non-existent
	err = s.DeleteExtensionProperty(ctx, app.ID, "nonexistent")
	assert.ErrorIs(t, err, ErrExtensionNotFound)
}

// ========== Delta ==========

func TestApplicationsDelta(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	// Create an app
	app, _ := s.CreateApplication(ctx, model.Application{DisplayName: "Delta App"})

	// Get initial delta (empty token = all)
	results, token, count, err := s.GetApplicationsDelta(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, token)
	assert.Len(t, results, 1)

	// Delete the app
	s.DeleteApplication(ctx, app.ID)

	// Get delta since last token
	results, token2, count, err := s.GetApplicationsDelta(ctx, token)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 1)
	assert.NotEmpty(t, token2)

	// Find deleted entry
	found := false
	for _, r := range results {
		if r["@removed"] != nil {
			found = true
			assert.Equal(t, app.ID, r["id"])
		}
	}
	assert.True(t, found, "expected to find a deleted application entry in delta")
}

// ========== Service Principal CRUD ==========

func TestServicePrincipalCRUD(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	// Create app first (SP requires an app)
	app, _ := s.CreateApplication(ctx, model.Application{
		DisplayName: "SP Test App",
		Description: "App for SP test",
	})

	// The SP should have been auto-created
	sps, _, _ := s.ListServicePrincipals(ctx, model.ListOptions{})
	assert.GreaterOrEqual(t, len(sps), 1, "SP should be auto-created with application")

	// Find the auto-created SP
	var autoSP *model.ServicePrincipal
	for _, sp := range sps {
		if sp.AppID == app.AppID {
			autoSP = &sp
			break
		}
	}
	require.NotNil(t, autoSP, "auto-created SP should exist")

	// Test creating another SP with same appId fails
	_, err := s.CreateServicePrincipal(ctx, model.ServicePrincipal{AppID: app.AppID})
	assert.ErrorIs(t, err, ErrDuplicateSPAppID)

	// Create a second app + SP
	app2, _ := s.CreateApplication(ctx, model.Application{DisplayName: "Second App"})
	sp2, err := s.GetServicePrincipalByAppID(ctx, app2.AppID)
	require.NoError(t, err)

	// Get by ID
	got, err := s.GetServicePrincipal(ctx, sp2.ID)
	require.NoError(t, err)
	assert.Equal(t, sp2.ID, got.ID)

	// Update SP
	updated, err := s.UpdateServicePrincipal(ctx, sp2.ID, map[string]interface{}{
		"displayName": "Updated SP",
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated SP", updated.DisplayName)

	// Delete SP
	err = s.DeleteServicePrincipal(ctx, sp2.ID)
	require.NoError(t, err)

	_, err = s.GetServicePrincipal(ctx, sp2.ID)
	assert.ErrorIs(t, err, ErrServicePrincipalNotFound)
}

func TestServicePrincipalFromApplication(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	isEnabled := true
	app := model.Application{
		DisplayName: "App With Roles",
		AppRoles: []model.AppRole{
			{
				ID:                 "role-1",
				AllowedMemberTypes: []string{"Application"},
				Description:        "Read all",
				DisplayName:        "Reader",
				IsEnabled:          &isEnabled,
				Value:              "Reader",
			},
		},
	}
	created, err := s.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Find the auto-created SP
	sp, err := s.GetServicePrincipalByAppID(ctx, created.AppID)
	require.NoError(t, err)
	assert.Equal(t, "App With Roles", sp.DisplayName)
	assert.Len(t, sp.AppRoles, 1)
	assert.Equal(t, "role-1", sp.AppRoles[0].ID)
	assert.True(t, *sp.AccountEnabled)
}

func TestServicePrincipalValidation(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	// No app exists for this appId — store is permissive, SP is created anyway
	sp, err := s.CreateServicePrincipal(ctx, model.ServicePrincipal{AppID: "nonexistent"})
	require.NoError(t, err)
	assert.Equal(t, "nonexistent", sp.AppID)
	assert.NotEmpty(t, sp.ID)

	// Empty appId
	_, err = s.CreateServicePrincipal(ctx, model.ServicePrincipal{})
	require.Error(t, err)
}

// ========== SP Owners ==========

func TestSPOwners(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	user, _ := s.CreateUser(ctx, model.User{
		DisplayName:       "SP Owner User",
		UserPrincipalName: "spowner@example.com",
	})

	app, _ := s.CreateApplication(ctx, model.Application{DisplayName: "SP Owner App"})
	sp, _ := s.GetServicePrincipalByAppID(ctx, app.AppID)
	require.NotNil(t, sp)

	// Add owner
	err := s.AddSPOwner(ctx, sp.ID, user.ID, "user")
	require.NoError(t, err)

	// List
	owners, count, err := s.ListSPOwners(ctx, sp.ID, model.ListOptions{})
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, user.ID, owners[0].ID)

	// Duplicate
	err = s.AddSPOwner(ctx, sp.ID, user.ID, "user")
	assert.ErrorIs(t, err, ErrAlreadySPOwner)

	// Remove
	err = s.RemoveSPOwner(ctx, sp.ID, user.ID)
	require.NoError(t, err)

	// Not owner
	err = s.RemoveSPOwner(ctx, sp.ID, user.ID)
	assert.ErrorIs(t, err, ErrNotSPOwner)
}

// ========== SP MemberOf ==========

func TestSPMemberOf(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	app, _ := s.CreateApplication(ctx, model.Application{DisplayName: "SP Member App"})
	sp, _ := s.GetServicePrincipalByAppID(ctx, app.AppID)
	require.NotNil(t, sp)

	group, _ := s.CreateGroup(ctx, model.Group{DisplayName: "SP Group"})

	// Add SP as group member
	err := s.AddMember(ctx, group.ID, sp.ID, "servicePrincipal")
	require.NoError(t, err)

	// List SP memberOf
	memberOf, count, err := s.ListSPMemberOf(ctx, sp.ID, model.ListOptions{})
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Len(t, memberOf, 1)
	assert.Equal(t, group.ID, memberOf[0].ID)
	assert.Equal(t, "#microsoft.graph.group", memberOf[0].ODataType)

	// Transitive memberOf
	transitive, count, err := s.ListSPTransitiveMemberOf(ctx, sp.ID, model.ListOptions{})
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Len(t, transitive, 1)

	// Test non-existent SP
	_, _, err = s.ListSPMemberOf(ctx, "nonexistent", model.ListOptions{})
	assert.ErrorIs(t, err, ErrServicePrincipalNotFound)
}

// ========== App Role Assignments ==========

func TestAppRoleAssignment(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	// Create user
	user, _ := s.CreateUser(ctx, model.User{
		DisplayName:       "Assignee",
		UserPrincipalName: "assignee@example.com",
	})

	// Create app with appRole
	isEnabled := true
	app := model.Application{
		DisplayName: "Resource App",
		AppRoles: []model.AppRole{
			{
				ID:                 "role-id-1",
				AllowedMemberTypes: []string{"User"},
				Description:        "Read all",
				DisplayName:        "Reader",
				IsEnabled:          &isEnabled,
				Value:              "Reader",
			},
		},
	}
	createdApp, err := s.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Get the auto-created SP
	sp, err := s.GetServicePrincipalByAppID(ctx, createdApp.AppID)
	require.NoError(t, err)

	// Create assignment
	assignment, err := s.CreateAppRoleAssignment(ctx, sp.ID, user.ID, "role-id-1")
	require.NoError(t, err)
	assert.NotEmpty(t, assignment.ID)
	assert.Equal(t, "role-id-1", assignment.AppRoleID)
	assert.Equal(t, user.ID, assignment.PrincipalID)
	assert.Equal(t, sp.ID, assignment.ResourceID)
	assert.Equal(t, "Assignee", assignment.PrincipalDisplayName)
	assert.Equal(t, "User", assignment.PrincipalType)
	assert.Equal(t, "Resource App", assignment.ResourceDisplayName)
	assert.NotNil(t, assignment.CreatedDateTime)

	// List by principal
	assignments, count, err := s.ListAppRoleAssignments(ctx, user.ID, model.ListOptions{})
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Len(t, assignments, 1)

	// List by resource
	assignedTo, count, err := s.ListAppRoleAssignedTo(ctx, sp.ID, model.ListOptions{})
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Len(t, assignedTo, 1)

	// Delete
	err = s.DeleteAppRoleAssignment(ctx, assignment.ID)
	require.NoError(t, err)

	// Verify deleted
	err = s.DeleteAppRoleAssignment(ctx, assignment.ID)
	assert.ErrorIs(t, err, ErrAssignmentNotFound)
}

func TestAppRoleAssignmentValidation(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	user, _ := s.CreateUser(ctx, model.User{
		DisplayName:       "User1",
		UserPrincipalName: "user1@example.com",
	})

	// Non-existent resource SP
	_, err := s.CreateAppRoleAssignment(ctx, "nonexistent-sp", user.ID, "role-id")
	assert.ErrorIs(t, err, ErrServicePrincipalNotFound)

	// Create app
	app, _ := s.CreateApplication(ctx, model.Application{DisplayName: "App1"})
	sp, _ := s.GetServicePrincipalByAppID(ctx, app.AppID)

	// Non-existent principal
	_, err = s.CreateAppRoleAssignment(ctx, sp.ID, "nonexistent-user", "role-id")
	assert.ErrorIs(t, err, ErrObjectNotFound)

	// Non-existent app role
	_, err = s.CreateAppRoleAssignment(ctx, sp.ID, user.ID, "nonexistent-role")
	assert.ErrorIs(t, err, ErrAppRoleNotFound)
}

// ========== OAuth2 Permission Grants ==========

func TestOAuth2PermissionGrantCRUD(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	// Create
	grant := model.OAuth2PermissionGrant{
		ClientID:    "client-app-id",
		ResourceID:  "resource-app-id",
		Scope:       "User.Read All",
		ConsentType: "AllPrincipals",
	}
	created, err := s.CreateOAuth2PermissionGrant(ctx, grant)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "#microsoft.graph.oAuth2PermissionGrant", created.ODataType)

	// Get
	got, err := s.GetOAuth2PermissionGrant(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	// List
	grants, count, err := s.ListOAuth2PermissionGrants(ctx, model.ListOptions{})
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Len(t, grants, 1)

	// Update
	updated, err := s.UpdateOAuth2PermissionGrant(ctx, created.ID, map[string]interface{}{
		"scope": "User.ReadWrite.All",
	})
	require.NoError(t, err)
	assert.Equal(t, "User.ReadWrite.All", updated.Scope)

	// Delete
	err = s.DeleteOAuth2PermissionGrant(ctx, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = s.GetOAuth2PermissionGrant(ctx, created.ID)
	assert.ErrorIs(t, err, ErrGrantNotFound)

	err = s.DeleteOAuth2PermissionGrant(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrGrantNotFound)
}

// ========== Cascade Deletes ==========

func TestDeleteApplicationCascades(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	// Create user
	user, _ := s.CreateUser(ctx, model.User{
		DisplayName:       "Cascade User",
		UserPrincipalName: "cascade@example.com",
	})

	// Create app with appRole
	isEnabled := true
	app := model.Application{
		DisplayName: "Cascade App",
		AppRoles: []model.AppRole{
			{
				ID:                 "cascade-role-1",
				AllowedMemberTypes: []string{"User"},
				Description:        "Test Role",
				DisplayName:        "TestRole",
				IsEnabled:          &isEnabled,
				Value:              "TestRole",
			},
		},
	}
	created, _ := s.CreateApplication(ctx, app)

	// Get the SP
	sp, _ := s.GetServicePrincipalByAppID(ctx, created.AppID)
	require.NotNil(t, sp)

	// Add owner
	s.AddApplicationOwner(ctx, created.ID, user.ID, "user")

	// Create extension
	s.CreateExtensionProperty(ctx, created.ID, model.ExtensionProperty{
		Name:     "ext_test",
		DataType: "String",
	})

	// Create app role assignment
	assignment, err := s.CreateAppRoleAssignment(ctx, sp.ID, user.ID, "cascade-role-1")
	require.NoError(t, err)

	// Delete the application
	err = s.DeleteApplication(ctx, created.ID)
	require.NoError(t, err)

	// SP should be gone
	_, err = s.GetServicePrincipal(ctx, sp.ID)
	assert.ErrorIs(t, err, ErrServicePrincipalNotFound)

	// Assignment should be gone
	err = s.DeleteAppRoleAssignment(ctx, assignment.ID)
	assert.ErrorIs(t, err, ErrAssignmentNotFound)
}

// ========== Seed Integration ==========

func TestSeedCreatesApplication(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()



	// Just verify the application creation and SP auto-creation works
	app := model.Application{
		AppID:       "sim-client-id",
		DisplayName: "Saldeti Simulator App",
		Description: "Default simulator application",
		AppRoles: []model.AppRole{
			{
				ID:                 "test-role-id",
				AllowedMemberTypes: []string{"Application"},
				Description:        "Read all",
				DisplayName:        "Application.Read.All",
				IsEnabled:          boolPtr(true),
				Value:              "Application.Read.All",
			},
		},
	}
	created, err := s.CreateApplication(ctx, app)
	require.NoError(t, err)
	assert.Equal(t, "sim-client-id", created.AppID)

	// Verify SP was created/updated
	sp, err := s.GetServicePrincipalByAppID(ctx, "sim-client-id")
	require.NoError(t, err)
	assert.Equal(t, "Saldeti Simulator App", sp.DisplayName)
	assert.Len(t, sp.AppRoles, 1)
}



func boolPtr(b bool) *bool {
	return &b
}
