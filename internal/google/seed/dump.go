package seed

import (
	"context"
	"fmt"
	"sort"

	"github.com/saldeti/saldeti/internal/google/model"
	"github.com/saldeti/saldeti/internal/google/store"
)

// DumpStore serializes the Google Workspace runtime store back into a
// GoogleSeedConfig suitable for marshalling to a seed JSON file. All entity
// lists are sorted for deterministic output.
func DumpStore(s store.Store) (*GoogleSeedConfig, error) {
	ctx := context.Background()
	customerID := "my_customer"
	cfg := &GoogleSeedConfig{}

	// ========== 1. Dump clients ==========
	clients, err := s.ListClients(ctx)
	if err != nil {
		return nil, fmt.Errorf("dumping clients: %w", err)
	}
	for _, c := range clients {
		cfg.Clients = append(cfg.Clients, GoogleSeedClient{
			ClientID:     c.ClientID,
			ClientSecret: c.ClientSecret,
		})
	}

	// ========== 2. Dump users (sorted by PrimaryEmail) ==========
	users, _, err := s.ListUsers(ctx, model.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("dumping users: %w", err)
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].PrimaryEmail < users[j].PrimaryEmail
	})

	idToEmail := make(map[string]string)
	for _, u := range users {
		aliases, err := s.ListUserAliases(ctx, u.PrimaryEmail)
		if err != nil {
			return nil, fmt.Errorf("dumping aliases for user %s: %w", u.PrimaryEmail, err)
		}
		sort.Strings(aliases)
		cfg.Users = append(cfg.Users, GoogleSeedUser{
			PrimaryEmail: u.PrimaryEmail,
			GivenName:    u.GivenName,
			FamilyName:   u.FamilyName,
			Suspended:    u.Suspended,
			IsAdmin:      u.IsAdmin,
			OrgUnitPath:  u.OrgUnitPath,
			Aliases:      aliases,
		})
		idToEmail[u.ID] = u.PrimaryEmail
	}

	// ========== 3. Dump groups (sorted by Email) ==========
	groups, _, err := s.ListGroups(ctx, model.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("dumping groups: %w", err)
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Email < groups[j].Email
	})

	for _, g := range groups {
		var memberEmails []string
		members, _, err := s.ListMembers(ctx, g.Email, model.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("dumping members of group %s: %w", g.Email, err)
		}
		for _, m := range members {
			if m.Type == "USER" {
				memberEmails = append(memberEmails, m.Email)
			}
		}
		sort.Strings(memberEmails)

		aliases, err := s.ListGroupAliases(ctx, g.Email)
		if err != nil {
			return nil, fmt.Errorf("dumping aliases for group %s: %w", g.Email, err)
		}
		sort.Strings(aliases)

		cfg.Groups = append(cfg.Groups, GoogleSeedGroup{
			Email:        g.Email,
			Name:         g.Name,
			Description:  g.Description,
			MemberEmails: memberEmails,
			Aliases:      aliases,
		})
	}

	// ========== 3b. Dump group settings (sorted by group Email) ==========
	for _, g := range groups {
		gs, err := s.GetGroupSettings(ctx, g.Email)
		if err != nil {
			return nil, fmt.Errorf("dumping group settings for %s: %w", g.Email, err)
		}
		// Only include groups that have at least one non-default setting.
		sgs := GoogleSeedGroupSettings{GroupEmail: g.Email}
		hasSettings := false
		if gs.WhoCanPostMessage != "" {
			sgs.WhoCanPostMessage = gs.WhoCanPostMessage
			hasSettings = true
		}
		if gs.IsArchived {
			sgs.IsArchived = &gs.IsArchived
			hasSettings = true
		}
		if gs.AllowExternalMembers {
			sgs.AllowExternalMembers = &gs.AllowExternalMembers
			hasSettings = true
		}
		if gs.ArchiveOnly {
			sgs.ArchiveOnly = &gs.ArchiveOnly
			hasSettings = true
		}
		if gs.WhoCanJoin != "" {
			sgs.WhoCanJoin = gs.WhoCanJoin
			hasSettings = true
		}
		if gs.WhoCanViewGroup != "" {
			sgs.WhoCanViewGroup = gs.WhoCanViewGroup
			hasSettings = true
		}
		if gs.WhoCanViewMembership != "" {
			sgs.WhoCanViewMembership = gs.WhoCanViewMembership
			hasSettings = true
		}
		if gs.WhoCanInvite != "" {
			sgs.WhoCanInvite = gs.WhoCanInvite
			hasSettings = true
		}
		if gs.WhoCanAdd != "" {
			sgs.WhoCanAdd = gs.WhoCanAdd
			hasSettings = true
		}
		if gs.WhoCanModerateMembers != "" {
			sgs.WhoCanModerateMembers = gs.WhoCanModerateMembers
			hasSettings = true
		}
		if gs.WhoCanModerateContent != "" {
			sgs.WhoCanModerateContent = gs.WhoCanModerateContent
			hasSettings = true
		}
		if gs.MessageModerationLevel != "" {
			sgs.MessageModerationLevel = gs.MessageModerationLevel
			hasSettings = true
		}
		if gs.PrimaryLanguage != "" {
			sgs.PrimaryLanguage = gs.PrimaryLanguage
			hasSettings = true
		}
		if gs.IncludeCustomFooter {
			sgs.IncludeCustomFooter = &gs.IncludeCustomFooter
			hasSettings = true
		}
		if gs.CustomFooterText != "" {
			sgs.CustomFooterText = gs.CustomFooterText
			hasSettings = true
		}
		if gs.MaxMessageBytes != 0 {
			sgs.MaxMessageBytes = gs.MaxMessageBytes
			hasSettings = true
		}
		if hasSettings {
			cfg.GroupSettings = append(cfg.GroupSettings, sgs)
		}
	}

	// ========== 4. Dump org units (sorted by OrgUnitPath) ==========
	orgUnits, err := s.ListOrgUnits(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("dumping org units: %w", err)
	}
	sort.Slice(orgUnits, func(i, j int) bool {
		return orgUnits[i].OrgUnitPath < orgUnits[j].OrgUnitPath
	})
	for _, ou := range orgUnits {
		cfg.OrgUnits = append(cfg.OrgUnits, GoogleSeedOrgUnit{
			Name:              ou.Name,
			Description:       ou.Description,
			ParentOrgUnitPath: ou.ParentOrgUnitPath,
			BlockInheritance:  ou.BlockInheritance,
		})
	}

	// ========== 5. Dump roles (sorted by RoleName) ==========
	roles, err := s.ListRoles(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("dumping roles: %w", err)
	}
	sort.Slice(roles, func(i, j int) bool {
		return roles[i].RoleName < roles[j].RoleName
	})
	roleIDToName := make(map[string]string)
	for _, r := range roles {
		var privileges []string
		for _, rp := range r.RolePrivileges {
			privileges = append(privileges, rp.PrivilegeName)
		}
		sort.Strings(privileges)
		cfg.Roles = append(cfg.Roles, GoogleSeedRole{
			RoleName:        r.RoleName,
			RoleDescription: r.RoleDescription,
			Privileges:      privileges,
		})
		roleIDToName[r.RoleId] = r.RoleName
	}

	// ========== 6. Dump role assignments ==========
	// Resolve AssignedTo (user ID) back to email and RoleId back to role name.
	roleAssignments, err := s.ListRoleAssignments(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("dumping role assignments: %w", err)
	}
	for _, ra := range roleAssignments {
		assignedEmail, ok := idToEmail[ra.AssignedTo]
		if !ok {
			// Skip assignments whose assignee user is no longer present.
			continue
		}
		roleIDOrName := ra.RoleId
		if name, ok := roleIDToName[ra.RoleId]; ok {
			roleIDOrName = name
		}
		cfg.RoleAssignments = append(cfg.RoleAssignments, GoogleSeedRoleAssignment{
			AssignedToEmail: assignedEmail,
			RoleIDOrName:    roleIDOrName,
		})
	}
	sort.Slice(cfg.RoleAssignments, func(i, j int) bool {
		if cfg.RoleAssignments[i].AssignedToEmail != cfg.RoleAssignments[j].AssignedToEmail {
			return cfg.RoleAssignments[i].AssignedToEmail < cfg.RoleAssignments[j].AssignedToEmail
		}
		return cfg.RoleAssignments[i].RoleIDOrName < cfg.RoleAssignments[j].RoleIDOrName
	})

	// ========== 7. Dump domains (sorted by DomainName) ==========
	domains, err := s.ListDomains(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("dumping domains: %w", err)
	}
	sort.Slice(domains, func(i, j int) bool {
		return domains[i].DomainName < domains[j].DomainName
	})
	for _, d := range domains {
		cfg.Domains = append(cfg.Domains, GoogleSeedDomain{
			DomainName: d.DomainName,
			IsPrimary:  d.IsPrimary,
		})
	}

	// ========== 8. Dump ChromeOS devices (sorted by SerialNumber) ==========
	chromeDevices, _, err := s.ListChromeOSDevices(ctx, customerID, model.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("dumping ChromeOS devices: %w", err)
	}
	sort.Slice(chromeDevices, func(i, j int) bool {
		return chromeDevices[i].SerialNumber < chromeDevices[j].SerialNumber
	})
	for _, d := range chromeDevices {
		cfg.ChromeOSDevices = append(cfg.ChromeOSDevices, GoogleSeedChromeOSDevice{
			SerialNumber:  d.SerialNumber,
			AnnotatedUser: d.AnnotatedUser,
			OrgUnitPath:   d.OrgUnitPath,
			Notes:         d.Notes,
		})
	}

	// ========== 9. Dump mobile devices (sorted by SerialNumber) ==========
	mobileDevices, _, err := s.ListMobileDevices(ctx, customerID, model.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("dumping mobile devices: %w", err)
	}
	sort.Slice(mobileDevices, func(i, j int) bool {
		return mobileDevices[i].SerialNumber < mobileDevices[j].SerialNumber
	})
	for _, d := range mobileDevices {
		cfg.MobileDevices = append(cfg.MobileDevices, GoogleSeedMobileDevice{
			SerialNumber: d.SerialNumber,
			Model:        d.Model,
			OS:           d.Os,
			Status:       d.Status,
		})
	}

	return cfg, nil
}