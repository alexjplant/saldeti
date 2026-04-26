package seed

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/saldeti/saldeti/internal/model"
	"github.com/saldeti/saldeti/internal/store"
)

// LoadFromFile reads a JSON seed file and parses it into a SeedConfig.
func LoadFromFile(path string) (*SeedConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read seed file: %w", err)
	}

	var cfg SeedConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse seed file: %w", err)
	}

	// Validate the configuration
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// validateConfig checks that the configuration is valid.
func validateConfig(cfg *SeedConfig) error {
	// Validate client fields
	for i, client := range cfg.Clients {
		if client.ClientID == "" {
			return fmt.Errorf("client[%d]: client_id is required", i)
		}
		if client.ClientSecret == "" {
			return fmt.Errorf("client[%d]: client_secret is required", i)
		}
		if client.TenantID == "" {
			return fmt.Errorf("client[%d]: tenant_id is required", i)
		}
	}

	// Validate user fields
	for i, user := range cfg.Users {
		if user.Email == "" {
			return fmt.Errorf("user[%d]: email is required", i)
		}
		if user.DisplayName == "" {
			return fmt.Errorf("user[%d]: display_name is required", i)
		}
		if user.Password == "" {
			return fmt.Errorf("user[%d]: password is required", i)
		}
	}

	// Validate group fields
	for i, group := range cfg.Groups {
		if group.DisplayName == "" {
			return fmt.Errorf("group[%d]: display_name is required", i)
		}
	}

	// Build set of all user emails for validation
	userEmails := make(map[string]bool)
	for _, user := range cfg.Users {
		userEmails[user.Email] = true
	}

	// Validate ManagerUPN on users
	for i, user := range cfg.Users {
		if user.ManagerUPN != "" {
			if !userEmails[user.ManagerUPN] {
				return fmt.Errorf("user[%d]: manager_upn %s does not reference any user", i, user.ManagerUPN)
			}
		}
	}

	// Validate MemberUPNs on groups
	for i, group := range cfg.Groups {
		for _, upn := range group.MemberUPNs {
			if !userEmails[upn] {
				return fmt.Errorf("group[%d]: member_upns %s does not reference any user", i, upn)
			}
		}
	}

	// Build set of all group display names for validation
	groupNames := make(map[string]bool)
	for _, group := range cfg.Groups {
		groupNames[group.DisplayName] = true
	}

	// Validate MemberGroupNames on groups
	for i, group := range cfg.Groups {
		for _, groupName := range group.MemberGroupNames {
			if !groupNames[groupName] {
				return fmt.Errorf("group[%d]: member_group_names %s does not reference any group", i, groupName)
			}
		}
	}

	// Validate OwnerUPNs on groups
	for i, group := range cfg.Groups {
		for _, upn := range group.OwnerUPNs {
			if !userEmails[upn] {
				return fmt.Errorf("group[%d]: owner_upns %s does not reference any user", i, upn)
			}
		}
	}

	numUsers := len(cfg.Users)
	numGroups := len(cfg.Groups)

	// Validate membership indices
	for i, membership := range cfg.Memberships {
		if membership.GroupIndex == nil {
			return fmt.Errorf("membership[%d]: group_index is required", i)
		}
		if *membership.GroupIndex < 0 || *membership.GroupIndex >= numGroups {
			return fmt.Errorf("membership[%d]: group_index %d is out of range (0-%d)", i, *membership.GroupIndex, numGroups-1)
		}
		if membership.UserIndex != nil {
			if *membership.UserIndex < 0 || *membership.UserIndex >= numUsers {
				return fmt.Errorf("membership[%d]: user_index %d is out of range (0-%d)", i, *membership.UserIndex, numUsers-1)
			}
		}
		if membership.MemberGroupIndex != nil {
			if *membership.MemberGroupIndex < 0 || *membership.MemberGroupIndex >= numGroups {
				return fmt.Errorf("membership[%d]: member_group_index %d is out of range (0-%d)", i, *membership.MemberGroupIndex, numGroups-1)
			}
		}
		if membership.UserIndex != nil && membership.MemberGroupIndex != nil {
			return fmt.Errorf("membership[%d]: cannot set both user_index and member_group_index", i)
		}
		if membership.UserIndex == nil && membership.MemberGroupIndex == nil {
			return fmt.Errorf("membership[%d]: either user_index or member_group_index must be set", i)
		}
	}

	// Validate manager indices
	for i, mgr := range cfg.Managers {
		if mgr.UserIndex < 0 || mgr.UserIndex >= numUsers {
			return fmt.Errorf("manager[%d]: user_index %d is out of range (0-%d)", i, mgr.UserIndex, numUsers-1)
		}
		if mgr.ManagerIndex < 0 || mgr.ManagerIndex >= numUsers {
			return fmt.Errorf("manager[%d]: manager_index %d is out of range (0-%d)", i, mgr.ManagerIndex, numUsers-1)
		}
	}

	// Validate ownership indices
	for i, ownership := range cfg.Ownerships {
		if ownership.UserIndex < 0 || ownership.UserIndex >= numUsers {
			return fmt.Errorf("ownership[%d]: user_index %d is out of range (0-%d)", i, ownership.UserIndex, numUsers-1)
		}
		if ownership.GroupIndex < 0 || ownership.GroupIndex >= numGroups {
			return fmt.Errorf("ownership[%d]: group_index %d is out of range (0-%d)", i, ownership.GroupIndex, numGroups-1)
		}
	}

	// Validate application fields
	for i, app := range cfg.Applications {
		if app.DisplayName == "" {
			return fmt.Errorf("application[%d]: display_name is required", i)
		}
	}

	// Validate app role assignment fields
	for i, assignment := range cfg.AppRoleAssignments {
		if assignment.PrincipalIndex < 0 || assignment.PrincipalIndex >= numUsers {
			return fmt.Errorf("app_role_assignment[%d]: principal_index %d is out of range (0-%d)", i, assignment.PrincipalIndex, numUsers-1)
		}
		if assignment.ResourceAppID == "" {
			return fmt.Errorf("app_role_assignment[%d]: resource_app_id is required", i)
		}
		if assignment.RoleValue == "" {
			return fmt.Errorf("app_role_assignment[%d]: role_value is required", i)
		}
	}

	// Validate OAuth2 grant fields
	for i, grant := range cfg.OAuth2Grants {
		if grant.ClientAppID == "" {
			return fmt.Errorf("oauth2_grant[%d]: client_app_id is required", i)
		}
		if grant.ResourceAppID == "" {
			return fmt.Errorf("oauth2_grant[%d]: resource_app_id is required", i)
		}
		if grant.Scope == "" {
			return fmt.Errorf("oauth2_grant[%d]: scope is required", i)
		}
		if grant.ConsentType == "" {
			return fmt.Errorf("oauth2_grant[%d]: consent_type is required", i)
		}
		if grant.ConsentType != "AllPrincipals" && grant.ConsentType != "Principal" {
			return fmt.Errorf("oauth2_grant[%d]: consent_type must be 'AllPrincipals' or 'Principal'", i)
		}
		if grant.ConsentType == "Principal" && grant.PrincipalUPN == "" {
			return fmt.Errorf("oauth2_grant[%d]: principal_upn is required when consent_type is 'Principal'", i)
		}
	}

	return nil
}

// SeedFromConfig seeds the store with data from a SeedConfig.
func SeedFromConfig(s store.Store, cfg *SeedConfig) error {
	ctx := context.Background()

	// Register clients
	for _, client := range cfg.Clients {
		err := s.RegisterClient(ctx, client.ClientID, client.ClientSecret, client.TenantID)
		if err != nil && !errors.Is(err, store.ErrDuplicateClient) {
			return fmt.Errorf("failed to register client %s: %w", client.ClientID, err)
		}
	}

	// Create users and store their IDs
	userIDs := make([]string, len(cfg.Users))
	now := time.Now()
	for i, user := range cfg.Users {
		accountEnabled := true
		if user.Enabled != nil {
			accountEnabled = *user.Enabled
		}

		userType := "Member"
		if user.IsGuest {
			userType = "Guest"
		}

		u := model.User{
			ODataType:         "#microsoft.graph.user",
			ID:                user.ID,
			DisplayName:       user.DisplayName,
			UserPrincipalName: user.Email,
			Mail:              user.Email,
			GivenName:         user.GivenName,
			Surname:           user.Surname,
			JobTitle:          user.JobTitle,
			Department:        user.Department,
			AccountEnabled:    &accountEnabled,
			CreatedDateTime:   &now,
			UserType:          userType,
			PasswordProfile: &model.PasswordProfile{
				Password: user.Password,
			},
		}

		// Process assigned licenses from seed
		if len(user.AssignedLicenses) > 0 {
			licenses := make([]model.AssignedLicense, 0, len(user.AssignedLicenses))
			for _, sl := range user.AssignedLicenses {
				skuID, found := model.FindSkuByPartNumber(sl.SkuPartNumber)
				if !found {
					return fmt.Errorf("user[%d]: unknown skuPartNumber %q in assigned_licenses", i, sl.SkuPartNumber)
				}
				// Convert disabled plan names to GUIDs
				disabledPlans := make([]string, 0)
				if sl.DisabledPlans != nil {
					for _, planName := range sl.DisabledPlans {
						planID, found := model.FindServicePlanID(sl.SkuPartNumber, planName)
						if !found {
							return fmt.Errorf("user[%d]: unknown service plan name %q for sku %q", i, planName, sl.SkuPartNumber)
						}
						disabledPlans = append(disabledPlans, planID)
					}
				}
				licenses = append(licenses, model.AssignedLicense{
					SkuID:         skuID,
					SkuPartNumber: sl.SkuPartNumber,
					DisabledPlans: disabledPlans,
				})
			}
			u.AssignedLicenses = licenses
		}

		createdUser, err := s.CreateUser(ctx, u)
		if err != nil {
			if errors.Is(err, store.ErrDuplicateUPN) {
				// Look up the existing user to get their real ID
				existing, lookupErr := s.GetUserByUPN(ctx, u.UserPrincipalName)
				if lookupErr != nil {
					return fmt.Errorf("user %s already exists but lookup failed: %w", u.UserPrincipalName, lookupErr)
				}
				userIDs[i] = existing.ID
				continue
			}
			return fmt.Errorf("failed to create user %s: %w", user.Email, err)
		}
		userIDs[i] = createdUser.ID
	}

	// Build map of UPN to user ID for resolving new fields
	upnToID := make(map[string]string)
	for i, user := range cfg.Users {
		upnToID[user.Email] = userIDs[i]
	}

	// Create groups and store their IDs
	groupIDs := make([]string, len(cfg.Groups))
	securityEnabled := true
	mailEnabled := false

	// Track groups that need to be looked up (when they already exist)
	groupsToLookup := make(map[int]string) // index -> display name

	for i, group := range cfg.Groups {
		visibility := group.Visibility
		if visibility == "" {
			visibility = "Public"
		}

		// Determine security and mail settings based on group types
		isUnified := false
		for _, gt := range group.GroupTypes {
			if gt == "Unified" {
				isUnified = true
				break
			}
		}

		secEnabled := &securityEnabled
		mlEnabled := &mailEnabled
		if isUnified {
			secEnabled = boolPtr(false)
			mlEnabled = boolPtr(true)
		}

		g := model.Group{
			ODataType:       "#microsoft.graph.group",
			ID:              group.ID,
			DisplayName:     group.DisplayName,
			Description:     group.Description,
			MailNickname:    group.MailNickname,
			Visibility:      visibility,
			GroupTypes:      group.GroupTypes,
			SecurityEnabled: secEnabled,
			MailEnabled:     mlEnabled,
			CreatedDateTime: &now,
		}

		createdGroup, err := s.CreateGroup(ctx, g)
		if err != nil {
			if errors.Is(err, store.ErrDuplicateGroup) {
				// Mark this group for later lookup
				groupsToLookup[i] = group.DisplayName
				continue
			}
			return fmt.Errorf("failed to create group %s: %w", group.DisplayName, err)
		}
		groupIDs[i] = createdGroup.ID
	}

	// For any groups that already existed, look up their IDs
	if len(groupsToLookup) > 0 {
		allGroups, _, err := s.ListGroups(ctx, model.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list groups for ID lookup: %w", err)
		}

		// Build a map of display name to ID
		nameToID := make(map[string]string)
		for _, grp := range allGroups {
			nameToID[grp.DisplayName] = grp.ID
		}

		// Look up IDs for groups that already existed
		for idx, displayName := range groupsToLookup {
			groupID, ok := nameToID[displayName]
			if !ok {
				return fmt.Errorf("group %s already exists but was not found in group list", displayName)
			}
			groupIDs[idx] = groupID
		}
	}

	// Build map of group display name to group ID for resolving new fields
	groupNameToID := make(map[string]string)
	for i, group := range cfg.Groups {
		groupNameToID[group.DisplayName] = groupIDs[i]
	}

	// Process manager_upn from users (new schema)
	for _, user := range cfg.Users {
		if user.ManagerUPN != "" {
			userID := upnToID[user.Email]
			managerID := upnToID[user.ManagerUPN]
			if err := s.SetManager(ctx, userID, managerID); err != nil {
				return fmt.Errorf("failed to set manager %s for user %s: %w", user.ManagerUPN, user.Email, err)
			}
		}
	}

	// Process member_upns, member_group_names, owner_upns from groups (new schema)
	for _, group := range cfg.Groups {
		groupID := groupNameToID[group.DisplayName]
		for _, upn := range group.MemberUPNs {
			userID := upnToID[upn]
			if err := s.AddMember(ctx, groupID, userID, "user"); err != nil && !errors.Is(err, store.ErrAlreadyMember) {
				return fmt.Errorf("failed to add user %s to group %s: %w", upn, group.DisplayName, err)
			}
		}
		for _, memberGroupName := range group.MemberGroupNames {
			memberGroupID := groupNameToID[memberGroupName]
			if err := s.AddMember(ctx, groupID, memberGroupID, "group"); err != nil && !errors.Is(err, store.ErrAlreadyMember) {
				return fmt.Errorf("failed to add group %s to group %s: %w", memberGroupName, group.DisplayName, err)
			}
		}
		for _, upn := range group.OwnerUPNs {
			userID := upnToID[upn]
			if err := s.AddOwner(ctx, groupID, userID, "user"); err != nil && !errors.Is(err, store.ErrAlreadyOwner) {
				return fmt.Errorf("failed to add owner %s to group %s: %w", upn, group.DisplayName, err)
			}
		}
	}

	// Create memberships
	for _, membership := range cfg.Memberships {
		groupID := groupIDs[*membership.GroupIndex]

		if membership.UserIndex != nil {
			userID := userIDs[*membership.UserIndex]
			err := s.AddMember(ctx, groupID, userID, "user")
			if err != nil && !errors.Is(err, store.ErrAlreadyMember) {
				return fmt.Errorf("failed to add user %d to group %d: %w", *membership.UserIndex, *membership.GroupIndex, err)
			}
		}

		if membership.MemberGroupIndex != nil {
			memberGroupID := groupIDs[*membership.MemberGroupIndex]
			err := s.AddMember(ctx, groupID, memberGroupID, "group")
			if err != nil && !errors.Is(err, store.ErrAlreadyMember) {
				return fmt.Errorf("failed to add group %d to group %d: %w", *membership.MemberGroupIndex, *membership.GroupIndex, err)
			}
		}
	}

	// Set managers
	for _, mgr := range cfg.Managers {
		userID := userIDs[mgr.UserIndex]
		managerID := userIDs[mgr.ManagerIndex]
		err := s.SetManager(ctx, userID, managerID)
		if err != nil {
			return fmt.Errorf("failed to set manager for user %d: %w", mgr.UserIndex, err)
		}
	}

	// Set ownerships
	for _, ownership := range cfg.Ownerships {
		groupID := groupIDs[ownership.GroupIndex]
		userID := userIDs[ownership.UserIndex]
		err := s.AddOwner(ctx, groupID, userID, "user")
		if err != nil && !errors.Is(err, store.ErrAlreadyOwner) {
			return fmt.Errorf("failed to add user %d as owner of group %d: %w", ownership.UserIndex, ownership.GroupIndex, err)
		}
	}

	// ===== Create applications =====
	appObjIDs := make([]string, len(cfg.Applications))
	for i, seedApp := range cfg.Applications {
		// Convert SeedAppRole to model.AppRole
		appRoles := make([]model.AppRole, 0, len(seedApp.AppRoles))
		for _, sr := range seedApp.AppRoles {
			isEnabled := sr.IsEnabled
			appRoles = append(appRoles, model.AppRole{
				ID:                 uuid.New().String(),
				AllowedMemberTypes: sr.AllowedMemberTypes,
				Description:        sr.Description,
				DisplayName:        sr.DisplayName,
				IsEnabled:          &isEnabled,
				Value:              sr.Value,
			})
		}

		app := model.Application{
			ODataType:      "#microsoft.graph.application",
			AppID:          seedApp.AppID,
			DisplayName:    seedApp.DisplayName,
			Description:    seedApp.Description,
			SignInAudience: seedApp.SignInAudience,
			IdentifierUris: seedApp.IdentifierUris,
			AppRoles:       appRoles,
		}
		if len(app.IdentifierUris) == 0 {
			app.IdentifierUris = []string{}
		}

		createdApp, err := s.CreateApplication(ctx, app)
		if err != nil {
			if errors.Is(err, store.ErrDuplicateAppID) {
				// Look up existing application
				existing, lookupErr := s.GetApplicationByAppID(ctx, app.AppID)
				if lookupErr != nil {
					return fmt.Errorf("application %s already exists but lookup failed: %w", seedApp.AppID, lookupErr)
				}
				appObjIDs[i] = existing.ID
				continue
			}
			return fmt.Errorf("failed to create application %s: %w", seedApp.DisplayName, err)
		}
		appObjIDs[i] = createdApp.ID
	}

	// Process application owner_upns
	for i, seedApp := range cfg.Applications {
		appObjID := appObjIDs[i]
		for _, upn := range seedApp.OwnerUPNs {
			userID, ok := upnToID[upn]
			if !ok {
				return fmt.Errorf("application[%d]: owner_upn %s does not reference any user", i, upn)
			}
			if err := s.AddApplicationOwner(ctx, appObjID, userID, "user"); err != nil && !errors.Is(err, store.ErrAlreadyAppOwner) {
				return fmt.Errorf("failed to add owner %s to application %s: %w", upn, seedApp.DisplayName, err)
			}
		}
	}

	// Process ServicePrincipals
	for _, seedSP := range cfg.ServicePrincipals {
		// Check if SP already exists (auto-created by CreateApplication)
		_, err := s.GetServicePrincipalByAppID(ctx, seedSP.AppID)
		if err == nil {
			// SP already exists, skip
			continue
		}
		if !errors.Is(err, store.ErrServicePrincipalNotFound) {
			return fmt.Errorf("failed to check SP for appId %s: %w", seedSP.AppID, err)
		}
		// Create SP
		sp := model.ServicePrincipal{
			AppID: seedSP.AppID,
		}
		if _, err := s.CreateServicePrincipal(ctx, sp); err != nil {
			return fmt.Errorf("failed to create service principal for appId %s: %w", seedSP.AppID, err)
		}
	}

	// Process AppRoleAssignments
	for i, assignment := range cfg.AppRoleAssignments {
		principalID := userIDs[assignment.PrincipalIndex]

		// Find resource SP by appId
		resourceSP, err := s.GetServicePrincipalByAppID(ctx, assignment.ResourceAppID)
		if err != nil {
			return fmt.Errorf("app_role_assignment[%d]: failed to find SP for resource_app_id %s: %w", i, assignment.ResourceAppID, err)
		}

		// Find matching app role by value
		var appRoleID string
		for _, role := range resourceSP.AppRoles {
			if role.Value == assignment.RoleValue {
				appRoleID = role.ID
				break
			}
		}
		if appRoleID == "" {
			return fmt.Errorf("app_role_assignment[%d]: role value %s not found on SP %s", i, assignment.RoleValue, assignment.ResourceAppID)
		}

		if _, err := s.CreateAppRoleAssignment(ctx, resourceSP.ID, principalID, appRoleID); err != nil {
			return fmt.Errorf("failed to create app role assignment[%d]: %w", i, err)
		}
	}

	// Process OAuth2Grants
	for i, seedGrant := range cfg.OAuth2Grants {
		// Find client SP by appId
		clientSP, err := s.GetServicePrincipalByAppID(ctx, seedGrant.ClientAppID)
		if err != nil {
			return fmt.Errorf("oauth2_grant[%d]: failed to find client SP for %s: %w", i, seedGrant.ClientAppID, err)
		}

		// Find resource SP by appId
		resourceSP, err := s.GetServicePrincipalByAppID(ctx, seedGrant.ResourceAppID)
		if err != nil {
			return fmt.Errorf("oauth2_grant[%d]: failed to find resource SP for %s: %w", i, seedGrant.ResourceAppID, err)
		}

		grant := model.OAuth2PermissionGrant{
			ODataType:   "#microsoft.graph.oAuth2PermissionGrant",
			ClientID:    clientSP.ID,
			ResourceID:  resourceSP.ID,
			Scope:       seedGrant.Scope,
			ConsentType: seedGrant.ConsentType,
		}

		if seedGrant.ConsentType == "Principal" {
			principalID, ok := upnToID[seedGrant.PrincipalUPN]
			if !ok {
				return fmt.Errorf("oauth2_grant[%d]: principal_upn %s does not reference any user", i, seedGrant.PrincipalUPN)
			}
			grant.PrincipalID = principalID
		}

		if _, err := s.CreateOAuth2PermissionGrant(ctx, grant); err != nil {
			return fmt.Errorf("failed to create oauth2 grant[%d]: %w", i, err)
		}
	}

	return nil
}

func boolPtr(b bool) *bool {
	return &b
}
