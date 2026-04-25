package store

import (
	"context"
	"github.com/saldeti/saldeti/internal/model"
)

type Store interface {
	// Users
	GetUser(ctx context.Context, id string) (*model.User, error)
	GetUserByUPN(ctx context.Context, upn string) (*model.User, error)
	CreateUser(ctx context.Context, user model.User) (model.User, error)
	ListUsers(ctx context.Context, opts model.ListOptions) ([]model.User, int, error)
	UpdateUser(ctx context.Context, id string, patch map[string]interface{}) (*model.User, error)
	DeleteUser(ctx context.Context, id string) error

	// Tenant/App registration for auth
	GetClient(ctx context.Context, clientID string) (clientID_returned string, clientSecret string, tenantID string, err error)
	RegisterClient(ctx context.Context, clientID, clientSecret, tenantID string) error
	ListClients(ctx context.Context) ([]Client, error)

	// Groups
	ListGroups(ctx context.Context, opts model.ListOptions) ([]model.Group, int, error)
	GetGroup(ctx context.Context, id string) (*model.Group, error)
	CreateGroup(ctx context.Context, group model.Group) (model.Group, error)
	UpdateGroup(ctx context.Context, id string, patch map[string]interface{}) (*model.Group, error)
	DeleteGroup(ctx context.Context, id string) error

	// Membership
	AddMember(ctx context.Context, groupID, objectID, objectType string) error
	RemoveMember(ctx context.Context, groupID, objectID string) error
	ListMembers(ctx context.Context, groupID string, opts model.ListOptions) ([]model.DirectoryObject, int, error)
	ListTransitiveMembers(ctx context.Context, groupID string, opts model.ListOptions) ([]model.DirectoryObject, int, error)
	AddOwner(ctx context.Context, groupID, objectID, objectType string) error
	RemoveOwner(ctx context.Context, groupID, objectID string) error
	ListOwners(ctx context.Context, groupID string, opts model.ListOptions) ([]model.DirectoryObject, int, error)

	// Group navigation
	ListGroupMemberOf(ctx context.Context, groupID string, opts model.ListOptions) ([]model.DirectoryObject, int, error)
	ListGroupTransitiveMemberOf(ctx context.Context, groupID string, opts model.ListOptions) ([]model.DirectoryObject, int, error)

	// Group membership actions
	CheckMemberGroups(ctx context.Context, objectID string, groupIDs []string) ([]string, error)
	GetMemberGroups(ctx context.Context, objectID string, securityEnabledOnly bool) ([]string, error)

	// Object resolution
	ResolveObjectType(ctx context.Context, objectID string) (string, error)

	// User navigation properties
	GetManager(ctx context.Context, userID string) (*model.DirectoryObject, error)
	SetManager(ctx context.Context, userID, managerID string) error
	RemoveManager(ctx context.Context, userID string) error
	ListDirectReports(ctx context.Context, userID string, opts model.ListOptions) ([]model.DirectoryObject, int, error)
	ListUserMemberOf(ctx context.Context, userID string, opts model.ListOptions) ([]model.DirectoryObject, int, error)
	ListUserTransitiveMemberOf(ctx context.Context, userID string, opts model.ListOptions) ([]model.DirectoryObject, int, error)

	// Directory object operations
	GetDirectoryObjects(ctx context.Context, ids []string, types []string) ([]map[string]interface{}, error)

	// Delta query
	GetUsersDelta(ctx context.Context, deltaToken string) ([]map[string]interface{}, string, int, error)
	GetGroupsDelta(ctx context.Context, deltaToken string) ([]map[string]interface{}, string, int, error)

	// Licenses
	ListSubscribedSkus(ctx context.Context) ([]model.SubscribedSku, error)
	AssignLicense(ctx context.Context, userID string, addLicenses []model.LicenseAssignment, removeLicenses []model.LicenseRemoval) (*model.User, error)

	// Applications
	ListApplications(ctx context.Context, opts model.ListOptions) ([]model.Application, int, error)
	GetApplication(ctx context.Context, id string) (*model.Application, error)
	GetApplicationByAppID(ctx context.Context, appId string) (*model.Application, error)
	CreateApplication(ctx context.Context, app model.Application) (model.Application, error)
	UpdateApplication(ctx context.Context, id string, patch map[string]interface{}) (*model.Application, error)
	DeleteApplication(ctx context.Context, id string) error
	AddApplicationPassword(ctx context.Context, appID string, cred model.PasswordCredential) (model.PasswordCredential, error)
	RemoveApplicationPassword(ctx context.Context, appID, keyID string) error
	AddApplicationKey(ctx context.Context, appID string, cred model.KeyCredential) (model.KeyCredential, error)
	RemoveApplicationKey(ctx context.Context, appID, keyID string) error
	ListApplicationOwners(ctx context.Context, appID string, opts model.ListOptions) ([]model.DirectoryObject, int, error)
	AddApplicationOwner(ctx context.Context, appID, objectID, objectType string) error
	RemoveApplicationOwner(ctx context.Context, appID, objectID string) error
	ListExtensionProperties(ctx context.Context, appID string) ([]model.ExtensionProperty, error)
	CreateExtensionProperty(ctx context.Context, appID string, ep model.ExtensionProperty) (model.ExtensionProperty, error)
	DeleteExtensionProperty(ctx context.Context, appID, extID string) error
	GetApplicationsDelta(ctx context.Context, deltaToken string) ([]map[string]interface{}, string, int, error)

	// Service Principals (expanded)
	ListServicePrincipals(ctx context.Context, opts model.ListOptions) ([]model.ServicePrincipal, int, error)
	GetServicePrincipal(ctx context.Context, id string) (*model.ServicePrincipal, error)
	GetServicePrincipalByAppID(ctx context.Context, appId string) (*model.ServicePrincipal, error)
	CreateServicePrincipal(ctx context.Context, sp model.ServicePrincipal) (model.ServicePrincipal, error)
	UpdateServicePrincipal(ctx context.Context, id string, patch map[string]interface{}) (*model.ServicePrincipal, error)
	DeleteServicePrincipal(ctx context.Context, id string) error
	ListSPOwners(ctx context.Context, spID string, opts model.ListOptions) ([]model.DirectoryObject, int, error)
	AddSPOwner(ctx context.Context, spID, objectID, objectType string) error
	RemoveSPOwner(ctx context.Context, spID, objectID string) error
	ListSPMemberOf(ctx context.Context, spID string, opts model.ListOptions) ([]model.DirectoryObject, int, error)
	ListSPTransitiveMemberOf(ctx context.Context, spID string, opts model.ListOptions) ([]model.DirectoryObject, int, error)
	UpdateSPCredentials(ctx context.Context, spID string, update func(*model.ServicePrincipal) error) error

	// App Role Assignments
	CreateAppRoleAssignment(ctx context.Context, resourceID, principalID, appRoleID string) (model.AppRoleAssignment, error)
	ListAppRoleAssignments(ctx context.Context, principalID string, opts model.ListOptions) ([]model.AppRoleAssignment, int, error)
	ListAppRoleAssignedTo(ctx context.Context, resourceID string, opts model.ListOptions) ([]model.AppRoleAssignment, int, error)
	DeleteAppRoleAssignment(ctx context.Context, assignmentID string) error

	// OAuth2 Permission Grants
	ListOAuth2PermissionGrants(ctx context.Context, opts model.ListOptions) ([]model.OAuth2PermissionGrant, int, error)
	GetOAuth2PermissionGrant(ctx context.Context, id string) (*model.OAuth2PermissionGrant, error)
	CreateOAuth2PermissionGrant(ctx context.Context, grant model.OAuth2PermissionGrant) (model.OAuth2PermissionGrant, error)
	UpdateOAuth2PermissionGrant(ctx context.Context, id string, patch map[string]interface{}) (*model.OAuth2PermissionGrant, error)
	DeleteOAuth2PermissionGrant(ctx context.Context, id string) error
}
