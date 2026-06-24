package seed

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/saldeti/saldeti/internal/google/model"
	"github.com/saldeti/saldeti/internal/google/store"
)

// LoadFromFile reads a JSON seed file and parses it into a GoogleSeedConfig.
func LoadFromFile(path string) (*GoogleSeedConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read seed file %s: %w", path, err)
	}

	var cfg GoogleSeedConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse seed file %s: %w", path, err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("seed validation failed: %w", err)
	}

	return &cfg, nil
}

// validateConfig validates required fields and cross-references in the seed config.
func validateConfig(cfg *GoogleSeedConfig) error {
	// Collect user emails for cross-reference validation
	userEmails := make(map[string]bool)
	groupEmails := make(map[string]bool)

	for i, client := range cfg.Clients {
		if client.ClientID == "" {
			return fmt.Errorf("clients[%d]: client_id is required", i)
		}
		if client.ClientSecret == "" {
			return fmt.Errorf("clients[%d]: client_secret is required", i)
		}
	}

	for i, user := range cfg.Users {
		if user.PrimaryEmail == "" {
			return fmt.Errorf("users[%d]: primary_email is required", i)
		}
		userEmails[user.PrimaryEmail] = true
	}

	for i, group := range cfg.Groups {
		if group.Email == "" {
			return fmt.Errorf("groups[%d]: email is required", i)
		}
		groupEmails[group.Email] = true
	}

	for i, ou := range cfg.OrgUnits {
		if ou.Name == "" {
			return fmt.Errorf("org_units[%d]: name is required", i)
		}
	}

	for i, role := range cfg.Roles {
		if role.RoleName == "" {
			return fmt.Errorf("roles[%d]: role_name is required", i)
		}
	}

	for i, ra := range cfg.RoleAssignments {
		if ra.AssignedToEmail == "" {
			return fmt.Errorf("role_assignments[%d]: assigned_to_email is required", i)
		}
		if ra.RoleIDOrName == "" {
			return fmt.Errorf("role_assignments[%d]: role_id_or_name is required", i)
		}
	}

	for i, domain := range cfg.Domains {
		if domain.DomainName == "" {
			return fmt.Errorf("domains[%d]: domain_name is required", i)
		}
	}

	for i, device := range cfg.ChromeOSDevices {
		if device.SerialNumber == "" {
			return fmt.Errorf("chromeos_devices[%d]: serial_number is required", i)
		}
	}

	for i, device := range cfg.MobileDevices {
		if device.SerialNumber == "" {
			return fmt.Errorf("mobile_devices[%d]: serial_number is required", i)
		}
	}

	for i, gs := range cfg.GroupSettings {
		if gs.GroupEmail == "" {
			return fmt.Errorf("group_settings[%d]: group_email is required", i)
		}
		if !groupEmails[gs.GroupEmail] {
			return fmt.Errorf("group_settings[%d]: group_email %s does not reference any group", i, gs.GroupEmail)
		}
	}

	// Cross-reference: member_emails in groups must reference existing user primary_email values
	for i, group := range cfg.Groups {
		for j, memberEmail := range group.MemberEmails {
			if !userEmails[memberEmail] {
				return fmt.Errorf("groups[%d].member_emails[%d]: %s does not reference any user", i, j, memberEmail)
			}
		}
	}

	// Cross-reference: assigned_to_email in role_assignments must reference existing user primary_email values
	for i, ra := range cfg.RoleAssignments {
		if !userEmails[ra.AssignedToEmail] {
			return fmt.Errorf("role_assignments[%d]: assigned_to_email %s does not reference any user", i, ra.AssignedToEmail)
		}
	}

	return nil
}

// SeedFromConfig populates a Google Workspace store from a seed config.
func SeedFromConfig(s store.Store, cfg *GoogleSeedConfig) error {
	ctx := context.Background()
	customerID := "my_customer"

	// 1. Register clients
	for _, client := range cfg.Clients {
		if err := s.RegisterClient(ctx, client.ClientID, client.ClientSecret); err != nil {
			if !errors.Is(err, store.ErrAlreadyExists) {
				return fmt.Errorf("failed to register client %s: %w", client.ClientID, err)
			}
		}
	}

	// 2. Create users
	emailToID := make(map[string]string)
	for _, user := range cfg.Users {
		displayName := strings.TrimSpace(user.GivenName + " " + user.FamilyName)
		u := model.User{
			Kind:         "admin#directory#user",
			PrimaryEmail: user.PrimaryEmail,
			GivenName:    user.GivenName,
			FamilyName:   user.FamilyName,
			DisplayName:  displayName,
			Suspended:    user.Suspended,
			OrgUnitPath:  user.OrgUnitPath,
		}
		if displayName != "" {
			u.Name = &model.UserName{
				GivenName:  user.GivenName,
				FamilyName: user.FamilyName,
				FullName:   displayName,
			}
		}
		// The Google Workspace model does not store plain-text passwords.
		// Setting ChangePasswordAtNextLogin signals that the user must set a
		// password on first login when a seed password is provided.
		if user.Password != "" {
			u.ChangePasswordAtNextLogin = true
		}
		userID := ""
		createdUser, err := s.CreateUser(ctx, u)
		if err != nil {
			if errors.Is(err, store.ErrAlreadyExists) {
				existing, lookupErr := s.GetUser(ctx, user.PrimaryEmail)
				if lookupErr != nil {
					return fmt.Errorf("user %s already exists but lookup failed: %w", user.PrimaryEmail, lookupErr)
				}
				userID = existing.ID
			} else {
				return fmt.Errorf("failed to create user %s: %w", user.PrimaryEmail, err)
			}
		} else {
			userID = createdUser.ID
		}
		emailToID[user.PrimaryEmail] = userID

		if user.IsAdmin {
			if err := s.MakeAdmin(ctx, userID, true); err != nil {
				return fmt.Errorf("failed to make user %s admin: %w", user.PrimaryEmail, err)
			}
		}
	}

	// 3. Add user aliases
	for _, user := range cfg.Users {
		userID, ok := emailToID[user.PrimaryEmail]
		if !ok {
			continue
		}
		for _, alias := range user.Aliases {
			if err := s.AddUserAlias(ctx, userID, alias); err != nil {
				return fmt.Errorf("failed to add alias %s for user %s: %w", alias, user.PrimaryEmail, err)
			}
		}
	}

	// 4. Create groups
	groupEmailToID := make(map[string]string)
	for _, group := range cfg.Groups {
		g := model.Group{
			Kind:        "admin#directory#group",
			Email:       group.Email,
			Name:        group.Name,
			Description: group.Description,
		}
		createdGroup, err := s.CreateGroup(ctx, g)
		if err != nil {
			if errors.Is(err, store.ErrAlreadyExists) {
				existing, lookupErr := s.GetGroup(ctx, group.Email)
				if lookupErr != nil {
					return fmt.Errorf("group %s already exists but lookup failed: %w", group.Email, lookupErr)
				}
				groupEmailToID[group.Email] = existing.ID
				continue
			}
			return fmt.Errorf("failed to create group %s: %w", group.Email, err)
		}
		groupEmailToID[group.Email] = createdGroup.ID
	}

	// 5. Add group members
	for _, group := range cfg.Groups {
		groupID, ok := groupEmailToID[group.Email]
		if !ok {
			continue
		}
		for _, memberEmail := range group.MemberEmails {
			userID, ok := emailToID[memberEmail]
			if !ok {
				return fmt.Errorf("group %s: member_emails %s does not reference any user", group.Email, memberEmail)
			}
			member := model.Member{
				Email: memberEmail,
				ID:    userID,
				Role:  "MEMBER",
				Type:  "USER",
			}
			if _, err := s.AddMember(ctx, groupID, member); err != nil {
				if !errors.Is(err, store.ErrAlreadyExists) {
					return fmt.Errorf("failed to add member %s to group %s: %w", memberEmail, group.Email, err)
				}
			}
		}
	}

	// 6. Add group aliases
	for _, group := range cfg.Groups {
		groupID, ok := groupEmailToID[group.Email]
		if !ok {
			continue
		}
		for _, alias := range group.Aliases {
			if err := s.AddGroupAlias(ctx, groupID, alias); err != nil {
				return fmt.Errorf("failed to add alias %s for group %s: %w", alias, group.Email, err)
			}
		}
	}

	// 7. Create org units
	for _, seedOU := range cfg.OrgUnits {
		parentPath := seedOU.ParentOrgUnitPath
		if parentPath == "" {
			parentPath = "/"
		}
		orgUnitPath := buildOrgUnitPath(parentPath, seedOU.Name)
		ou := model.OrgUnit{
			Kind:              "admin#directory#orgUnit",
			Name:              seedOU.Name,
			Description:       seedOU.Description,
			ParentOrgUnitPath: parentPath,
			BlockInheritance:  seedOU.BlockInheritance,
			OrgUnitPath:       orgUnitPath,
		}
		if _, err := s.CreateOrgUnit(ctx, customerID, ou); err != nil {
			return fmt.Errorf("failed to create org unit %s: %w", seedOU.Name, err)
		}
	}

	// 8. Create roles
	roleNameToID := make(map[string]string)
	for _, seedRole := range cfg.Roles {
		privileges := make([]model.RolePrivilege, 0, len(seedRole.Privileges))
		for _, p := range seedRole.Privileges {
			privileges = append(privileges, model.RolePrivilege{
				PrivilegeName: p,
			})
		}
		role := model.Role{
			Kind:            "admin#directory#role",
			RoleName:        seedRole.RoleName,
			RoleDescription: seedRole.RoleDescription,
			RolePrivileges:  privileges,
		}
		createdRole, err := s.CreateRole(ctx, customerID, role)
		if err != nil {
			return fmt.Errorf("failed to create role %s: %w", seedRole.RoleName, err)
		}
		roleNameToID[seedRole.RoleName] = createdRole.RoleId
	}

	// 9. Create role assignments
	for _, seedRA := range cfg.RoleAssignments {
		userID, ok := emailToID[seedRA.AssignedToEmail]
		if !ok {
			return fmt.Errorf("role_assignment: assigned_to_email %s does not reference any user", seedRA.AssignedToEmail)
		}
		roleID := seedRA.RoleIDOrName
		if mappedID, ok := roleNameToID[seedRA.RoleIDOrName]; ok {
			roleID = mappedID
		}
		ra := model.RoleAssignment{
			Kind:       "admin#directory#roleAssignment",
			RoleId:     roleID,
			AssignedTo: userID,
		}
		if _, err := s.CreateRoleAssignment(ctx, customerID, ra); err != nil {
			return fmt.Errorf("failed to create role assignment for %s: %w", seedRA.AssignedToEmail, err)
		}
	}

	// 10. Create domains
	for _, seedDomain := range cfg.Domains {
		domain := model.Domain{
			Kind:       "admin#directory#domain",
			DomainName: seedDomain.DomainName,
			IsPrimary:  seedDomain.IsPrimary,
		}
		if _, err := s.AddDomain(ctx, customerID, domain); err != nil {
			return fmt.Errorf("failed to add domain %s: %w", seedDomain.DomainName, err)
		}
	}

	// 11. Create ChromeOS devices
	for _, seedDevice := range cfg.ChromeOSDevices {
		status := seedDevice.Status
		if status == "" {
			status = "ACTIVE"
		}
		device := model.ChromeOSDevice{
			Kind:          "admin#directory#chromeosdevice",
			SerialNumber:  seedDevice.SerialNumber,
			AnnotatedUser: seedDevice.AnnotatedUser,
			OrgUnitPath:   seedDevice.OrgUnitPath,
			Notes:         seedDevice.Notes,
			Status:        status,
		}
		if _, err := s.CreateChromeOSDevice(ctx, customerID, device); err != nil {
			return fmt.Errorf("failed to create ChromeOS device %s: %w", seedDevice.SerialNumber, err)
		}
	}

	// 12. Create mobile devices
	for _, seedDevice := range cfg.MobileDevices {
		device := model.MobileDevice{
			Kind:         "admin#directory#mobiledevice",
			SerialNumber: seedDevice.SerialNumber,
			Model:        seedDevice.Model,
			Os:           seedDevice.OS,
			Status:       seedDevice.Status,
		}
		if _, err := s.CreateMobileDevice(ctx, customerID, device); err != nil {
			return fmt.Errorf("failed to create mobile device %s: %w", seedDevice.SerialNumber, err)
		}
	}

	// 13. Apply group settings (must be after groups are created)
	for _, gs := range cfg.GroupSettings {
		patch := map[string]any{}
		if gs.WhoCanPostMessage != "" {
			patch["whoCanPostMessage"] = gs.WhoCanPostMessage
		}
		if gs.IsArchived != nil {
			patch["isArchived"] = *gs.IsArchived
		}
		if gs.AllowExternalMembers != nil {
			patch["allowExternalMembers"] = *gs.AllowExternalMembers
		}
		if gs.ArchiveOnly != nil {
			patch["archiveOnly"] = *gs.ArchiveOnly
		}
		if gs.WhoCanJoin != "" {
			patch["whoCanJoin"] = gs.WhoCanJoin
		}
		if gs.WhoCanViewGroup != "" {
			patch["whoCanViewGroup"] = gs.WhoCanViewGroup
		}
		if gs.WhoCanViewMembership != "" {
			patch["whoCanViewMembership"] = gs.WhoCanViewMembership
		}
		if gs.WhoCanInvite != "" {
			patch["whoCanInvite"] = gs.WhoCanInvite
		}
		if gs.WhoCanAdd != "" {
			patch["whoCanAdd"] = gs.WhoCanAdd
		}
		if gs.WhoCanModerateMembers != "" {
			patch["whoCanModerateMembers"] = gs.WhoCanModerateMembers
		}
		if gs.WhoCanModerateContent != "" {
			patch["whoCanModerateContent"] = gs.WhoCanModerateContent
		}
		if gs.MessageModerationLevel != "" {
			patch["messageModerationLevel"] = gs.MessageModerationLevel
		}
		if gs.PrimaryLanguage != "" {
			patch["primaryLanguage"] = gs.PrimaryLanguage
		}
		if gs.IncludeCustomFooter != nil {
			patch["includeCustomFooter"] = *gs.IncludeCustomFooter
		}
		if gs.CustomFooterText != "" {
			patch["customFooterText"] = gs.CustomFooterText
		}
		if gs.MaxMessageBytes != 0 {
			patch["maxMessageBytes"] = gs.MaxMessageBytes
		}
		if len(patch) == 0 {
			continue
		}
		if _, err := s.PatchGroupSettings(ctx, gs.GroupEmail, patch); err != nil {
			return fmt.Errorf("failed to apply group settings for %s: %w", gs.GroupEmail, err)
		}
	}

	return nil
}

func buildOrgUnitPath(parentPath, name string) string {
	parentPath = strings.TrimRight(parentPath, "/")
	if parentPath == "" {
		return "/" + name
	}
	return parentPath + "/" + name
}
