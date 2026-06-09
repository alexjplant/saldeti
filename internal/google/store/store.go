package store

import (
	"context"

	"github.com/saldeti/saldeti/internal/google/model"
)

type Store interface {
	// Ping verifies the store is accessible.
	Ping(ctx context.Context) error

	// Auth
	GetClient(ctx context.Context, clientID string) (clientSecret string, err error)
	RegisterClient(ctx context.Context, clientID, clientSecret string) error

	// Users (Tier 1B)
	CreateUser(ctx context.Context, user model.User) (model.User, error)
	GetUser(ctx context.Context, userKey string) (*model.User, error)
	ListUsers(ctx context.Context, opts model.ListOptions) ([]model.User, string, error)
	UpdateUser(ctx context.Context, userKey string, user model.User) (*model.User, error)
	PatchUser(ctx context.Context, userKey string, patch map[string]interface{}) (*model.User, error)
	DeleteUser(ctx context.Context, userKey string) error
	MakeAdmin(ctx context.Context, userKey string, isAdmin bool) error
	UndeleteUser(ctx context.Context, userKey string) (*model.User, error)
	SignOutUser(ctx context.Context, userKey string) error

	// User Aliases
	AddUserAlias(ctx context.Context, userKey string, alias string) error
	ListUserAliases(ctx context.Context, userKey string) ([]string, error)
	RemoveUserAlias(ctx context.Context, userKey string, alias string) error

	// User Photos
	GetUserPhoto(ctx context.Context, userKey string) (*model.UserPhoto, error)
	UpdateUserPhoto(ctx context.Context, userKey string, photo model.UserPhoto) error
	DeleteUserPhoto(ctx context.Context, userKey string) error

	// Groups (Tier 1C)
	CreateGroup(ctx context.Context, group model.Group) (model.Group, error)
	GetGroup(ctx context.Context, groupKey string) (*model.Group, error)
	ListGroups(ctx context.Context, opts model.ListOptions) ([]model.Group, string, error)
	UpdateGroup(ctx context.Context, groupKey string, group model.Group) (*model.Group, error)
	PatchGroup(ctx context.Context, groupKey string, patch map[string]interface{}) (*model.Group, error)
	DeleteGroup(ctx context.Context, groupKey string) error

	// Group Aliases
	AddGroupAlias(ctx context.Context, groupKey string, alias string) error
	ListGroupAliases(ctx context.Context, groupKey string) ([]string, error)
	RemoveGroupAlias(ctx context.Context, groupKey string, alias string) error

	// Members (Tier 1D)
	ListMembers(ctx context.Context, groupKey string, opts model.ListOptions) ([]model.Member, string, error)
	GetMember(ctx context.Context, groupKey, memberKey string) (*model.Member, error)
	AddMember(ctx context.Context, groupKey string, member model.Member) (model.Member, error)
	UpdateMember(ctx context.Context, groupKey, memberKey string, member model.Member) (*model.Member, error)
	RemoveMember(ctx context.Context, groupKey, memberKey string) error
	HasMember(ctx context.Context, groupKey, memberKey string) (bool, error)

	// OrgUnits (Tier 2A)
	ListOrgUnits(ctx context.Context, customerID string) ([]model.OrgUnit, error)
	GetOrgUnit(ctx context.Context, customerID, orgUnitPath string) (*model.OrgUnit, error)
	CreateOrgUnit(ctx context.Context, customerID string, ou model.OrgUnit) (model.OrgUnit, error)
	UpdateOrgUnit(ctx context.Context, customerID, orgUnitPath string, ou model.OrgUnit) (*model.OrgUnit, error)
	PatchOrgUnit(ctx context.Context, customerID, orgUnitPath string, patch map[string]interface{}) (*model.OrgUnit, error)
	DeleteOrgUnit(ctx context.Context, customerID, orgUnitPath string) error

	// Roles (Tier 2B)
	ListRoles(ctx context.Context, customerID string) ([]model.Role, error)
	GetRole(ctx context.Context, customerID, roleID string) (*model.Role, error)
	CreateRole(ctx context.Context, customerID string, role model.Role) (model.Role, error)
	UpdateRole(ctx context.Context, customerID, roleID string, role model.Role) (*model.Role, error)
	PatchRole(ctx context.Context, customerID, roleID string, patch map[string]interface{}) (*model.Role, error)
	DeleteRole(ctx context.Context, customerID, roleID string) error

	// Role Assignments (Tier 2C)
	ListRoleAssignments(ctx context.Context, customerID string) ([]model.RoleAssignment, error)
	GetRoleAssignment(ctx context.Context, customerID, assignmentID string) (*model.RoleAssignment, error)
	CreateRoleAssignment(ctx context.Context, customerID string, ra model.RoleAssignment) (model.RoleAssignment, error)
	DeleteRoleAssignment(ctx context.Context, customerID, assignmentID string) error

	// Privileges (Tier 2D)
	ListPrivileges(ctx context.Context, customerID string) ([]model.Privilege, error)

	// Customers (Tier 3A)
	GetCustomer(ctx context.Context, customerKey string) (*model.Customer, error)
	UpdateCustomer(ctx context.Context, customerKey string, customer model.Customer) (*model.Customer, error)
	PatchCustomer(ctx context.Context, customerKey string, patch map[string]interface{}) (*model.Customer, error)

	// Domains (Tier 3B)
	ListDomains(ctx context.Context, customerID string) ([]model.Domain, error)
	GetDomain(ctx context.Context, customerID, domainName string) (*model.Domain, error)
	AddDomain(ctx context.Context, customerID string, domain model.Domain) (model.Domain, error)
	DeleteDomain(ctx context.Context, customerID, domainName string) error

	// Domain Aliases (Tier 3C)
	ListDomainAliases(ctx context.Context, customerID string) ([]model.DomainAlias, error)
	GetDomainAlias(ctx context.Context, customerID, aliasName string) (*model.DomainAlias, error)
	CreateDomainAlias(ctx context.Context, customerID string, da model.DomainAlias) (model.DomainAlias, error)
	DeleteDomainAlias(ctx context.Context, customerID, aliasName string) error

	// ChromeOS Devices (Tier 4A)
	ListChromeOSDevices(ctx context.Context, customerID string, opts model.ListOptions) ([]model.ChromeOSDevice, string, error)
	GetChromeOSDevice(ctx context.Context, customerID, deviceID string) (*model.ChromeOSDevice, error)
	PatchChromeOSDevice(ctx context.Context, customerID, deviceID string, patch map[string]interface{}) (*model.ChromeOSDevice, error)
	UpdateChromeOSDevice(ctx context.Context, customerID, deviceID string, device model.ChromeOSDevice) (*model.ChromeOSDevice, error)
	MoveChromeOSDevices(ctx context.Context, customerID string, deviceIDs []string, orgUnitPath string) error
	BatchChangeChromeOSStatus(ctx context.Context, customerID string, deviceIDs []string, action string) error
	CountChromeOSDevices(ctx context.Context, customerID string) (int64, error)
	IssueChromeOSCommand(ctx context.Context, customerID, deviceID string, cmd model.ChromeOSCommand) (model.ChromeOSCommandResult, error)
	GetChromeOSCommand(ctx context.Context, customerID, deviceID, commandID string) (*model.ChromeOSCommandResult, error)

	// Mobile Devices (Tier 4B)
	ListMobileDevices(ctx context.Context, customerID string, opts model.ListOptions) ([]model.MobileDevice, string, error)
	GetMobileDevice(ctx context.Context, customerID, resourceID string) (*model.MobileDevice, error)
	DeleteMobileDevice(ctx context.Context, customerID, resourceID string) error
	MobileDeviceAction(ctx context.Context, customerID, resourceID string, action model.MobileDeviceAction) error

	// Cloud Identity Devices (Tier 4C)
	ListCIDevices(ctx context.Context, opts model.ListOptions) ([]model.CloudIdentityDevice, string, error)
	GetCIDevice(ctx context.Context, name string) (*model.CloudIdentityDevice, error)
	CreateCIDevice(ctx context.Context, device model.CloudIdentityDevice) (model.CloudIdentityDevice, error)
	DeleteCIDevice(ctx context.Context, name string) error
	CancelWipeCIDevice(ctx context.Context, name string) error
	WipeCIDevice(ctx context.Context, name string) error
	ListDeviceUsers(ctx context.Context, parent string) ([]model.DeviceUser, error)
	GetDeviceUser(ctx context.Context, name string) (*model.DeviceUser, error)
	DeleteDeviceUser(ctx context.Context, name string) error
	ApproveDeviceUser(ctx context.Context, name string) error
	BlockDeviceUser(ctx context.Context, name string) error
	WipeDeviceUser(ctx context.Context, name string) error
	CancelWipeDeviceUser(ctx context.Context, name string) error
	LookupDeviceUser(ctx context.Context, parent string) ([]model.DeviceUser, error)

	// Cloud Identity Groups (Tier 5A)
	ListCIGroups(ctx context.Context, opts model.ListOptions) ([]model.CloudIdentityGroup, string, error)
	GetCIGroup(ctx context.Context, name string) (*model.CloudIdentityGroup, error)
	CreateCIGroup(ctx context.Context, group model.CloudIdentityGroup) (model.CloudIdentityGroup, error)
	UpdateCIGroup(ctx context.Context, name string, group model.CloudIdentityGroup) (*model.CloudIdentityGroup, error)
	DeleteCIGroup(ctx context.Context, name string) error
	LookupCIGroup(ctx context.Context, key model.EntityKey) (*model.CloudIdentityGroup, error)
	SearchCIGroups(ctx context.Context, query string) ([]model.CloudIdentityGroup, error)
	GetCIGroupSecuritySettings(ctx context.Context, name string) (*model.SecuritySettings, error)
	UpdateCIGroupSecuritySettings(ctx context.Context, name string, settings model.SecuritySettings) (*model.SecuritySettings, error)

	// Cloud Identity Memberships (Tier 5B)
	ListCIMemberships(ctx context.Context, parent string) ([]model.Membership, error)
	GetCIMembership(ctx context.Context, name string) (*model.Membership, error)
	CreateCIMembership(ctx context.Context, parent string, membership model.Membership) (model.Membership, error)
	DeleteCIMembership(ctx context.Context, name string) error
	LookupCIMembership(ctx context.Context, parent string, key model.EntityKey) (*model.Membership, error)
	ModifyMembershipRoles(ctx context.Context, name string, roles model.ModifyMembershipRolesRequest) (*model.Membership, error)
	CheckTransitiveMembership(ctx context.Context, parent string, key model.EntityKey) (bool, error)
	GetMembershipGraph(ctx context.Context, parent string, query string) (*model.MembershipGraph, error)
	SearchTransitiveGroups(ctx context.Context, parent string, query string) ([]model.CloudIdentityGroup, error)
	SearchTransitiveMemberships(ctx context.Context, parent string, query string) ([]model.Membership, error)
	SearchDirectGroups(ctx context.Context, parent string, query string) ([]model.Membership, error)

	// Reports (Tier 6)
	ListActivities(ctx context.Context, userKey, applicationName string) ([]model.Activity, error)
	ListUsageReports(ctx context.Context, date, userKey, entityType, entityKey string) ([]model.UsageReport, error)

	// Security (Tier 7)
	ListTokens(ctx context.Context, userKey string) ([]model.Token, error)
	GetToken(ctx context.Context, userKey, clientID string) (*model.Token, error)
	DeleteToken(ctx context.Context, userKey, clientID string) error
	ListASPs(ctx context.Context, userKey string) ([]model.ASP, error)
	GetASP(ctx context.Context, userKey, codeID string) (*model.ASP, error)
	DeleteASP(ctx context.Context, userKey, codeID string) error
	ListVerificationCodes(ctx context.Context, userKey string) ([]model.VerificationCode, error)
	GenerateVerificationCodes(ctx context.Context, userKey string) error
	InvalidateVerificationCodes(ctx context.Context, userKey string) error
	TurnOff2SV(ctx context.Context, userKey string) error

	// User Invitations (Tier 7C)
	ListUserInvitations(ctx context.Context, parent string) ([]model.UserInvitation, error)
	GetUserInvitation(ctx context.Context, name string) (*model.UserInvitation, error)
	IsInvitableUser(ctx context.Context, name string) (bool, error)
	SendUserInvitation(ctx context.Context, name string) error
	CancelUserInvitation(ctx context.Context, name string) error

	// Custom Schemas (Tier 8A)
	ListSchemas(ctx context.Context, customerID string) ([]model.Schema, error)
	GetSchema(ctx context.Context, customerID, schemaKey string) (*model.Schema, error)
	CreateSchema(ctx context.Context, customerID string, schema model.Schema) (model.Schema, error)
	UpdateSchema(ctx context.Context, customerID, schemaKey string, schema model.Schema) (*model.Schema, error)
	PatchSchema(ctx context.Context, customerID, schemaKey string, patch map[string]interface{}) (*model.Schema, error)
	DeleteSchema(ctx context.Context, customerID, schemaKey string) error

	// Calendar Resources (Tier 8B)
	ListCalendarResources(ctx context.Context, customerID string) ([]model.CalendarResource, error)
	GetCalendarResource(ctx context.Context, customerID, resourceID string) (*model.CalendarResource, error)
	CreateCalendarResource(ctx context.Context, customerID string, resource model.CalendarResource) (model.CalendarResource, error)
	UpdateCalendarResource(ctx context.Context, customerID, resourceID string, resource model.CalendarResource) (*model.CalendarResource, error)
	PatchCalendarResource(ctx context.Context, customerID, resourceID string, patch map[string]interface{}) (*model.CalendarResource, error)
	DeleteCalendarResource(ctx context.Context, customerID, resourceID string) error

	// Buildings (Tier 8C)
	ListBuildings(ctx context.Context, customerID string) ([]model.Building, error)
	GetBuilding(ctx context.Context, customerID, buildingID string) (*model.Building, error)
	CreateBuilding(ctx context.Context, customerID string, building model.Building) (model.Building, error)
	UpdateBuilding(ctx context.Context, customerID, buildingID string, building model.Building) (*model.Building, error)
	PatchBuilding(ctx context.Context, customerID, buildingID string, patch map[string]interface{}) (*model.Building, error)
	DeleteBuilding(ctx context.Context, customerID, buildingID string) error

	// Features (Tier 8D)
	ListFeatures(ctx context.Context, customerID string) ([]model.Feature, error)
	GetFeature(ctx context.Context, customerID, featureKey string) (*model.Feature, error)
	CreateFeature(ctx context.Context, customerID string, feature model.Feature) (model.Feature, error)
	UpdateFeature(ctx context.Context, customerID, featureKey string, feature model.Feature) (*model.Feature, error)
	PatchFeature(ctx context.Context, customerID, featureKey string, patch map[string]interface{}) (*model.Feature, error)
	DeleteFeature(ctx context.Context, customerID, featureKey string) error
	RenameFeature(ctx context.Context, customerID, oldName, newName string) error

	// Groups Settings (Tier 8E)
	GetGroupSettings(ctx context.Context, groupUniqueId string) (*model.GroupSettings, error)
	UpdateGroupSettings(ctx context.Context, groupUniqueId string, settings model.GroupSettings) (*model.GroupSettings, error)
	PatchGroupSettings(ctx context.Context, groupUniqueId string, patch map[string]interface{}) (*model.GroupSettings, error)

	// Data Transfer (Tier 8F)
	ListTransferApplications(ctx context.Context) ([]model.TransferApplication, error)
	GetTransferApplication(ctx context.Context, applicationID string) (*model.TransferApplication, error)
	ListTransfers(ctx context.Context) ([]model.DataTransfer, error)
	GetTransfer(ctx context.Context, transferID string) (*model.DataTransfer, error)
	CreateTransfer(ctx context.Context, transfer model.DataTransfer) (model.DataTransfer, error)

	// Workspace Events/Subscriptions (Tier 8G)
	ListSubscriptions(ctx context.Context) ([]model.Subscription, error)
	GetSubscription(ctx context.Context, name string) (*model.Subscription, error)
	CreateSubscription(ctx context.Context, sub model.Subscription) (model.Subscription, error)
	UpdateSubscription(ctx context.Context, name string, sub model.Subscription) (*model.Subscription, error)
	DeleteSubscription(ctx context.Context, name string) error
	ReactivateSubscription(ctx context.Context, name string) error
}