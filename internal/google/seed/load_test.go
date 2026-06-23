package seed

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saldeti/saldeti/internal/google/model"
	"github.com/saldeti/saldeti/internal/google/store"
)

func TestGoogleLoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_google_seed.json")

	jsonContent := `{
		"clients": [
			{
				"client_id": "test-google-client-id",
				"client_secret": "test-google-client-secret"
			}
		],
		"users": [
			{
				"primary_email": "admin@example.com",
				"given_name": "Admin",
				"family_name": "User",
				"password": "TestPassword123!",
				"is_admin": true,
				"org_unit_path": "/"
			}
		],
		"groups": [
			{
				"email": "team@example.com",
				"name": "Team Group",
				"description": "A test group",
				"member_emails": ["admin@example.com"]
			}
		],
		"org_units": [
			{
				"name": "Engineering",
				"description": "Engineering department",
				"parent_org_unit_path": "/"
			}
		],
		"roles": [
			{
				"role_name": "HelpdeskAdmin",
				"role_description": "Helpdesk admin",
				"privileges": ["USERS_READ"]
			}
		],
		"role_assignments": [
			{
				"assigned_to_email": "admin@example.com",
				"role_id_or_name": "HelpdeskAdmin"
			}
		],
		"domains": [
			{
				"domain_name": "example.com",
				"is_primary": true
			}
		],
		"chromeos_devices": [
			{
				"serial_number": "CHROME-001",
				"annotated_user": "admin@example.com",
				"org_unit_path": "/",
				"notes": "Test Chromebook"
			}
		],
		"mobile_devices": [
			{
				"serial_number": "MOB-001",
				"model": "Pixel 8",
				"os": "Android 14",
				"status": "ACTIVE"
			}
		]
	}`

	if err := os.WriteFile(tmpFile, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cfg, err := LoadFromFile(tmpFile)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	// Verify clients
	if len(cfg.Clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(cfg.Clients))
	}
	if cfg.Clients[0].ClientID != "test-google-client-id" {
		t.Errorf("Expected client_id 'test-google-client-id', got '%s'", cfg.Clients[0].ClientID)
	}
	if cfg.Clients[0].ClientSecret != "test-google-client-secret" {
		t.Errorf("Expected client_secret 'test-google-client-secret', got '%s'", cfg.Clients[0].ClientSecret)
	}

	// Verify users
	if len(cfg.Users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(cfg.Users))
	}
	if cfg.Users[0].PrimaryEmail != "admin@example.com" {
		t.Errorf("Expected primary_email 'admin@example.com', got '%s'", cfg.Users[0].PrimaryEmail)
	}
	if cfg.Users[0].GivenName != "Admin" {
		t.Errorf("Expected given_name 'Admin', got '%s'", cfg.Users[0].GivenName)
	}
	if cfg.Users[0].FamilyName != "User" {
		t.Errorf("Expected family_name 'User', got '%s'", cfg.Users[0].FamilyName)
	}
	if cfg.Users[0].IsAdmin != true {
		t.Errorf("Expected is_admin true, got %v", cfg.Users[0].IsAdmin)
	}

	// Verify groups
	if len(cfg.Groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(cfg.Groups))
	}
	if cfg.Groups[0].Email != "team@example.com" {
		t.Errorf("Expected group email 'team@example.com', got '%s'", cfg.Groups[0].Email)
	}
	if len(cfg.Groups[0].MemberEmails) != 1 {
		t.Errorf("Expected 1 member_email, got %d", len(cfg.Groups[0].MemberEmails))
	}

	// Verify org_units
	if len(cfg.OrgUnits) != 1 {
		t.Errorf("Expected 1 org_unit, got %d", len(cfg.OrgUnits))
	}
	if cfg.OrgUnits[0].Name != "Engineering" {
		t.Errorf("Expected org_unit name 'Engineering', got '%s'", cfg.OrgUnits[0].Name)
	}

	// Verify roles
	if len(cfg.Roles) != 1 {
		t.Errorf("Expected 1 role, got %d", len(cfg.Roles))
	}
	if cfg.Roles[0].RoleName != "HelpdeskAdmin" {
		t.Errorf("Expected role_name 'HelpdeskAdmin', got '%s'", cfg.Roles[0].RoleName)
	}

	// Verify role_assignments
	if len(cfg.RoleAssignments) != 1 {
		t.Errorf("Expected 1 role_assignment, got %d", len(cfg.RoleAssignments))
	}
	if cfg.RoleAssignments[0].AssignedToEmail != "admin@example.com" {
		t.Errorf("Expected assigned_to_email 'admin@example.com', got '%s'", cfg.RoleAssignments[0].AssignedToEmail)
	}

	// Verify domains
	if len(cfg.Domains) != 1 {
		t.Errorf("Expected 1 domain, got %d", len(cfg.Domains))
	}
	if cfg.Domains[0].DomainName != "example.com" {
		t.Errorf("Expected domain_name 'example.com', got '%s'", cfg.Domains[0].DomainName)
	}

	// Verify chromeos_devices
	if len(cfg.ChromeOSDevices) != 1 {
		t.Errorf("Expected 1 chromeos_device, got %d", len(cfg.ChromeOSDevices))
	}
	if cfg.ChromeOSDevices[0].SerialNumber != "CHROME-001" {
		t.Errorf("Expected serial_number 'CHROME-001', got '%s'", cfg.ChromeOSDevices[0].SerialNumber)
	}

	// Verify mobile_devices
	if len(cfg.MobileDevices) != 1 {
		t.Errorf("Expected 1 mobile_device, got %d", len(cfg.MobileDevices))
	}
	if cfg.MobileDevices[0].SerialNumber != "MOB-001" {
		t.Errorf("Expected serial_number 'MOB-001', got '%s'", cfg.MobileDevices[0].SerialNumber)
	}
}

func TestGoogleLoadFromFileValidation(t *testing.T) {
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
						"client_secret": "test-secret"
					}
				],
				"users": []
			}`,
			wantErr: true,
			errMsg:  "client_id is required",
		},
		{
			name: "missing client_secret",
			json: `{
				"clients": [
					{
						"client_id": "test-id"
					}
				],
				"users": []
			}`,
			wantErr: true,
			errMsg:  "client_secret is required",
		},
		{
			name: "missing user primary_email",
			json: `{
				"clients": [
					{
						"client_id": "test-id",
						"client_secret": "test-secret"
					}
				],
				"users": [
					{
						"given_name": "Test"
					}
				]
			}`,
			wantErr: true,
			errMsg:  "primary_email is required",
		},
		{
			name: "missing group email",
			json: `{
				"clients": [
					{
						"client_id": "test-id",
						"client_secret": "test-secret"
					}
				],
				"groups": [
					{
						"name": "Test Group"
					}
				]
			}`,
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name: "missing org_unit name",
			json: `{
				"clients": [{"client_id": "t", "client_secret": "s"}],
				"org_units": [
					{
						"description": "no name"
					}
				]
			}`,
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing role_name",
			json: `{
				"clients": [{"client_id": "t", "client_secret": "s"}],
				"roles": [
					{
						"role_description": "no name"
					}
				]
			}`,
			wantErr: true,
			errMsg:  "role_name is required",
		},
		{
			name: "missing assigned_to_email in role_assignment",
			json: `{
				"clients": [{"client_id": "t", "client_secret": "s"}],
				"role_assignments": [
					{
						"role_id_or_name": "some-role"
					}
				]
			}`,
			wantErr: true,
			errMsg:  "assigned_to_email is required",
		},
		{
			name: "missing role_id_or_name in role_assignment",
			json: `{
				"clients": [{"client_id": "t", "client_secret": "s"}],
				"role_assignments": [
					{
						"assigned_to_email": "user@example.com"
					}
				]
			}`,
			wantErr: true,
			errMsg:  "role_id_or_name is required",
		},
		{
			name: "missing domain_name",
			json: `{
				"clients": [{"client_id": "t", "client_secret": "s"}],
				"domains": [
					{
						"is_primary": true
					}
				]
			}`,
			wantErr: true,
			errMsg:  "domain_name is required",
		},
		{
			name: "missing chromeos_device serial_number",
			json: `{
				"clients": [{"client_id": "t", "client_secret": "s"}],
				"chromeos_devices": [
					{
						"notes": "no serial"
					}
				]
			}`,
			wantErr: true,
			errMsg:  "serial_number is required",
		},
		{
			name: "missing mobile_device serial_number",
			json: `{
				"clients": [{"client_id": "t", "client_secret": "s"}],
				"mobile_devices": [
					{
						"model": "Pixel 8"
					}
				]
			}`,
			wantErr: true,
			errMsg:  "serial_number is required",
		},
		{
			name: "member_emails references non-existent user",
			json: `{
				"clients": [{"client_id": "t", "client_secret": "s"}],
				"users": [
					{
						"primary_email": "admin@example.com"
					}
				],
				"groups": [
					{
						"email": "team@example.com",
						"member_emails": ["nonexistent@example.com"]
					}
				]
			}`,
			wantErr: true,
			errMsg:  "member_emails",
		},
		{
			name: "assigned_to_email references non-existent user",
			json: `{
				"clients": [{"client_id": "t", "client_secret": "s"}],
				"role_assignments": [
					{
						"assigned_to_email": "nonexistent@example.com",
						"role_id_or_name": "HelpdeskAdmin"
					}
				]
			}`,
			wantErr: true,
			errMsg:  "assigned_to_email",
		},
		{
			name: "missing group_email in group_settings",
			json: `{
				"clients": [{"client_id": "t", "client_secret": "s"}],
				"groups": [
					{"email": "team@example.com"}
				],
				"group_settings": [
					{"who_can_post_message": "ALL_MEMBERS_CAN_POST"}
				]
			}`,
			wantErr: true,
			errMsg:  "group_email is required",
		},
		{
			name: "group_settings references non-existent group",
			json: `{
				"clients": [{"client_id": "t", "client_secret": "s"}],
				"groups": [
					{"email": "team@example.com"}
				],
				"group_settings": [
					{"group_email": "nonexistent@example.com", "who_can_post_message": "ALL_MEMBERS_CAN_POST"}
				]
			}`,
			wantErr: true,
			errMsg:  "does not reference any group",
		},
		{
			name: "valid minimal config",
			json: `{
				"clients": [{"client_id": "t", "client_secret": "s"}]
			}`,
			wantErr: false,
		},
		{
			name:    "file not found",
			json:    "",
			wantErr: true,
			errMsg:  "failed to read",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tmpFile string
			var err error

			if tt.json != "" {
				tmpDir := t.TempDir()
				tmpFile = filepath.Join(tmpDir, "test_google_seed.json")
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

func TestGoogleSeedFromConfig(t *testing.T) {
	s := store.NewMemoryStore()

	cfg := &GoogleSeedConfig{
		Clients: []GoogleSeedClient{
			{
				ClientID:     "test-google-client-id",
				ClientSecret: "test-google-client-secret",
			},
		},
		Users: []GoogleSeedUser{
			{
				PrimaryEmail: "admin@example.com",
				GivenName:    "Admin",
				FamilyName:   "User",
				Password:     "TestPassword123!",
				IsAdmin:      true,
				OrgUnitPath:  "/",
			},
			{
				PrimaryEmail: "alice@example.com",
				GivenName:    "Alice",
				FamilyName:   "Smith",
				Password:     "TestPassword123!",
				OrgUnitPath:  "/Engineering",
			},
		},
		Groups: []GoogleSeedGroup{
			{
				Email:        "team@example.com",
				Name:         "Team Group",
				Description:  "A test group",
				MemberEmails: []string{"admin@example.com", "alice@example.com"},
			},
		},
		OrgUnits: []GoogleSeedOrgUnit{
			{
				Name:              "Engineering",
				Description:       "Engineering department",
				ParentOrgUnitPath: "/",
			},
		},
		Roles: []GoogleSeedRole{
			{
				RoleName:        "HelpdeskAdmin",
				RoleDescription: "Helpdesk administrator",
				Privileges:      []string{"USERS_READ"},
			},
		},
		RoleAssignments: []GoogleSeedRoleAssignment{
			{
				AssignedToEmail: "alice@example.com",
				RoleIDOrName:    "HelpdeskAdmin",
			},
		},
		Domains: []GoogleSeedDomain{
			{
				DomainName: "example.com",
				IsPrimary:  true,
			},
		},
		ChromeOSDevices: []GoogleSeedChromeOSDevice{
			{
				SerialNumber:  "CHROME-001",
				AnnotatedUser: "alice@example.com",
				OrgUnitPath:   "/Engineering",
				Notes:         "Test Chromebook",
			},
		},
		MobileDevices: []GoogleSeedMobileDevice{
			{
				SerialNumber: "MOB-001",
				Model:        "Pixel 8",
				OS:           "Android 14",
				Status:       "ACTIVE",
			},
		},
		GroupSettings: []GoogleSeedGroupSettings{
			{
				GroupEmail:           "team@example.com",
				WhoCanPostMessage:    "ALL_MEMBERS_CAN_POST",
				AllowExternalMembers: boolPtr(true),
				IsArchived:           boolPtr(true),
				PrimaryLanguage:      "en",
			},
		},
	}

	ctx := context.Background()
	err := SeedFromConfig(s, cfg)
	if err != nil {
		t.Fatalf("SeedFromConfig() failed: %v", err)
	}

	// Verify client was registered
	_, err = s.GetClient(ctx, "test-google-client-id")
	if err != nil {
		t.Errorf("Failed to get client: %v", err)
	}

	// Verify users were created
	users, _, err := s.ListUsers(ctx, model.ListOptions{})
	if err != nil {
		t.Errorf("Failed to list users: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	// Verify admin user is admin
	admin, err := s.GetUser(ctx, "admin@example.com")
	if err != nil {
		t.Errorf("Failed to get admin user: %v", err)
	}
	if !admin.IsAdmin {
		t.Errorf("Expected admin user to be admin, got isAdmin=%v", admin.IsAdmin)
	}

	// Verify group was created
	groups, _, err := s.ListGroups(ctx, model.ListOptions{})
	if err != nil {
		t.Errorf("Failed to list groups: %v", err)
	}
	if len(groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(groups))
	}

	// Verify group has members
	members, _, err := s.ListMembers(ctx, groups[0].ID, model.ListOptions{})
	if err != nil {
		t.Errorf("Failed to list members: %v", err)
	}
	if len(members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(members))
	}

	// Verify org unit was created
	orgUnits, err := s.ListOrgUnits(ctx, "my_customer")
	if err != nil {
		t.Errorf("Failed to list org units: %v", err)
	}
	foundEng := false
	for _, ou := range orgUnits {
		if ou.Name == "Engineering" {
			foundEng = true
			if ou.ParentOrgUnitPath != "/" {
				t.Errorf("Expected parent org unit path '/', got '%s'", ou.ParentOrgUnitPath)
			}
			if ou.OrgUnitPath != "/Engineering" {
				t.Errorf("Expected org unit path '/Engineering', got '%s'", ou.OrgUnitPath)
			}
		}
	}
	if !foundEng {
		t.Errorf("Expected Engineering org unit to be created")
	}

	// Verify role was created
	roles, err := s.ListRoles(ctx, "my_customer")
	if err != nil {
		t.Errorf("Failed to list roles: %v", err)
	}
	if len(roles) != 1 {
		t.Errorf("Expected 1 role, got %d", len(roles))
	}
	if roles[0].RoleName != "HelpdeskAdmin" {
		t.Errorf("Expected role name 'HelpdeskAdmin', got '%s'", roles[0].RoleName)
	}

	// Verify role assignment was created
	roleAssignments, err := s.ListRoleAssignments(ctx, "my_customer")
	if err != nil {
		t.Errorf("Failed to list role assignments: %v", err)
	}
	if len(roleAssignments) != 1 {
		t.Errorf("Expected 1 role assignment, got %d", len(roleAssignments))
	}

	// Verify domain was created
	domains, err := s.ListDomains(ctx, "my_customer")
	if err != nil {
		t.Errorf("Failed to list domains: %v", err)
	}
	if len(domains) != 1 {
		t.Errorf("Expected 1 domain, got %d", len(domains))
	}

	// Verify ChromeOS device was created
	chromeDevices, _, err := s.ListChromeOSDevices(ctx, "my_customer", model.ListOptions{})
	if err != nil {
		t.Errorf("Failed to list ChromeOS devices: %v", err)
	}
	if len(chromeDevices) != 1 {
		t.Errorf("Expected 1 ChromeOS device, got %d", len(chromeDevices))
	}
	if chromeDevices[0].SerialNumber != "CHROME-001" {
		t.Errorf("Expected serial number 'CHROME-001', got '%s'", chromeDevices[0].SerialNumber)
	}

	// Verify mobile device was created
	mobileDevices, _, err := s.ListMobileDevices(ctx, "my_customer", model.ListOptions{})
	if err != nil {
		t.Errorf("Failed to list mobile devices: %v", err)
	}
	if len(mobileDevices) != 1 {
		t.Errorf("Expected 1 mobile device, got %d", len(mobileDevices))
	}
	if mobileDevices[0].SerialNumber != "MOB-001" {
		t.Errorf("Expected serial number 'MOB-001', got '%s'", mobileDevices[0].SerialNumber)
	}

	// Verify group settings were applied
	gs, err := s.GetGroupSettings(ctx, "team@example.com")
	if err != nil {
		t.Errorf("Failed to get group settings: %v", err)
	}
	if gs.WhoCanPostMessage != "ALL_MEMBERS_CAN_POST" {
		t.Errorf("Expected whoCanPostMessage 'ALL_MEMBERS_CAN_POST', got '%s'", gs.WhoCanPostMessage)
	}
	if !gs.AllowExternalMembers {
		t.Errorf("Expected allowExternalMembers true, got %v", gs.AllowExternalMembers)
	}
	if !gs.IsArchived {
		t.Errorf("Expected isArchived true, got %v", gs.IsArchived)
	}
	if gs.PrimaryLanguage != "en" {
		t.Errorf("Expected primaryLanguage 'en', got '%s'", gs.PrimaryLanguage)
	}
}

func TestGoogleSeedFromConfigWithAliases(t *testing.T) {
	s := store.NewMemoryStore()

	cfg := &GoogleSeedConfig{
		Clients: []GoogleSeedClient{
			{
				ClientID:     "test-google-client-id",
				ClientSecret: "test-google-client-secret",
			},
		},
		Users: []GoogleSeedUser{
			{
				PrimaryEmail: "admin@example.com",
				GivenName:    "Admin",
				FamilyName:   "User",
				Password:     "TestPassword123!",
				Aliases:      []string{"admin.alias@example.com"},
			},
		},
		Groups: []GoogleSeedGroup{
			{
				Email:   "team@example.com",
				Name:    "Team Group",
				Aliases: []string{"team-alias@example.com"},
			},
		},
	}

	err := SeedFromConfig(s, cfg)
	if err != nil {
		t.Fatalf("SeedFromConfig() failed: %v", err)
	}

	// Verify user alias
	admin, err := s.GetUser(context.Background(), "admin@example.com")
	if err != nil {
		t.Fatalf("Failed to get admin user: %v", err)
	}
	aliases, err := s.ListUserAliases(context.Background(), admin.ID)
	if err != nil {
		t.Errorf("Failed to list user aliases: %v", err)
	}
	if len(aliases) != 1 || aliases[0] != "admin.alias@example.com" {
		t.Errorf("Expected user alias 'admin.alias@example.com', got %v", aliases)
	}

	// Verify group alias
	groups, _, err := s.ListGroups(context.Background(), model.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list groups: %v", err)
	}
	groupAliases, err := s.ListGroupAliases(context.Background(), groups[0].ID)
	if err != nil {
		t.Errorf("Failed to list group aliases: %v", err)
	}
	if len(groupAliases) != 1 || groupAliases[0] != "team-alias@example.com" {
		t.Errorf("Expected group alias 'team-alias@example.com', got %v", groupAliases)
	}
}

func TestBuildOrgUnitPath(t *testing.T) {
	tests := []struct {
		parentPath string
		name       string
		want       string
	}{
		{"/", "Engineering", "/Engineering"},
		{"", "Engineering", "/Engineering"},
		{"/Engineering", "Frontend", "/Engineering/Frontend"},
		{"/Engineering/Frontend", "WebTeam", "/Engineering/Frontend/WebTeam"},
	}
	for _, tt := range tests {
		got := buildOrgUnitPath(tt.parentPath, tt.name)
		if got != tt.want {
			t.Errorf("buildOrgUnitPath(%q, %q) = %q, want %q", tt.parentPath, tt.name, got, tt.want)
		}
	}
}

func boolPtr(b bool) *bool {
	return &b
}
