package seed

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/saldeti/saldeti/internal/entra/model"
	"github.com/saldeti/saldeti/internal/entra/store"
)

func DumpStore(s store.Store) (*SeedConfig, error) {
	ctx := context.Background()
	cfg := &SeedConfig{}

	// ========== 1. Dump clients ==========
	clients, err := s.ListClients(ctx)
	if err != nil {
		return nil, fmt.Errorf("dumping clients: %w", err)
	}
	for _, c := range clients {
		cfg.Clients = append(cfg.Clients, SeedClient{
			ClientID:     c.ClientID,
			ClientSecret: c.ClientSecret,
			TenantID:     c.TenantID,
		})
	}

	// ========== 2. Dump users (sorted by UPN for determinism) ==========
	users, _, err := s.ListUsers(ctx, model.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("dumping users: %w", err)
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].UserPrincipalName < users[j].UserPrincipalName
	})

	userIDToUPN := make(map[string]string)
	userIDToIndex := make(map[string]int)

	for i, u := range users {
		isGuest := u.UserType == "Guest"
		su := SeedUser{
			Email:       u.UserPrincipalName,
			DisplayName: u.DisplayName,
			GivenName:   u.GivenName,
			Surname:     u.Surname,
			JobTitle:    u.JobTitle,
			Department:  u.Department,
			Enabled:     u.AccountEnabled,
			IsGuest:     isGuest,
		}
		if u.PasswordProfile != nil {
			su.Password = u.PasswordProfile.Password
		}

		// Dump AssignedLicenses
		for _, al := range u.AssignedLicenses {
			sl := model.SeedLicense{
				SkuPartNumber: al.SkuPartNumber,
			}
			for _, planID := range al.DisabledPlans {
				planName, found := model.FindServicePlanName(al.SkuPartNumber, planID)
				if found {
					sl.DisabledPlans = append(sl.DisabledPlans, planName)
				} else {
					sl.DisabledPlans = append(sl.DisabledPlans, planID)
				}
			}
			su.AssignedLicenses = append(su.AssignedLicenses, sl)
		}

		// Resolve manager
		mgr, err := s.GetManager(ctx, u.ID)
		if err == nil && mgr != nil {
			mgrUser, err := s.GetUser(ctx, mgr.ID)
			if err == nil {
				su.ManagerUPN = mgrUser.UserPrincipalName
			}
		} else if !errors.Is(err, store.ErrManagerNotFound) {
			return nil, fmt.Errorf("dumping manager for user %s: %w", u.ID, err)
		}

		cfg.Users = append(cfg.Users, su)
		userIDToUPN[u.ID] = u.UserPrincipalName
		userIDToIndex[u.ID] = i
	}

	// ========== 3. Dump groups (sorted by DisplayName for determinism) ==========
	groups, _, err := s.ListGroups(ctx, model.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("dumping groups: %w", err)
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].DisplayName < groups[j].DisplayName
	})

	// Build groupID → DisplayName map BEFORE processing memberships
	groupIDToName := make(map[string]string)
	for _, g := range groups {
		groupIDToName[g.ID] = g.DisplayName
	}

	for _, g := range groups {
		sg := SeedGroup{
			DisplayName:  g.DisplayName,
			Description:  g.Description,
			MailNickname: g.MailNickname,
			Visibility:   g.Visibility,
			GroupTypes:   g.GroupTypes,
		}

		// Dump members
		members, _, err := s.ListMembers(ctx, g.ID, model.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("dumping members of group %s: %w", g.ID, err)
		}
		for _, m := range members {
			switch m.ODataType {
			case "#microsoft.graph.user":
				if upn, ok := userIDToUPN[m.ID]; ok {
					sg.MemberUPNs = append(sg.MemberUPNs, upn)
				}
			case "#microsoft.graph.group":
				if name, ok := groupIDToName[m.ID]; ok {
					sg.MemberGroupNames = append(sg.MemberGroupNames, name)
				}
			case "#microsoft.graph.servicePrincipal":
				if sp, err := s.GetServicePrincipal(ctx, m.ID); err == nil {
					sg.MemberSPAppIDs = append(sg.MemberSPAppIDs, sp.AppID)
				}
			}
		}
		sort.Strings(sg.MemberUPNs)
		sort.Strings(sg.MemberGroupNames)
		sort.Strings(sg.MemberSPAppIDs)

		// Dump owners
		owners, _, err := s.ListOwners(ctx, g.ID, model.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("dumping owners of group %s: %w", g.ID, err)
		}
		for _, o := range owners {
			if o.ODataType == "#microsoft.graph.user" {
				if upn, ok := userIDToUPN[o.ID]; ok {
					sg.OwnerUPNs = append(sg.OwnerUPNs, upn)
				}
			}
		}
		sort.Strings(sg.OwnerUPNs)

		cfg.Groups = append(cfg.Groups, sg)
	}

	// ========== 4. Dump applications (sorted by AppID for determinism) ==========
	apps, _, err := s.ListApplications(ctx, model.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("dumping applications: %w", err)
	}
	sort.Slice(apps, func(i, j int) bool {
		return apps[i].AppID < apps[j].AppID
	})

	for _, app := range apps {
		sa := SeedApplication{
			DisplayName:    app.DisplayName,
			AppID:          app.AppID,
			Description:    app.Description,
			SignInAudience: app.SignInAudience,
			IdentifierUris: app.IdentifierUris,
		}

		for _, role := range app.AppRoles {
			isEnabled := role.IsEnabled != nil && *role.IsEnabled
			sa.AppRoles = append(sa.AppRoles, SeedAppRole{
				AllowedMemberTypes: role.AllowedMemberTypes,
				Description:        role.Description,
				DisplayName:        role.DisplayName,
				IsEnabled:          isEnabled,
				Value:              role.Value,
			})
		}

		appOwners, _, err := s.ListApplicationOwners(ctx, app.ID, model.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("dumping owners of application %s: %w", app.ID, err)
		}
		for _, o := range appOwners {
			if upn, ok := userIDToUPN[o.ID]; ok {
				sa.OwnerUPNs = append(sa.OwnerUPNs, upn)
			}
		}
		sort.Strings(sa.OwnerUPNs)

		cfg.Applications = append(cfg.Applications, sa)
	}

	// ========== 5. Dump service principals (sorted by AppID) ==========
	sps, _, err := s.ListServicePrincipals(ctx, model.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("dumping service principals: %w", err)
	}
	sort.Slice(sps, func(i, j int) bool {
		return sps[i].AppID < sps[j].AppID
	})

	for _, sp := range sps {
		seedSP := SeedServicePrincipal{
			AppID: sp.AppID,
		}
		// Dump SP owners
		spOwners, _, err := s.ListSPOwners(ctx, sp.ID, model.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("dumping owners of service principal %s: %w", sp.AppID, err)
		}
		for _, owner := range spOwners {
			if owner.ODataType == "#microsoft.graph.user" {
				if upn, ok := userIDToUPN[owner.ID]; ok {
					seedSP.OwnerUPNs = append(seedSP.OwnerUPNs, upn)
				}
			}
		}
		sort.Strings(seedSP.OwnerUPNs)
		cfg.ServicePrincipals = append(cfg.ServicePrincipals, seedSP)
	}

	// ========== 6. Dump app role assignments ==========
	for _, u := range users {
		assignments, _, err := s.ListAppRoleAssignments(ctx, u.ID, model.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("dumping app role assignments for user %s: %w", u.ID, err)
		}
		for _, a := range assignments {
			resourceSP, err := s.GetServicePrincipal(ctx, a.ResourceID)
			if err != nil {
				continue
			}
			var roleValue string
			for _, role := range resourceSP.AppRoles {
				if role.ID == a.AppRoleID {
					roleValue = role.Value
					break
				}
			}
			if roleValue == "" {
				continue
			}
			principalIdx := userIDToIndex[u.ID]
			cfg.AppRoleAssignments = append(cfg.AppRoleAssignments, SeedAppRoleAssignment{
				PrincipalIndex: principalIdx,
				ResourceAppID:  resourceSP.AppID,
				RoleValue:      roleValue,
			})
		}
	}

	// Dump app role assignments for groups
	for _, g := range groups {
		assignments, _, err := s.ListAppRoleAssignments(ctx, g.ID, model.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("dumping app role assignments for group %s: %w", g.DisplayName, err)
		}
		for _, a := range assignments {
			resourceSP, err := s.GetServicePrincipal(ctx, a.ResourceID)
			if err != nil {
				continue
			}
			var roleValue string
			for _, role := range resourceSP.AppRoles {
				if role.ID == a.AppRoleID {
					roleValue = role.Value
					break
				}
			}
			if roleValue == "" {
				continue
			}
			cfg.AppRoleAssignments = append(cfg.AppRoleAssignments, SeedAppRoleAssignment{
				PrincipalType: "group",
				PrincipalName: g.DisplayName,
				ResourceAppID: resourceSP.AppID,
				RoleValue:     roleValue,
			})
		}
	}

	// Dump app role assignments for service principals
	for _, sp := range sps {
		assignments, _, err := s.ListAppRoleAssignments(ctx, sp.ID, model.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("dumping app role assignments for service principal %s: %w", sp.AppID, err)
		}
		for _, a := range assignments {
			resourceSP, err := s.GetServicePrincipal(ctx, a.ResourceID)
			if err != nil {
				continue
			}
			var roleValue string
			for _, role := range resourceSP.AppRoles {
				if role.ID == a.AppRoleID {
					roleValue = role.Value
					break
				}
			}
			if roleValue == "" {
				continue
			}
			cfg.AppRoleAssignments = append(cfg.AppRoleAssignments, SeedAppRoleAssignment{
				PrincipalType: "service_principal",
				PrincipalName: sp.AppID,
				ResourceAppID: resourceSP.AppID,
				RoleValue:     roleValue,
			})
		}
	}

	// ========== 7. Dump OAuth2 permission grants ==========
	grants, _, err := s.ListOAuth2PermissionGrants(ctx, model.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("dumping oauth2 grants: %w", err)
	}

	for _, g := range grants {
		clientSP, err := s.GetServicePrincipal(ctx, g.ClientID)
		if err != nil {
			continue
		}
		resourceSP, err := s.GetServicePrincipal(ctx, g.ResourceID)
		if err != nil {
			continue
		}

		sg := SeedOAuth2Grant{
			ClientAppID:   clientSP.AppID,
			ResourceAppID: resourceSP.AppID,
			Scope:         g.Scope,
			ConsentType:   g.ConsentType,
		}

		if g.ConsentType == "Principal" && g.PrincipalID != "" {
			sg.PrincipalUPN = userIDToUPN[g.PrincipalID]
		}

		cfg.OAuth2Grants = append(cfg.OAuth2Grants, sg)
	}

	return cfg, nil
}
