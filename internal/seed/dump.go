package seed

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/saldeti/saldeti/internal/model"
	"github.com/saldeti/saldeti/internal/store"
)

func DumpStore(s store.Store) (*SeedConfig, error) {
	ctx := context.Background()
	cfg := &SeedConfig{}

	// Dump clients
	clients, err := s.ListClients(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range clients {
		cfg.Clients = append(cfg.Clients, SeedClient{
			ClientID:     c.ClientID,
			ClientSecret: c.ClientSecret,
			TenantID:     c.TenantID,
		})
	}

	// Dump users
	users, _, err := s.ListUsers(ctx, model.ListOptions{})
	if err != nil {
		return nil, err
	}
	// Sort users by ID for deterministic ordering
	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})
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
		cfg.Users = append(cfg.Users, su)
		userIDToIndex[u.ID] = i
	}

	// Dump groups
	groups, _, err := s.ListGroups(ctx, model.ListOptions{})
	if err != nil {
		return nil, err
	}
	// Sort groups by ID for deterministic ordering
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].ID < groups[j].ID
	})
	groupIDToIndex := make(map[string]int)
	for i, g := range groups {
		sg := SeedGroup{
			DisplayName:  g.DisplayName,
			Description:  g.Description,
			MailNickname: g.MailNickname,
			Visibility:   g.Visibility,
			GroupTypes:   g.GroupTypes,
		}
		cfg.Groups = append(cfg.Groups, sg)
		groupIDToIndex[g.ID] = i
	}

	// Dump memberships
	for gi, g := range groups {
		members, _, err := s.ListMembers(ctx, g.ID, model.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("dumping members of group %s: %w", g.ID, err)
		}
		for _, m := range members {
			if idx, ok := userIDToIndex[m.ID]; ok {
				cfg.Memberships = append(cfg.Memberships, SeedMembership{
					UserIndex:  intPtr(idx),
					GroupIndex: intPtr(gi),
				})
			} else if gidx, ok := groupIDToIndex[m.ID]; ok {
				cfg.Memberships = append(cfg.Memberships, SeedMembership{
					MemberGroupIndex: intPtr(gidx),
					GroupIndex:       intPtr(gi),
				})
			}
		}
	}

	// Dump owners
	for gi, g := range groups {
		owners, _, err := s.ListOwners(ctx, g.ID, model.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("dumping owners of group %s: %w", g.ID, err)
		}
		for _, o := range owners {
			if idx, ok := userIDToIndex[o.ID]; ok {
				cfg.Ownerships = append(cfg.Ownerships, SeedOwnership{
					UserIndex:  idx,
					GroupIndex: gi,
				})
			}
		}
	}

	// Dump managers
	for ui, u := range users {
		mgr, err := s.GetManager(ctx, u.ID)
		if err != nil {
			// If the user simply doesn't have a manager, that's fine - just skip
			if errors.Is(err, store.ErrManagerNotFound) {
				continue
			}
			return nil, fmt.Errorf("dumping manager for user %s: %w", u.ID, err)
		}
		if mgr == nil {
			continue
		}
		if mi, ok := userIDToIndex[mgr.ID]; ok {
			cfg.Managers = append(cfg.Managers, SeedManager{
				UserIndex:    ui,
				ManagerIndex: mi,
			})
		}
	}

	return cfg, nil
}
