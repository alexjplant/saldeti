package seed

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/saldeti/saldeti/internal/entra/model"
	"github.com/saldeti/saldeti/internal/entra/store"
)

func TestDumpStore(t *testing.T) {
	// Seed a store with the default seed data
	s := store.NewMemoryStore()
	err := Seed(s)
	require.NoError(t, err)

	// Dump the store
	cfg, err := DumpStore(s)
	require.NoError(t, err)

	// Verify clients
	require.Len(t, cfg.Clients, 1)
	assert.Equal(t, "sim-client-id", cfg.Clients[0].ClientID)
	assert.Equal(t, "sim-client-secret", cfg.Clients[0].ClientSecret)
	assert.Equal(t, "sim-tenant-id", cfg.Clients[0].TenantID)

	// Verify users (admin + 10 sample)
	require.Len(t, cfg.Users, 11)

	// Find and verify admin user (don't assume specific index)
	adminFound := false
	for _, u := range cfg.Users {
		if u.Email == "admin@saldeti.local" {
			adminFound = true
			assert.Equal(t, "Admin User", u.DisplayName)
			assert.Equal(t, "Admin", u.GivenName)
			assert.Equal(t, "User", u.Surname)
			break
		}
	}
	assert.True(t, adminFound, "Admin user not found in dump")

	// Grace Lee should be disabled
	found := false
	for _, u := range cfg.Users {
		if u.Email == "grace.lee@saldeti.local" {
			found = true
			assert.NotNil(t, u.Enabled)
			assert.False(t, *u.Enabled)
		}
	}
	assert.True(t, found, "Grace Lee not found in dump")

	// Ivan Guest should be guest
	found = false
	for _, u := range cfg.Users {
		if u.Email == "ivan.guest@external.com" {
			found = true
			assert.True(t, u.IsGuest)
		}
	}
	assert.True(t, found, "Ivan Guest not found in dump")

	// Verify groups
	require.Len(t, cfg.Groups, 5)

	// Verify UPN-based memberships exist on groups (new-style dump)
	totalMemberUPNs := 0
	for _, g := range cfg.Groups {
		totalMemberUPNs += len(g.MemberUPNs)
	}
	assert.NotEmpty(t, totalMemberUPNs, "Expected member_upns in dumped groups")

	// Verify ManagerUPNs exist on users (new-style dump)
	managerCount := 0
	for _, u := range cfg.Users {
		if u.ManagerUPN != "" {
			managerCount++
		}
	}
	assert.NotZero(t, managerCount, "Expected manager_upn on at least some users")

	// Verify applications were dumped
	require.NotEmpty(t, cfg.Applications, "Expected applications in dump")
	assert.Equal(t, "Saldeti Simulator App", cfg.Applications[0].DisplayName)
	assert.Equal(t, "sim-client-id", cfg.Applications[0].AppID)
	require.Len(t, cfg.Applications[0].AppRoles, 1)
	assert.Equal(t, "Application.Read.All", cfg.Applications[0].AppRoles[0].Value)

	// Verify service principals were dumped
	assert.NotEmpty(t, cfg.ServicePrincipals, "Expected service_principals in dump")

	// Verify app role assignments were dumped
	assert.NotEmpty(t, cfg.AppRoleAssignments, "Expected app_role_assignments in dump")

	// Verify OAuth2 grants were dumped
	assert.NotEmpty(t, cfg.OAuth2Grants, "Expected oauth2_grants in dump")

	// Verify NO old-style index-based arrays are populated
	assert.Nil(t, cfg.Memberships, "Old-style memberships should not be populated")
	assert.Nil(t, cfg.Ownerships, "Old-style ownerships should not be populated")
	assert.Nil(t, cfg.Managers, "Old-style managers should not be populated")
}

func TestDumpRoundTrip(t *testing.T) {
	// Load the sample seed.json, seed a store, dump it, compare
	original, err := LoadFromFile("../../../examples/seed.json")
	require.NoError(t, err)

	s := store.NewMemoryStore()
	err = SeedFromConfig(s, original)
	require.NoError(t, err)

	dumped, err := DumpStore(s)
	require.NoError(t, err)

	// Same number of entities
	assert.Len(t, dumped.Clients, len(original.Clients))
	assert.Len(t, dumped.Users, len(original.Users))
	assert.Len(t, dumped.Groups, len(original.Groups))
	assert.Len(t, dumped.Applications, len(original.Applications))

	// Build maps for easier lookup
	originalUsersByEmail := make(map[string]SeedUser)
	for _, u := range original.Users {
		originalUsersByEmail[u.Email] = u
	}

	originalGroupsByName := make(map[string]SeedGroup)
	for _, g := range original.Groups {
		originalGroupsByName[g.DisplayName] = g
	}

	// Verify client data matches (order doesn't matter, just check all are present)
	dumpedClientsByID := make(map[string]SeedClient)
	for _, c := range dumped.Clients {
		dumpedClientsByID[c.ClientID] = c
	}
	for _, c := range original.Clients {
		dumped, ok := dumpedClientsByID[c.ClientID]
		assert.True(t, ok, "Client %s not found in dump", c.ClientID)
		if ok {
			assert.Equal(t, c.ClientSecret, dumped.ClientSecret)
			assert.Equal(t, c.TenantID, dumped.TenantID)
		}
	}

	// Verify user data matches (order doesn't matter, just check all are present)
	for _, u := range dumped.Users {
		orig, ok := originalUsersByEmail[u.Email]
		assert.True(t, ok, "User %s not found in original", u.Email)
		if ok {
			assert.Equal(t, orig.DisplayName, u.DisplayName)
			assert.Equal(t, orig.GivenName, u.GivenName)
			assert.Equal(t, orig.Surname, u.Surname)
			assert.Equal(t, orig.JobTitle, u.JobTitle)
			assert.Equal(t, orig.Department, u.Department)
			assert.Equal(t, orig.IsGuest, u.IsGuest)
			// Check enabled status matches
			if orig.Enabled != nil {
				require.NotNil(t, u.Enabled, "User %s: Expected non-nil Enabled", u.Email)
				assert.Equal(t, *orig.Enabled, *u.Enabled, "User %s: Enabled mismatch", u.Email)
			} else {
				if u.Enabled != nil {
					assert.True(t, *u.Enabled, "User %s: Expected Enabled to be true or nil", u.Email)
				}
			}
			// Check manager
			assert.Equal(t, orig.ManagerUPN, u.ManagerUPN, "User %s: manager_upn mismatch", u.Email)
			// Check licenses
			assert.Equal(t, orig.AssignedLicenses, u.AssignedLicenses, "User %s: licenses mismatch", u.Email)
		}
	}

	// Verify group data matches
	for _, g := range dumped.Groups {
		orig, ok := originalGroupsByName[g.DisplayName]
		assert.True(t, ok, "Group %s not found in original", g.DisplayName)
		if ok {
			assert.Equal(t, orig.Description, g.Description)
			assert.Equal(t, orig.Visibility, g.Visibility)
			// Sort both for comparison
			sortedOrig := make([]string, len(orig.MemberUPNs))
			copy(sortedOrig, orig.MemberUPNs)
			sort.Strings(sortedOrig)
			sortedDump := make([]string, len(g.MemberUPNs))
			copy(sortedDump, g.MemberUPNs)
			sort.Strings(sortedDump)
			assert.Equal(t, sortedOrig, sortedDump, "Group %s: member_upns mismatch", g.DisplayName)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	ctx := context.Background()

	// 1. Seed a store with the default Seed() data (which covers all entity types)
	s1 := store.NewMemoryStore()
	err := Seed(s1)
	require.NoError(t, err, "Seed() should succeed")

	// 2. Dump the store
	cfg, err := DumpStore(s1)
	require.NoError(t, err, "DumpStore() should succeed")

	// 3. Marshal to JSON and write to temp file
	data, err := json.MarshalIndent(cfg, "", "  ")
	require.NoError(t, err, "json.MarshalIndent should succeed")

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "roundtrip_seed.json")
	err = os.WriteFile(tmpFile, data, 0644)
	require.NoError(t, err, "writing temp file should succeed")

	// 4. Load from file
	loaded, err := LoadFromFile(tmpFile)
	require.NoError(t, err, "LoadFromFile should succeed")

	// 5. Seed a fresh store from the loaded config
	s2 := store.NewMemoryStore()
	err = SeedFromConfig(s2, loaded)
	require.NoError(t, err, "SeedFromConfig should succeed on round-tripped data")

	// ========== 6. Compare data between the two stores ==========

	// --- Compare users ---
	users1, _, err := s1.ListUsers(ctx, model.ListOptions{})
	require.NoError(t, err)
	users2, _, err := s2.ListUsers(ctx, model.ListOptions{})
	require.NoError(t, err)
	require.Len(t, users2, len(users1), "user count mismatch")

	users1ByUPN := make(map[string]model.User)
	for _, u := range users1 {
		users1ByUPN[u.UserPrincipalName] = u
	}
	users2ByUPN := make(map[string]model.User)
	for _, u := range users2 {
		users2ByUPN[u.UserPrincipalName] = u
	}

	for upn, u1 := range users1ByUPN {
		u2, ok := users2ByUPN[upn]
		require.True(t, ok, "User %s not found in round-tripped store", upn)
		assert.Equal(t, u1.DisplayName, u2.DisplayName, "User %s: displayName mismatch", upn)
		assert.Equal(t, u1.GivenName, u2.GivenName, "User %s: givenName mismatch", upn)
		assert.Equal(t, u1.Surname, u2.Surname, "User %s: surname mismatch", upn)
		assert.Equal(t, u1.JobTitle, u2.JobTitle, "User %s: jobTitle mismatch", upn)
		assert.Equal(t, u1.Department, u2.Department, "User %s: department mismatch", upn)
		assert.Equal(t, u1.UserType, u2.UserType, "User %s: userType mismatch", upn)

		// Compare enabled status
		if u1.AccountEnabled != nil {
			require.NotNil(t, u2.AccountEnabled, "User %s: accountEnabled nil", upn)
			assert.Equal(t, *u1.AccountEnabled, *u2.AccountEnabled, "User %s: accountEnabled mismatch", upn)
		}

		// Compare licenses
		assert.Len(t, u2.AssignedLicenses, len(u1.AssignedLicenses), "User %s: license count mismatch", upn)
		lics1 := make(map[string]model.AssignedLicense)
		for _, l := range u1.AssignedLicenses {
			lics1[l.SkuPartNumber] = l
		}
		for _, l2 := range u2.AssignedLicenses {
			l1, ok := lics1[l2.SkuPartNumber]
			require.True(t, ok, "User %s: unexpected license %s", upn, l2.SkuPartNumber)
			assert.ElementsMatch(t, l1.DisabledPlans, l2.DisabledPlans,
				"User %s: disabledPlans mismatch for %s", upn, l2.SkuPartNumber)
		}

		// Compare managers
		mgr1, err1 := s1.GetManager(ctx, u1.ID)
		mgr2, err2 := s2.GetManager(ctx, u2.ID)
		if err1 == nil && mgr1 != nil {
			require.NoError(t, err2, "User %s: manager not found in round-tripped store", upn)
			mgrUser1, _ := s1.GetUser(ctx, mgr1.ID)
			mgrUser2, _ := s2.GetUser(ctx, mgr2.ID)
			assert.Equal(t, mgrUser1.UserPrincipalName, mgrUser2.UserPrincipalName,
				"User %s: manager UPN mismatch", upn)
		} else {
			assert.True(t, errors.Is(err2, store.ErrManagerNotFound),
				"User %s: expected no manager in round-tripped store", upn)
		}
	}

	// --- Compare groups ---
	groups1, _, err := s1.ListGroups(ctx, model.ListOptions{})
	require.NoError(t, err)
	groups2, _, err := s2.ListGroups(ctx, model.ListOptions{})
	require.NoError(t, err)
	require.Len(t, groups2, len(groups1), "group count mismatch")

	groups1ByName := make(map[string]model.Group)
	for _, g := range groups1 {
		groups1ByName[g.DisplayName] = g
	}
	groups2ByName := make(map[string]model.Group)
	for _, g := range groups2 {
		groups2ByName[g.DisplayName] = g
	}

	for name, g1 := range groups1ByName {
		g2, ok := groups2ByName[name]
		require.True(t, ok, "Group %s not found in round-tripped store", name)
		assert.Equal(t, g1.Description, g2.Description, "Group %s: description mismatch", name)
		assert.Equal(t, g1.Visibility, g2.Visibility, "Group %s: visibility mismatch", name)

		// Compare members by UPN
		members1, _, _ := s1.ListMembers(ctx, g1.ID, model.ListOptions{})
		members2, _, _ := s2.ListMembers(ctx, g2.ID, model.ListOptions{})

		memberUPNs1 := make(map[string]bool)
		memberGroupNames1 := make(map[string]bool)
		for _, m := range members1 {
			if m.ODataType == "#microsoft.graph.user" {
				if u, err := s1.GetUser(ctx, m.ID); err == nil {
					memberUPNs1[u.UserPrincipalName] = true
				}
			} else if m.ODataType == "#microsoft.graph.group" {
				if grp, err := s1.GetGroup(ctx, m.ID); err == nil {
					memberGroupNames1[grp.DisplayName] = true
				}
			}
		}
		memberUPNs2 := make(map[string]bool)
		memberGroupNames2 := make(map[string]bool)
		for _, m := range members2 {
			if m.ODataType == "#microsoft.graph.user" {
				if u, err := s2.GetUser(ctx, m.ID); err == nil {
					memberUPNs2[u.UserPrincipalName] = true
				}
			} else if m.ODataType == "#microsoft.graph.group" {
				if grp, err := s2.GetGroup(ctx, m.ID); err == nil {
					memberGroupNames2[grp.DisplayName] = true
				}
			}
		}
		assert.Equal(t, memberUPNs1, memberUPNs2, "Group %s: user member UPNs mismatch", name)
		assert.Equal(t, memberGroupNames1, memberGroupNames2, "Group %s: group member names mismatch", name)
		// Compare SP members by AppID
		memberSPAppIDs1 := make(map[string]bool)
		for _, m := range members1 {
			if m.ODataType == "#microsoft.graph.servicePrincipal" {
				if sp, err := s1.GetServicePrincipal(ctx, m.ID); err == nil {
					memberSPAppIDs1[sp.AppID] = true
				}
			}
		}
		memberSPAppIDs2 := make(map[string]bool)
		for _, m := range members2 {
			if m.ODataType == "#microsoft.graph.servicePrincipal" {
				if sp, err := s2.GetServicePrincipal(ctx, m.ID); err == nil {
					memberSPAppIDs2[sp.AppID] = true
				}
			}
		}
		assert.Equal(t, memberSPAppIDs1, memberSPAppIDs2, "Group %s: SP member AppIDs mismatch", name)

		// Compare owners by UPN
		owners1, _, _ := s1.ListOwners(ctx, g1.ID, model.ListOptions{})
		owners2, _, _ := s2.ListOwners(ctx, g2.ID, model.ListOptions{})

		ownerUPNs1 := make(map[string]bool)
		for _, o := range owners1 {
			if u, err := s1.GetUser(ctx, o.ID); err == nil {
				ownerUPNs1[u.UserPrincipalName] = true
			}
		}
		ownerUPNs2 := make(map[string]bool)
		for _, o := range owners2 {
			if u, err := s2.GetUser(ctx, o.ID); err == nil {
				ownerUPNs2[u.UserPrincipalName] = true
			}
		}
		assert.Equal(t, ownerUPNs1, ownerUPNs2, "Group %s: owner UPNs mismatch", name)
	}

	// --- Compare applications ---
	apps1, _, err := s1.ListApplications(ctx, model.ListOptions{})
	require.NoError(t, err)
	apps2, _, err := s2.ListApplications(ctx, model.ListOptions{})
	require.NoError(t, err)
	require.Len(t, apps2, len(apps1), "application count mismatch")

	apps1ByAppID := make(map[string]model.Application)
	for _, a := range apps1 {
		apps1ByAppID[a.AppID] = a
	}
	for _, a2 := range apps2 {
		a1, ok := apps1ByAppID[a2.AppID]
		require.True(t, ok, "Application %s not found in original store", a2.AppID)
		assert.Equal(t, a1.DisplayName, a2.DisplayName)
		assert.Equal(t, a1.Description, a2.Description)
		assert.Equal(t, a1.SignInAudience, a2.SignInAudience)

		// Compare app roles by value
		assert.Len(t, a2.AppRoles, len(a1.AppRoles), "App %s: role count mismatch", a2.AppID)
		roles1ByValue := make(map[string]model.AppRole)
		for _, r := range a1.AppRoles {
			roles1ByValue[r.Value] = r
		}
		for _, r2 := range a2.AppRoles {
			r1, ok := roles1ByValue[r2.Value]
			require.True(t, ok, "App %s: unexpected role %s", a2.AppID, r2.Value)
			assert.Equal(t, r1.DisplayName, r2.DisplayName)
			assert.Equal(t, r1.Description, r2.Description)
			assert.ElementsMatch(t, r1.AllowedMemberTypes, r2.AllowedMemberTypes)
		}
	}

	// --- Compare app role assignments ---
	type assignmentKey struct {
		resourceAppID string
		roleValue     string
	}
	assignmentsByUPN1 := make(map[string][]assignmentKey)
	for _, u := range users1 {
		as, _, _ := s1.ListAppRoleAssignments(ctx, u.ID, model.ListOptions{})
		for _, a := range as {
			sp, _ := s1.GetServicePrincipal(ctx, a.ResourceID)
			var roleValue string
			for _, role := range sp.AppRoles {
				if role.ID == a.AppRoleID {
					roleValue = role.Value
					break
				}
			}
			assignmentsByUPN1[u.UserPrincipalName] = append(assignmentsByUPN1[u.UserPrincipalName],
				assignmentKey{sp.AppID, roleValue})
		}
	}
	assignmentsByUPN2 := make(map[string][]assignmentKey)
	for _, u := range users2 {
		as, _, _ := s2.ListAppRoleAssignments(ctx, u.ID, model.ListOptions{})
		for _, a := range as {
			sp, _ := s2.GetServicePrincipal(ctx, a.ResourceID)
			var roleValue string
			for _, role := range sp.AppRoles {
				if role.ID == a.AppRoleID {
					roleValue = role.Value
					break
				}
			}
			assignmentsByUPN2[u.UserPrincipalName] = append(assignmentsByUPN2[u.UserPrincipalName],
				assignmentKey{sp.AppID, roleValue})
		}
	}
	assert.Equal(t, assignmentsByUPN1, assignmentsByUPN2, "app role assignments mismatch")
	// --- Compare group-level app role assignments ---
	assignmentsByGroupName1 := make(map[string][]assignmentKey)
	for _, g := range groups1 {
		as, _, _ := s1.ListAppRoleAssignments(ctx, g.ID, model.ListOptions{})
		for _, a := range as {
			sp, _ := s1.GetServicePrincipal(ctx, a.ResourceID)
			var roleValue string
			for _, role := range sp.AppRoles {
				if role.ID == a.AppRoleID {
					roleValue = role.Value
					break
				}
			}
			assignmentsByGroupName1[g.DisplayName] = append(assignmentsByGroupName1[g.DisplayName],
				assignmentKey{sp.AppID, roleValue})
		}
	}
	assignmentsByGroupName2 := make(map[string][]assignmentKey)
	for _, g := range groups2 {
		as, _, _ := s2.ListAppRoleAssignments(ctx, g.ID, model.ListOptions{})
		for _, a := range as {
			sp, _ := s2.GetServicePrincipal(ctx, a.ResourceID)
			var roleValue string
			for _, role := range sp.AppRoles {
				if role.ID == a.AppRoleID {
					roleValue = role.Value
					break
				}
			}
			assignmentsByGroupName2[g.DisplayName] = append(assignmentsByGroupName2[g.DisplayName],
				assignmentKey{sp.AppID, roleValue})
		}
	}
	assert.Equal(t, assignmentsByGroupName1, assignmentsByGroupName2, "group app role assignments mismatch")

	// --- Compare OAuth2 permission grants ---
	type grantKey struct {
		clientAppID   string
		resourceAppID string
		scope         string
		consentType   string
		principalUPN  string
	}
	grants1, _, err := s1.ListOAuth2PermissionGrants(ctx, model.ListOptions{})
	require.NoError(t, err)
	grants2, _, err := s2.ListOAuth2PermissionGrants(ctx, model.ListOptions{})
	require.NoError(t, err)

	grantKeys1 := make(map[grantKey]bool)
	for _, g := range grants1 {
		clientSP, _ := s1.GetServicePrincipal(ctx, g.ClientID)
		resourceSP, _ := s1.GetServicePrincipal(ctx, g.ResourceID)
		var principalUPN string
		if g.PrincipalID != "" {
			if u, err := s1.GetUser(ctx, g.PrincipalID); err == nil {
				principalUPN = u.UserPrincipalName
			}
		}
		grantKeys1[grantKey{clientSP.AppID, resourceSP.AppID, g.Scope, g.ConsentType, principalUPN}] = true
	}
	grantKeys2 := make(map[grantKey]bool)
	for _, g := range grants2 {
		clientSP, _ := s2.GetServicePrincipal(ctx, g.ClientID)
		resourceSP, _ := s2.GetServicePrincipal(ctx, g.ResourceID)
		var principalUPN string
		if g.PrincipalID != "" {
			if u, err := s2.GetUser(ctx, g.PrincipalID); err == nil {
				principalUPN = u.UserPrincipalName
			}
		}
		grantKeys2[grantKey{clientSP.AppID, resourceSP.AppID, g.Scope, g.ConsentType, principalUPN}] = true
	}
	assert.Equal(t, grantKeys1, grantKeys2, "OAuth2 permission grants mismatch")
}
