package store

import (
	"context"
	"fmt"
	"testing"

	"github.com/saldeti/saldeti/internal/google/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStore_ClientAuth(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	// Register a client
	err := s.RegisterClient(ctx, "client-1", "secret-1")
	require.NoError(t, err)

	// Get the client
	secret, err := s.GetClient(ctx, "client-1")
	require.NoError(t, err)
	assert.Equal(t, "secret-1", secret)

	// Duplicate registration returns ErrAlreadyExists
	err = s.RegisterClient(ctx, "client-1", "secret-2")
	assert.ErrorIs(t, err, ErrAlreadyExists)

	// Non-existent client returns ErrNotFound
	_, err = s.GetClient(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryStore_UserCRUD(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	// Create
	user := model.User{
		PrimaryEmail: "alice@example.com",
		GivenName:    "Alice",
		FamilyName:   "Smith",
		DisplayName:  "Alice Smith",
	}
	created, err := s.CreateUser(ctx, user)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "admin#directory#user", created.Kind)
	assert.NotEmpty(t, created.CreationTime)
	assert.Equal(t, "alice@example.com", created.PrimaryEmail)

	// Get by ID
	got, err := s.GetUser(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "alice@example.com", got.PrimaryEmail)

	// Get by email
	gotByEmail, err := s.GetUser(ctx, "alice@example.com")
	require.NoError(t, err)
	assert.Equal(t, created.ID, gotByEmail.ID)

	// Duplicate create returns ErrAlreadyExists
	_, err = s.CreateUser(ctx, model.User{PrimaryEmail: "alice@example.com"})
	assert.ErrorIs(t, err, ErrAlreadyExists)

	// Update
	updated, err := s.UpdateUser(ctx, created.ID, model.User{
		PrimaryEmail: "alice.updated@example.com",
		GivenName:    "Alice",
		FamilyName:   "Updated",
		DisplayName:  "Alice Updated",
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "Alice Updated", updated.DisplayName)

	// Get by new email
	gotNew, err := s.GetUser(ctx, "alice.updated@example.com")
	require.NoError(t, err)
	assert.Equal(t, created.ID, gotNew.ID)

	// Patch
	patched, err := s.PatchUser(ctx, created.ID, map[string]any{
		"name": map[string]any{
			"givenName":  "Alice",
			"familyName": "Patched",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, patched.Name)
	assert.Equal(t, "Patched", patched.Name.FamilyName)

	// List
	users, _, err := s.ListUsers(ctx, model.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, users, 1)

	// Delete
	err = s.DeleteUser(ctx, created.ID)
	require.NoError(t, err)

	// Verify gone
	_, err = s.GetUser(ctx, created.ID)
	assert.ErrorIs(t, err, ErrNotFound)

	_, err = s.GetUser(ctx, "alice.updated@example.com")
	assert.ErrorIs(t, err, ErrNotFound)

	// Delete non-existent returns ErrNotFound
	err = s.DeleteUser(ctx, created.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryStore_UserAliases(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	user, err := s.CreateUser(ctx, model.User{PrimaryEmail: "alias-test@example.com"})
	require.NoError(t, err)

	// Add alias
	err = s.AddUserAlias(ctx, user.ID, "alias1@example.com")
	require.NoError(t, err)

	// List aliases
	aliases, err := s.ListUserAliases(ctx, user.ID)
	require.NoError(t, err)
	assert.Contains(t, aliases, "alias1@example.com")

	// Add second alias
	err = s.AddUserAlias(ctx, user.ID, "alias2@example.com")
	require.NoError(t, err)

	aliases, err = s.ListUserAliases(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, aliases, 2)

	// Remove alias
	err = s.RemoveUserAlias(ctx, user.ID, "alias1@example.com")
	require.NoError(t, err)

	aliases, err = s.ListUserAliases(ctx, user.ID)
	require.NoError(t, err)
	assert.NotContains(t, aliases, "alias1@example.com")
	assert.Contains(t, aliases, "alias2@example.com")

	// Alias on non-existent user returns ErrNotFound
	err = s.AddUserAlias(ctx, "nonexistent", "nope@example.com")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryStore_UserPhotos(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	user, err := s.CreateUser(ctx, model.User{PrimaryEmail: "photo-test@example.com"})
	require.NoError(t, err)

	// No photo initially
	_, err = s.GetUserPhoto(ctx, user.ID)
	assert.ErrorIs(t, err, ErrNotFound)

	// Update photo
	photo := model.UserPhoto{
		PhotoData: "iVBORw0KGgo=",
		MimeType:  "image/png",
		Height:    96,
		Width:     96,
	}
	err = s.UpdateUserPhoto(ctx, user.ID, photo)
	require.NoError(t, err)

	// Get photo
	gotPhoto, err := s.GetUserPhoto(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "iVBORw0KGgo=", gotPhoto.PhotoData)
	assert.Equal(t, "image/png", gotPhoto.MimeType)
	assert.Equal(t, "photo-test@example.com", gotPhoto.PrimaryEmail)

	// Get photo by email
	gotPhotoByEmail, err := s.GetUserPhoto(ctx, "photo-test@example.com")
	require.NoError(t, err)
	assert.Equal(t, "iVBORw0KGgo=", gotPhotoByEmail.PhotoData)

	// Delete photo
	err = s.DeleteUserPhoto(ctx, user.ID)
	require.NoError(t, err)

	_, err = s.GetUserPhoto(ctx, user.ID)
	assert.ErrorIs(t, err, ErrNotFound)

	// Photo on non-existent user returns ErrNotFound
	err = s.UpdateUserPhoto(ctx, "nonexistent", photo)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryStore_UserAdmin(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	user, err := s.CreateUser(ctx, model.User{PrimaryEmail: "admin-test@example.com"})
	require.NoError(t, err)

	// Initially not admin
	got, err := s.GetUser(ctx, user.ID)
	require.NoError(t, err)
	assert.False(t, got.IsAdmin)

	// Make admin
	err = s.MakeAdmin(ctx, user.ID, true)
	require.NoError(t, err)
	got, err = s.GetUser(ctx, user.ID)
	require.NoError(t, err)
	assert.True(t, got.IsAdmin)

	// Revoke admin
	err = s.MakeAdmin(ctx, user.ID, false)
	require.NoError(t, err)
	got, err = s.GetUser(ctx, user.ID)
	require.NoError(t, err)
	assert.False(t, got.IsAdmin)

	// Make admin by email
	err = s.MakeAdmin(ctx, "admin-test@example.com", true)
	require.NoError(t, err)
	got, err = s.GetUser(ctx, user.ID)
	require.NoError(t, err)
	assert.True(t, got.IsAdmin)

	// Non-existent user
	err = s.MakeAdmin(ctx, "nonexistent", true)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryStore_GroupCRUD(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	// Create
	group := model.Group{
		Email:       "eng@example.com",
		Name:        "Engineering",
		Description: "Engineering team",
	}
	created, err := s.CreateGroup(ctx, group)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "admin#directory#group", created.Kind)
	assert.Equal(t, "eng@example.com", created.Email)

	// Get by ID
	got, err := s.GetGroup(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	// Get by email
	gotByEmail, err := s.GetGroup(ctx, "eng@example.com")
	require.NoError(t, err)
	assert.Equal(t, created.ID, gotByEmail.ID)

	// Duplicate create returns ErrAlreadyExists
	_, err = s.CreateGroup(ctx, model.Group{Email: "eng@example.com"})
	assert.ErrorIs(t, err, ErrAlreadyExists)

	// Update
	updated, err := s.UpdateGroup(ctx, created.ID, model.Group{
		Email:       "engineering@example.com",
		Name:        "Engineering Updated",
		Description: "Updated description",
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "Engineering Updated", updated.Name)

	// Get by new email
	gotNew, err := s.GetGroup(ctx, "engineering@example.com")
	require.NoError(t, err)
	assert.Equal(t, created.ID, gotNew.ID)

	// Patch
	patched, err := s.PatchGroup(ctx, created.ID, map[string]any{
		"description": "Patched description",
	})
	require.NoError(t, err)
	assert.Equal(t, "Patched description", patched.Description)

	// List
	groups, _, err := s.ListGroups(ctx, model.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, groups, 1)

	// Delete
	err = s.DeleteGroup(ctx, created.ID)
	require.NoError(t, err)

	_, err = s.GetGroup(ctx, created.ID)
	assert.ErrorIs(t, err, ErrNotFound)

	// Delete non-existent returns ErrNotFound
	err = s.DeleteGroup(ctx, created.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryStore_MemberCRUD(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	// Create a group first
	group, err := s.CreateGroup(ctx, model.Group{Email: "team@example.com"})
	require.NoError(t, err)

	// Add member
	member := model.Member{
		Email: "member1@example.com",
		Role:  "MEMBER",
		Type:  "USER",
	}
	added, err := s.AddMember(ctx, group.ID, member)
	require.NoError(t, err)
	assert.NotEmpty(t, added.ID)
	assert.Equal(t, "admin#directory#member", added.Kind)
	assert.Equal(t, "member1@example.com", added.Email)

	// Get member
	got, err := s.GetMember(ctx, group.ID, "member1@example.com")
	require.NoError(t, err)
	assert.Equal(t, added.ID, got.ID)

	// List members
	members, _, err := s.ListMembers(ctx, group.ID, model.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, members, 1)

	// HasMember
	has, err := s.HasMember(ctx, group.ID, "member1@example.com")
	require.NoError(t, err)
	assert.True(t, has)

	has, err = s.HasMember(ctx, group.ID, "nonexistent@example.com")
	require.NoError(t, err)
	assert.False(t, has)

	// Update member
	updated, err := s.UpdateMember(ctx, group.ID, "member1@example.com", model.Member{
		Email: "member1-updated@example.com",
		Role:  "MANAGER",
		Type:  "USER",
	})
	require.NoError(t, err)
	assert.Equal(t, added.ID, updated.ID)
	assert.Equal(t, "MANAGER", updated.Role)

	// Old email key no longer exists
	_, err = s.GetMember(ctx, group.ID, "member1@example.com")
	assert.ErrorIs(t, err, ErrNotFound)

	// New email key works
	gotNew, err := s.GetMember(ctx, group.ID, "member1-updated@example.com")
	require.NoError(t, err)
	assert.Equal(t, "MANAGER", gotNew.Role)

	// Remove member
	err = s.RemoveMember(ctx, group.ID, "member1-updated@example.com")
	require.NoError(t, err)

	_, err = s.GetMember(ctx, group.ID, "member1-updated@example.com")
	assert.ErrorIs(t, err, ErrNotFound)

	// Remove non-existent returns ErrNotFound
	err = s.RemoveMember(ctx, group.ID, "nonexistent@example.com")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryStore_OrgUnitCRUD(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	customerID := "C12345"

	// Create
	ou := model.OrgUnit{
		Name:        "Engineering",
		OrgUnitPath: "/engineering",
	}
	created, err := s.CreateOrgUnit(ctx, customerID, ou)
	require.NoError(t, err)
	assert.NotEmpty(t, created.OrgUnitId)
	assert.Equal(t, "admin#directory#orgUnit", created.Kind)
	assert.Equal(t, "/engineering", created.OrgUnitPath)

	// Get
	got, err := s.GetOrgUnit(ctx, customerID, "/engineering")
	require.NoError(t, err)
	assert.Equal(t, created.OrgUnitId, got.OrgUnitId)
	assert.Equal(t, "Engineering", got.Name)

	// List
	ous, err := s.ListOrgUnits(ctx, customerID)
	require.NoError(t, err)
	assert.Len(t, ous, 1)

	// List empty customer
	emptyOUs, err := s.ListOrgUnits(ctx, "C_NONEXISTENT")
	require.NoError(t, err)
	assert.Len(t, emptyOUs, 0)

	// Update
	updated, err := s.UpdateOrgUnit(ctx, customerID, "/engineering", model.OrgUnit{
		Name:        "Engineering Updated",
		OrgUnitPath: "/eng-updated",
	})
	require.NoError(t, err)
	assert.Equal(t, created.OrgUnitId, updated.OrgUnitId)
	assert.Equal(t, "Engineering Updated", updated.Name)

	// Old path is gone
	_, err = s.GetOrgUnit(ctx, customerID, "/engineering")
	assert.ErrorIs(t, err, ErrNotFound)

	// New path works
	gotNew, err := s.GetOrgUnit(ctx, customerID, "/eng-updated")
	require.NoError(t, err)
	assert.Equal(t, "Engineering Updated", gotNew.Name)

	// Patch
	patched, err := s.PatchOrgUnit(ctx, customerID, "/eng-updated", map[string]any{
		"name": "Patched OU",
	})
	require.NoError(t, err)
	assert.Equal(t, "Patched OU", patched.Name)

	// Delete
	err = s.DeleteOrgUnit(ctx, customerID, "/eng-updated")
	require.NoError(t, err)

	_, err = s.GetOrgUnit(ctx, customerID, "/eng-updated")
	assert.ErrorIs(t, err, ErrNotFound)

	// Delete non-existent returns ErrNotFound
	err = s.DeleteOrgUnit(ctx, customerID, "/eng-updated")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryStore_RoleCRUD(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	customerID := "C12345"

	// Create
	role := model.Role{
		RoleName:        "TestRole",
		RoleDescription: "A test role",
		RolePrivileges: []model.RolePrivilege{
			{PrivilegeName: "USERS_MANAGE", ServiceId: "serviceAdmin"},
		},
	}
	created, err := s.CreateRole(ctx, customerID, role)
	require.NoError(t, err)
	assert.NotEmpty(t, created.RoleId)
	assert.Equal(t, "TestRole", created.RoleName)

	// Get
	got, err := s.GetRole(ctx, customerID, created.RoleId)
	require.NoError(t, err)
	assert.Equal(t, created.RoleId, got.RoleId)
	assert.Equal(t, "TestRole", got.RoleName)

	// List
	roles, err := s.ListRoles(ctx, customerID)
	require.NoError(t, err)
	assert.Len(t, roles, 1)

	// Update
	updated, err := s.UpdateRole(ctx, customerID, created.RoleId, model.Role{
		RoleName:        "UpdatedRole",
		RoleDescription: "Updated description",
	})
	require.NoError(t, err)
	assert.Equal(t, created.RoleId, updated.RoleId)
	assert.Equal(t, "UpdatedRole", updated.RoleName)

	// Patch
	patched, err := s.PatchRole(ctx, customerID, created.RoleId, map[string]any{
		"roleDescription": "Patched description",
	})
	require.NoError(t, err)
	assert.Equal(t, "Patched description", patched.RoleDescription)

	// Delete
	err = s.DeleteRole(ctx, customerID, created.RoleId)
	require.NoError(t, err)

	_, err = s.GetRole(ctx, customerID, created.RoleId)
	assert.ErrorIs(t, err, ErrNotFound)

	// Delete non-existent returns ErrNotFound
	err = s.DeleteRole(ctx, customerID, created.RoleId)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryStore_RoleAssignmentCRUD(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	customerID := "C12345"

	// Create a role first
	role, err := s.CreateRole(ctx, customerID, model.Role{
		RoleName: "TestRole",
	})
	require.NoError(t, err)

	// Create a user to assign to
	user, err := s.CreateUser(ctx, model.User{PrimaryEmail: "assignee@example.com"})
	require.NoError(t, err)

	// Create role assignment
	ra := model.RoleAssignment{
		RoleId:     role.RoleId,
		AssignedTo: user.ID,
		ScopeType:  "CUSTOMER",
	}
	created, err := s.CreateRoleAssignment(ctx, customerID, ra)
	require.NoError(t, err)
	assert.NotEmpty(t, created.RoleAssignmentId)
	assert.Equal(t, "admin#directory#roleAssignment", created.Kind)
	assert.Equal(t, role.RoleId, created.RoleId)
	assert.Equal(t, user.ID, created.AssignedTo)

	// Get
	got, err := s.GetRoleAssignment(ctx, customerID, created.RoleAssignmentId)
	require.NoError(t, err)
	assert.Equal(t, created.RoleAssignmentId, got.RoleAssignmentId)

	// List
	assignments, err := s.ListRoleAssignments(ctx, customerID)
	require.NoError(t, err)
	assert.Len(t, assignments, 1)

	// Delete
	err = s.DeleteRoleAssignment(ctx, customerID, created.RoleAssignmentId)
	require.NoError(t, err)

	_, err = s.GetRoleAssignment(ctx, customerID, created.RoleAssignmentId)
	assert.ErrorIs(t, err, ErrNotFound)

	// Delete non-existent returns ErrNotFound
	err = s.DeleteRoleAssignment(ctx, customerID, created.RoleAssignmentId)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryStore_DomainCRUD(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	customerID := "C12345"

	// Add domain
	domain := model.Domain{
		DomainName: "example.com",
		IsPrimary:  true,
	}
	created, err := s.AddDomain(ctx, customerID, domain)
	require.NoError(t, err)
	assert.Equal(t, "admin#directory#domain", created.Kind)
	assert.NotEmpty(t, created.CreationTime)
	assert.Equal(t, "example.com", created.DomainName)

	// Get
	got, err := s.GetDomain(ctx, customerID, "example.com")
	require.NoError(t, err)
	assert.Equal(t, "example.com", got.DomainName)
	assert.True(t, got.IsPrimary)

	// List
	domains, err := s.ListDomains(ctx, customerID)
	require.NoError(t, err)
	assert.Len(t, domains, 1)

	// Add second domain
	_, err = s.AddDomain(ctx, customerID, model.Domain{
		DomainName: "example.org",
	})
	require.NoError(t, err)

	domains, err = s.ListDomains(ctx, customerID)
	require.NoError(t, err)
	assert.Len(t, domains, 2)

	// Delete
	err = s.DeleteDomain(ctx, customerID, "example.com")
	require.NoError(t, err)

	_, err = s.GetDomain(ctx, customerID, "example.com")
	assert.ErrorIs(t, err, ErrNotFound)

	// Second domain still exists
	got2, err := s.GetDomain(ctx, customerID, "example.org")
	require.NoError(t, err)
	assert.Equal(t, "example.org", got2.DomainName)

	// Delete non-existent returns ErrNotFound
	err = s.DeleteDomain(ctx, customerID, "nonexistent.com")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryStore_Pagination(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	// Create 25 users
	for i := 0; i < 25; i++ {
		_, err := s.CreateUser(ctx, model.User{
			PrimaryEmail: fmt.Sprintf("user-%d@test.com", i),
			DisplayName:  fmt.Sprintf("User %d", i),
		})
		require.NoError(t, err)
	}

	// Page 1: MaxResults=10
	page1, nextToken1, err := s.ListUsers(ctx, model.ListOptions{MaxResults: 10})
	require.NoError(t, err)
	assert.Len(t, page1, 10)
	assert.NotEmpty(t, nextToken1)

	// Page 2: use nextToken1
	page2, nextToken2, err := s.ListUsers(ctx, model.ListOptions{
		MaxResults: 10,
		PageToken:  nextToken1,
	})
	require.NoError(t, err)
	assert.Len(t, page2, 10)
	assert.NotEmpty(t, nextToken2)

	// Page 3: use nextToken2
	page3, nextToken3, err := s.ListUsers(ctx, model.ListOptions{
		MaxResults: 10,
		PageToken:  nextToken2,
	})
	require.NoError(t, err)
	assert.Len(t, page3, 5)
	assert.Empty(t, nextToken3)

	// Total across all pages
	total := len(page1) + len(page2) + len(page3)
	assert.Equal(t, 25, total)
}

func TestMemoryStore_PatchUserPreservesJsonIgnoreFields(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	created, err := s.CreateUser(ctx, model.User{
		PrimaryEmail: "bob@example.com",
		GivenName:    "Bob",
		FamilyName:   "Jones",
		DisplayName:  "Bob Jones",
	})
	require.NoError(t, err)

	// Patch an unrelated field; the json:"-" name fields must survive the round-trip.
	patched, err := s.PatchUser(ctx, created.ID, map[string]any{
		"orgUnitPath": "/Engineering",
	})
	require.NoError(t, err)
	assert.Equal(t, "Bob", patched.GivenName, `GivenName (json:"-") must survive patch`)
	assert.Equal(t, "Jones", patched.FamilyName, `FamilyName (json:"-") must survive patch`)
	assert.Equal(t, "Bob Jones", patched.DisplayName, `DisplayName (json:"-") must survive patch`)
	assert.Equal(t, "/Engineering", patched.OrgUnitPath)
}
