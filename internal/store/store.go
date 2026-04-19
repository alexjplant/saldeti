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
}