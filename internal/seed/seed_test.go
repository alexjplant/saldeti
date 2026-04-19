package seed

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saldeti/saldeti/internal/model"
	"github.com/saldeti/saldeti/internal/store"
)

func TestLoadFromFile(t *testing.T) {
	// Create a temporary JSON file with a minimal valid config
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_seed.json")

	jsonContent := `{
		"clients": [
			{
				"client_id": "test-client-id",
				"client_secret": "test-client-secret",
				"tenant_id": "test-tenant-id"
			}
		],
		"users": [
			{
				"email": "test@example.com",
				"display_name": "Test User",
				"password": "TestPassword123!",
				"given_name": "Test",
				"surname": "User",
				"enabled": true,
				"is_guest": false
			}
		],
		"groups": [
			{
				"display_name": "Test Group",
				"description": "A test group",
				"mail_nickname": "testgroup",
				"visibility": "Public"
			}
		],
		"memberships": [
			{
				"user_index": 0,
				"group_index": 0
			}
		],
		"managers": [
			{
				"user_index": 0,
				"manager_index": 0
			}
		]
	}`

	if err := os.WriteFile(tmpFile, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Load the config
	cfg, err := LoadFromFile(tmpFile)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	// Verify clients
	if len(cfg.Clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(cfg.Clients))
	}
	if cfg.Clients[0].ClientID != "test-client-id" {
		t.Errorf("Expected client_id 'test-client-id', got '%s'", cfg.Clients[0].ClientID)
	}
	if cfg.Clients[0].ClientSecret != "test-client-secret" {
		t.Errorf("Expected client_secret 'test-client-secret', got '%s'", cfg.Clients[0].ClientSecret)
	}
	if cfg.Clients[0].TenantID != "test-tenant-id" {
		t.Errorf("Expected tenant_id 'test-tenant-id', got '%s'", cfg.Clients[0].TenantID)
	}

	// Verify users
	if len(cfg.Users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(cfg.Users))
	}
	if cfg.Users[0].Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", cfg.Users[0].Email)
	}
	if cfg.Users[0].DisplayName != "Test User" {
		t.Errorf("Expected display_name 'Test User', got '%s'", cfg.Users[0].DisplayName)
	}
	if cfg.Users[0].Password != "TestPassword123!" {
		t.Errorf("Expected password 'TestPassword123!', got '%s'", cfg.Users[0].Password)
	}
	if cfg.Users[0].GivenName != "Test" {
		t.Errorf("Expected given_name 'Test', got '%s'", cfg.Users[0].GivenName)
	}
	if cfg.Users[0].Surname != "User" {
		t.Errorf("Expected surname 'User', got '%s'", cfg.Users[0].Surname)
	}
	if cfg.Users[0].Enabled == nil || *cfg.Users[0].Enabled != true {
		t.Errorf("Expected enabled true, got %v", cfg.Users[0].Enabled)
	}
	if cfg.Users[0].IsGuest != false {
		t.Errorf("Expected is_guest false, got %v", cfg.Users[0].IsGuest)
	}

	// Verify groups
	if len(cfg.Groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(cfg.Groups))
	}
	if cfg.Groups[0].DisplayName != "Test Group" {
		t.Errorf("Expected display_name 'Test Group', got '%s'", cfg.Groups[0].DisplayName)
	}
	if cfg.Groups[0].Description != "A test group" {
		t.Errorf("Expected description 'A test group', got '%s'", cfg.Groups[0].Description)
	}
	if cfg.Groups[0].MailNickname != "testgroup" {
		t.Errorf("Expected mail_nickname 'testgroup', got '%s'", cfg.Groups[0].MailNickname)
	}
	if cfg.Groups[0].Visibility != "Public" {
		t.Errorf("Expected visibility 'Public', got '%s'", cfg.Groups[0].Visibility)
	}

	// Verify memberships
	if len(cfg.Memberships) != 1 {
		t.Errorf("Expected 1 membership, got %d", len(cfg.Memberships))
	}
	if cfg.Memberships[0].UserIndex == nil || *cfg.Memberships[0].UserIndex != 0 {
		t.Errorf("Expected user_index 0, got %v", cfg.Memberships[0].UserIndex)
	}
	if cfg.Memberships[0].GroupIndex == nil || *cfg.Memberships[0].GroupIndex != 0 {
		t.Errorf("Expected group_index 0, got %v", cfg.Memberships[0].GroupIndex)
	}

	// Verify managers
	if len(cfg.Managers) != 1 {
		t.Errorf("Expected 1 manager, got %d", len(cfg.Managers))
	}
	if cfg.Managers[0].UserIndex != 0 {
		t.Errorf("Expected user_index 0, got %d", cfg.Managers[0].UserIndex)
	}
	if cfg.Managers[0].ManagerIndex != 0 {
		t.Errorf("Expected manager_index 0, got %d", cfg.Managers[0].ManagerIndex)
	}
}

func TestLoadFromFileValidation(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "missing client_id",
			json: `{
				"clients": [
					{
						"client_secret": "test-secret",
						"tenant_id": "test-tenant"
					}
				],
				"users": [],
				"groups": [],
				"memberships": [],
				"managers": []
			}`,
			wantErr: true,
			errMsg:  "client_id is required",
		},
		{
			name: "missing client_secret",
			json: `{
				"clients": [
					{
						"client_id": "test-id",
						"tenant_id": "test-tenant"
					}
				],
				"users": [],
				"groups": [],
				"memberships": [],
				"managers": []
			}`,
			wantErr: true,
			errMsg:  "client_secret is required",
		},
		{
			name: "missing tenant_id",
			json: `{
				"clients": [
					{
						"client_id": "test-id",
						"client_secret": "test-secret"
					}
				],
				"users": [],
				"groups": [],
				"memberships": [],
				"managers": []
			}`,
			wantErr: true,
			errMsg:  "tenant_id is required",
		},
		{
			name: "no clients - now allowed (data-only seed)",
			json: `{
				"clients": [],
				"users": [],
				"groups": [],
				"memberships": [],
				"managers": []
			}`,
			wantErr: false,
		},
		{
			name: "missing user email",
			json: `{
				"clients": [
					{
						"client_id": "test-id",
						"client_secret": "test-secret",
						"tenant_id": "test-tenant"
					}
				],
				"users": [
					{
						"display_name": "Test User",
						"password": "TestPassword123!"
					}
				],
				"groups": [],
				"memberships": [],
				"managers": []
			}`,
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name: "missing user display_name",
			json: `{
				"clients": [
					{
						"client_id": "test-id",
						"client_secret": "test-secret",
						"tenant_id": "test-tenant"
					}
				],
				"users": [
					{
						"email": "test@example.com",
						"password": "TestPassword123!"
					}
				],
				"groups": [],
				"memberships": [],
				"managers": []
			}`,
			wantErr: true,
			errMsg:  "display_name is required",
		},
		{
			name: "missing user password",
			json: `{
				"clients": [
					{
						"client_id": "test-id",
						"client_secret": "test-secret",
						"tenant_id": "test-tenant"
					}
				],
				"users": [
					{
						"email": "test@example.com",
						"display_name": "Test User"
					}
				],
				"groups": [],
				"memberships": [],
				"managers": []
			}`,
			wantErr: true,
			errMsg:  "password is required",
		},
		{
			name: "missing group display_name",
			json: `{
				"clients": [
					{
						"client_id": "test-id",
						"client_secret": "test-secret",
						"tenant_id": "test-tenant"
					}
				],
				"users": [],
				"groups": [
					{
						"description": "A test group"
					}
				],
				"memberships": [],
				"managers": []
			}`,
			wantErr: true,
			errMsg:  "display_name is required",
		},
		{
			name: "out of range membership user_index",
			json: `{
				"clients": [
					{
						"client_id": "test-id",
						"client_secret": "test-secret",
						"tenant_id": "test-tenant"
					}
				],
				"users": [
					{
						"email": "test@example.com",
						"display_name": "Test User",
						"password": "TestPassword123!"
					}
				],
				"groups": [
					{
						"display_name": "Test Group"
					}
				],
				"memberships": [
					{
						"user_index": 5,
						"group_index": 0
					}
				],
				"managers": []
			}`,
			wantErr: true,
			errMsg:  "user_index 5 is out of range",
		},
		{
			name: "out of range membership group_index",
			json: `{
				"clients": [
					{
						"client_id": "test-id",
						"client_secret": "test-secret",
						"tenant_id": "test-tenant"
					}
				],
				"users": [
					{
						"email": "test@example.com",
						"display_name": "Test User",
						"password": "TestPassword123!"
					}
				],
				"groups": [
					{
						"display_name": "Test Group"
					}
				],
				"memberships": [
					{
						"user_index": 0,
						"group_index": 5
					}
				],
				"managers": []
			}`,
			wantErr: true,
			errMsg:  "group_index 5 is out of range",
		},
		{
			name: "membership without user_index or member_group_index",
			json: `{
				"clients": [
					{
						"client_id": "test-id",
						"client_secret": "test-secret",
						"tenant_id": "test-tenant"
					}
				],
				"users": [
					{
						"email": "test@example.com",
						"display_name": "Test User",
						"password": "TestPassword123!"
					}
				],
				"groups": [
					{
						"display_name": "Test Group"
					}
				],
				"memberships": [
					{
						"group_index": 0
					}
				],
				"managers": []
			}`,
			wantErr: true,
			errMsg:  "either user_index or member_group_index must be set",
		},
		{
			name: "out of range manager user_index",
			json: `{
				"clients": [
					{
						"client_id": "test-id",
						"client_secret": "test-secret",
						"tenant_id": "test-tenant"
					}
				],
				"users": [
					{
						"email": "test@example.com",
						"display_name": "Test User",
						"password": "TestPassword123!"
					}
				],
				"groups": [],
				"memberships": [],
				"managers": [
					{
						"user_index": 5,
						"manager_index": 0
					}
				]
			}`,
			wantErr: true,
			errMsg:  "user_index 5 is out of range",
		},
		{
			name: "out of range manager manager_index",
			json: `{
				"clients": [
					{
						"client_id": "test-id",
						"client_secret": "test-secret",
						"tenant_id": "test-tenant"
					}
				],
				"users": [
					{
						"email": "test@example.com",
						"display_name": "Test User",
						"password": "TestPassword123!"
					}
				],
				"groups": [],
				"memberships": [],
				"managers": [
					{
						"user_index": 0,
						"manager_index": 5
					}
				]
			}`,
			wantErr: true,
			errMsg:  "manager_index 5 is out of range",
		},
		{
			name: "invalid JSON",
			json: `{ invalid json }`,
			wantErr: true,
			errMsg:  "failed to parse",
		},
		{
			name: "file not found",
			json: "",
			wantErr: true,
			errMsg:  "failed to read seed file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tmpFile string
			var err error

			if tt.json != "" {
				tmpDir := t.TempDir()
				tmpFile = filepath.Join(tmpDir, "test_seed.json")
				err = os.WriteFile(tmpFile, []byte(tt.json), 0644)
				if err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
			} else {
				tmpFile = filepath.Join(t.TempDir(), "nonexistent.json")
			}

			_, err = LoadFromFile(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("LoadFromFile() error = %v, expected to contain %q", err, tt.errMsg)
			}
		})
	}
}

func TestSeedFromConfig(t *testing.T) {
	// Create a memory store
	s := store.NewMemoryStore()

	// Create a minimal config
	trueVal := true
	cfg := &SeedConfig{
		Clients: []SeedClient{
			{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				TenantID:     "test-tenant-id",
			},
		},
		Users: []SeedUser{
			{
				Email:       "test@example.com",
				DisplayName: "Test User",
				Password:    "TestPassword123!",
				Enabled:     &trueVal,
				IsGuest:     false,
			},
		},
		Groups: []SeedGroup{
			{
				DisplayName:  "Test Group",
				Description:  "A test group",
				MailNickname: "testgroup",
				Visibility:   "Public",
			},
		},
		Memberships: []SeedMembership{
			{
				UserIndex:  intPtr(0),
				GroupIndex: intPtr(0),
			},
		},
		Managers: []SeedManager{
			{
				UserIndex:    0,
				ManagerIndex: 0,
			},
		},
	}

	// Seed the store
	err := SeedFromConfig(s, cfg)
	if err != nil {
		t.Fatalf("SeedFromConfig() failed: %v", err)
	}

	// Verify client was registered
	clientID, clientSecret, tenantID, err := s.GetClient(nil, "test-client-id")
	if err != nil {
		t.Errorf("Failed to get client: %v", err)
	}
	if clientID != "test-client-id" {
		t.Errorf("Expected client_id 'test-client-id', got '%s'", clientID)
	}
	if clientSecret != "test-client-secret" {
		t.Errorf("Expected client_secret 'test-client-secret', got '%s'", clientSecret)
	}
	if tenantID != "test-tenant-id" {
		t.Errorf("Expected tenant_id 'test-tenant-id', got '%s'", tenantID)
	}

	// Verify user was created
	user, err := s.GetUserByUPN(nil, "test@example.com")
	if err != nil {
		t.Errorf("Failed to get user by UPN: %v", err)
	}
	if user.DisplayName != "Test User" {
		t.Errorf("Expected display_name 'Test User', got '%s'", user.DisplayName)
	}
	if user.UserType != "Member" {
		t.Errorf("Expected user_type 'Member', got '%s'", user.UserType)
	}

	// Verify group was created
	groups, _, err := s.ListGroups(nil, model.ListOptions{})
	if err != nil {
		t.Errorf("Failed to list groups: %v", err)
	}
	if len(groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(groups))
	}
	if groups[0].DisplayName != "Test Group" {
		t.Errorf("Expected group display_name 'Test Group', got '%s'", groups[0].DisplayName)
	}

	// Verify membership was created
	members, _, err := s.ListMembers(nil, groups[0].ID, model.ListOptions{})
	if err != nil {
		t.Errorf("Failed to list members: %v", err)
	}
	if len(members) != 1 {
		t.Errorf("Expected 1 member, got %d", len(members))
	}
	if members[0].ID != user.ID {
		t.Errorf("Expected member ID %s, got %s", user.ID, members[0].ID)
	}

	// Verify manager was set
	manager, err := s.GetManager(nil, user.ID)
	if err != nil {
		t.Errorf("Failed to get manager: %v", err)
	}
	if manager.ID != user.ID {
		t.Errorf("Expected manager ID %s, got %s", user.ID, manager.ID)
	}
}

func TestSeedFromConfig_GuestUser(t *testing.T) {
	s := store.NewMemoryStore()

	falseVal := false
	cfg := &SeedConfig{
		Clients: []SeedClient{
			{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				TenantID:     "test-tenant-id",
			},
		},
		Users: []SeedUser{
			{
				Email:       "guest@external.com",
				DisplayName: "Guest User",
				Password:    "TestPassword123!",
				Enabled:     &falseVal,
				IsGuest:     true,
			},
		},
	}

	err := SeedFromConfig(s, cfg)
	if err != nil {
		t.Fatalf("SeedFromConfig() failed: %v", err)
	}

	// Verify user was created with UserType = "Guest"
	user, err := s.GetUserByUPN(nil, "guest@external.com")
	if err != nil {
		t.Errorf("Failed to get user by UPN: %v", err)
	}
	if user.UserType != "Guest" {
		t.Errorf("Expected user_type 'Guest', got '%s'", user.UserType)
	}
}

func TestSeedFromConfig_WithDisabledUser(t *testing.T) {
	s := store.NewMemoryStore()

	falseVal := false
	cfg := &SeedConfig{
		Clients: []SeedClient{
			{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				TenantID:     "test-tenant-id",
			},
		},
		Users: []SeedUser{
			{
				Email:       "disabled@example.com",
				DisplayName: "Disabled User",
				Password:    "TestPassword123!",
				Enabled:     &falseVal,
				IsGuest:     false,
			},
		},
	}

	err := SeedFromConfig(s, cfg)
	if err != nil {
		t.Fatalf("SeedFromConfig() failed: %v", err)
	}

	// Verify user was created with AccountEnabled = false
	user, err := s.GetUserByUPN(nil, "disabled@example.com")
	if err != nil {
		t.Errorf("Failed to get user by UPN: %v", err)
	}
	if user.AccountEnabled == nil || *user.AccountEnabled != false {
		t.Errorf("Expected AccountEnabled false, got %v", user.AccountEnabled)
	}
}

func TestSeedBackwardCompat(t *testing.T) {
	// Create a memory store
	s := store.NewMemoryStore()

	// Call the existing Seed() function
	err := Seed(s)
	if err != nil {
		t.Fatalf("Seed() failed: %v", err)
	}

	// Verify admin user exists
	admin, err := s.GetUserByUPN(nil, "admin@saldeti.local")
	if err != nil {
		t.Errorf("Failed to get admin user: %v", err)
	}
	if admin.DisplayName != "Admin User" {
		t.Errorf("Expected admin display_name 'Admin User', got '%s'", admin.DisplayName)
	}

	// Verify expected users exist (at least a few of them)
	expectedUsers := []string{
		"alice.smith@saldeti.local",
		"bob.jones@saldeti.local",
		"charlie.brown@saldeti.local",
		"ivan.guest@external.com",
	}
	for _, email := range expectedUsers {
		user, err := s.GetUserByUPN(nil, email)
		if err != nil {
			t.Errorf("Failed to get user %s: %v", email, err)
		}
		if user == nil {
			t.Errorf("Expected user %s to exist", email)
		}
	}

	// Verify guest user has correct type
	ivan, err := s.GetUserByUPN(nil, "ivan.guest@external.com")
	if err != nil {
		t.Errorf("Failed to get guest user: %v", err)
	}
	if ivan.UserType != "Guest" {
		t.Errorf("Expected guest user_type 'Guest', got '%s'", ivan.UserType)
	}

	// Verify disabled user is disabled
	grace, err := s.GetUserByUPN(nil, "grace.lee@saldeti.local")
	if err != nil {
		t.Errorf("Failed to get Grace user: %v", err)
	}
	if grace.AccountEnabled == nil || *grace.AccountEnabled {
		t.Errorf("Expected Grace to be disabled, got %v", grace.AccountEnabled)
	}

	// Verify expected groups exist
	groups, _, err := s.ListGroups(nil, model.ListOptions{})
	if err != nil {
		t.Errorf("Failed to list groups: %v", err)
	}
	expectedGroups := []string{
		"Engineering Team",
		"Marketing Team",
		"All Staff",
		"Leadership",
		"Project Alpha",
	}
	groupMap := make(map[string]bool)
	for _, g := range groups {
		groupMap[g.DisplayName] = true
	}
	for _, name := range expectedGroups {
		if !groupMap[name] {
			t.Errorf("Expected group '%s' to exist", name)
		}
	}

	// Verify client was registered
	_, _, _, err = s.GetClient(nil, "sim-client-id")
	if err != nil {
		t.Errorf("Failed to get client: %v", err)
	}

	// Verify memberships exist
	engineeringGroup, err := findGroupByName(groups, "Engineering Team")
	if err != nil {
		t.Errorf("Failed to find Engineering Team group: %v", err)
	}
	members, _, err := s.ListMembers(nil, engineeringGroup.ID, model.ListOptions{})
	if err != nil {
		t.Errorf("Failed to list Engineering Team members: %v", err)
	}
	// Engineering Team should have at least alice, bob, eve, grace
	if len(members) < 4 {
		t.Errorf("Expected at least 4 members in Engineering Team, got %d", len(members))
	}

	// Verify managers exist
	eve, err := s.GetUserByUPN(nil, "eve.wilson@saldeti.local")
	if err != nil {
		t.Errorf("Failed to get Eve user: %v", err)
	}
	alice, err := s.GetUserByUPN(nil, "alice.smith@saldeti.local")
	if err != nil {
		t.Errorf("Failed to get Alice user: %v", err)
	}
	manager, err := s.GetManager(nil, alice.ID)
	if err != nil {
		t.Errorf("Failed to get Alice's manager: %v", err)
	}
	if manager.ID != eve.ID {
		t.Errorf("Expected Alice's manager to be Eve, got %s", manager.ID)
	}
}

func findGroupByName(groups []model.Group, name string) (model.Group, error) {
	for _, g := range groups {
		if g.DisplayName == name {
			return g, nil
		}
	}
	return model.Group{}, fmt.Errorf("group %s not found", name)
}

func TestSeedFromConfig_Idempotent(t *testing.T) {
	s := store.NewMemoryStore()

	trueVal := true
	cfg := &SeedConfig{
		Clients: []SeedClient{
			{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				TenantID:     "test-tenant-id",
			},
		},
		Users: []SeedUser{
			{
				Email:       "test@example.com",
				DisplayName: "Test User",
				Password:    "TestPassword123!",
				Enabled:     &trueVal,
				IsGuest:     false,
			},
		},
	}

	// Seed once
	err := SeedFromConfig(s, cfg)
	if err != nil {
		t.Fatalf("First SeedFromConfig() failed: %v", err)
	}

	// Seed again - should not error
	err = SeedFromConfig(s, cfg)
	if err != nil {
		t.Errorf("Second SeedFromConfig() failed (should be idempotent): %v", err)
	}
}
