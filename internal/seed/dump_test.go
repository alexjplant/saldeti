package seed

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/saldeti/saldeti/internal/store"
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

	// Verify there are memberships
	assert.NotEmpty(t, cfg.Memberships, "Expected memberships in dump")

	// Verify managers
	assert.NotEmpty(t, cfg.Managers, "Expected managers in dump")
}

func TestDumpRoundTrip(t *testing.T) {
	// Load the sample seed.json, seed a store, dump it, compare
	original, err := LoadFromFile("../../seed.json")
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
	assert.Len(t, dumped.Memberships, len(original.Memberships))
	assert.Len(t, dumped.Managers, len(original.Managers))

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
			// Check enabled status matches (accountEnabled may be nil, but we compare the actual values)
			if orig.Enabled != nil {
				require.NotNil(t, u.Enabled, "User %s: Expected non-nil Enabled", u.Email)
				assert.Equal(t, *orig.Enabled, *u.Enabled, "User %s: Enabled mismatch", u.Email)
			} else {
				// In original, nil means default true
				if u.Enabled != nil {
					assert.True(t, *u.Enabled, "User %s: Expected Enabled to be true or nil", u.Email)
				}
			}
		}
	}

	// Verify group data matches (order doesn't matter, just check all are present)
	for _, g := range dumped.Groups {
		orig, ok := originalGroupsByName[g.DisplayName]
		assert.True(t, ok, "Group %s not found in original", g.DisplayName)
		if ok {
			assert.Equal(t, orig.Description, g.Description)
			assert.Equal(t, orig.Visibility, g.Visibility)
		}
	}
}
