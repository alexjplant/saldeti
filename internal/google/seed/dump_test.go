package seed

import (
	"testing"

	"github.com/saldeti/saldeti/internal/google/store"
)

func TestGoogleDumpStoreRoundTrip(t *testing.T) {
	s := store.NewMemoryStore()

	cfg := &GoogleSeedConfig{
		Clients: []GoogleSeedClient{
			{ClientID: "client-two", ClientSecret: "secret-two"},
			{ClientID: "client-one", ClientSecret: "secret-one"},
		},
		Users: []GoogleSeedUser{
			{PrimaryEmail: "zoe@example.com", GivenName: "Zoe", FamilyName: "Z", IsAdmin: true, OrgUnitPath: "/", Aliases: []string{"zoe.alias@example.com"}},
			{PrimaryEmail: "alice@example.com", GivenName: "Alice", FamilyName: "A", OrgUnitPath: "/Engineering"},
			{PrimaryEmail: "bob@example.com", GivenName: "Bob", FamilyName: "B", Suspended: true, OrgUnitPath: "/Marketing"},
		},
		Groups: []GoogleSeedGroup{
			{Email: "team@example.com", Name: "Team", Description: "The team", MemberEmails: []string{"zoe@example.com", "alice@example.com"}, Aliases: []string{"team-alias@example.com"}},
		},
		OrgUnits: []GoogleSeedOrgUnit{
			{Name: "Marketing", ParentOrgUnitPath: "/"},
			{Name: "Engineering", ParentOrgUnitPath: "/"},
		},
		Roles: []GoogleSeedRole{
			{RoleName: "HelpdeskAdmin", RoleDescription: "Helpdesk", Privileges: []string{"USERS_READ", "GROUPS_READ"}},
		},
		RoleAssignments: []GoogleSeedRoleAssignment{
			{AssignedToEmail: "alice@example.com", RoleIDOrName: "HelpdeskAdmin"},
		},
		Domains: []GoogleSeedDomain{
			{DomainName: "example.com", IsPrimary: true},
			{DomainName: "another.example.com", IsPrimary: false},
		},
		ChromeOSDevices: []GoogleSeedChromeOSDevice{
			{SerialNumber: "CHROME-002", AnnotatedUser: "alice@example.com", OrgUnitPath: "/Engineering", Notes: "second"},
			{SerialNumber: "CHROME-001", AnnotatedUser: "zoe@example.com", OrgUnitPath: "/", Notes: "first"},
		},
		MobileDevices: []GoogleSeedMobileDevice{
			{SerialNumber: "MOB-002", Model: "Pixel 9", OS: "Android 15", Status: "ACTIVE"},
			{SerialNumber: "MOB-001", Model: "Pixel 8", OS: "Android 14", Status: "ACTIVE"},
		},
	}

	if err := SeedFromConfig(s, cfg); err != nil {
		t.Fatalf("SeedFromConfig failed: %v", err)
	}

	dumped, err := DumpStore(s)
	if err != nil {
		t.Fatalf("DumpStore failed: %v", err)
	}

	// Clients sorted by ClientID.
	if len(dumped.Clients) != 2 {
		t.Fatalf("Expected 2 clients, got %d", len(dumped.Clients))
	}
	if dumped.Clients[0].ClientID != "client-one" || dumped.Clients[1].ClientID != "client-two" {
		t.Errorf("Clients not sorted by ClientID: %v %v", dumped.Clients[0].ClientID, dumped.Clients[1].ClientID)
	}

	// Users sorted by PrimaryEmail; password not stored.
	if len(dumped.Users) != 3 {
		t.Fatalf("Expected 3 users, got %d", len(dumped.Users))
	}
	if dumped.Users[0].PrimaryEmail != "alice@example.com" {
		t.Errorf("Users not sorted; first is %s", dumped.Users[0].PrimaryEmail)
	}
	if dumped.Users[0].Password != "" {
		t.Errorf("Password should not be stored/dumped, got %q", dumped.Users[0].Password)
	}
	// zoe is last and is admin with an alias.
	zoe := dumped.Users[2]
	if zoe.PrimaryEmail != "zoe@example.com" {
		t.Errorf("Expected zoe last, got %s", zoe.PrimaryEmail)
	}
	if !zoe.IsAdmin {
		t.Errorf("Expected zoe to be admin")
	}
	if len(zoe.Aliases) != 1 || zoe.Aliases[0] != "zoe.alias@example.com" {
		t.Errorf("Expected zoe alias, got %v", zoe.Aliases)
	}
	// bob suspended.
	if !dumped.Users[1].Suspended {
		t.Errorf("Expected bob suspended")
	}

	// Group with sorted member_emails and aliases.
	if len(dumped.Groups) != 1 {
		t.Fatalf("Expected 1 group, got %d", len(dumped.Groups))
	}
	g := dumped.Groups[0]
	if g.Email != "team@example.com" {
		t.Errorf("Expected team@example.com, got %s", g.Email)
	}
	if len(g.MemberEmails) != 2 || g.MemberEmails[0] != "alice@example.com" || g.MemberEmails[1] != "zoe@example.com" {
		t.Errorf("Expected sorted member emails [alice, zoe], got %v", g.MemberEmails)
	}
	if len(g.Aliases) != 1 || g.Aliases[0] != "team-alias@example.com" {
		t.Errorf("Expected group alias, got %v", g.Aliases)
	}

	// Org units sorted by OrgUnitPath (Engineering before Marketing).
	if len(dumped.OrgUnits) != 2 {
		t.Fatalf("Expected 2 org units, got %d", len(dumped.OrgUnits))
	}
	if dumped.OrgUnits[0].Name != "Engineering" {
		t.Errorf("Expected Engineering first, got %s", dumped.OrgUnits[0].Name)
	}

	// Role with sorted privileges.
	if len(dumped.Roles) != 1 {
		t.Fatalf("Expected 1 role, got %d", len(dumped.Roles))
	}
	if dumped.Roles[0].RoleName != "HelpdeskAdmin" {
		t.Errorf("Expected HelpdeskAdmin, got %s", dumped.Roles[0].RoleName)
	}
	if len(dumped.Roles[0].Privileges) != 2 {
		t.Errorf("Expected 2 privileges, got %v", dumped.Roles[0].Privileges)
	}

	// Role assignment: AssignedTo resolved ID->email, RoleId resolved to role name.
	if len(dumped.RoleAssignments) != 1 {
		t.Fatalf("Expected 1 role assignment, got %d", len(dumped.RoleAssignments))
	}
	ra := dumped.RoleAssignments[0]
	if ra.AssignedToEmail != "alice@example.com" {
		t.Errorf("Expected assigned_to_email alice@example.com, got %s", ra.AssignedToEmail)
	}
	if ra.RoleIDOrName != "HelpdeskAdmin" {
		t.Errorf("Expected role_id_or_name HelpdeskAdmin (resolved from UUID), got %s", ra.RoleIDOrName)
	}

	// Domains sorted by DomainName.
	if len(dumped.Domains) != 2 {
		t.Fatalf("Expected 2 domains, got %d", len(dumped.Domains))
	}
	if dumped.Domains[0].DomainName != "another.example.com" {
		t.Errorf("Expected another.example.com first, got %s", dumped.Domains[0].DomainName)
	}

	// ChromeOS devices sorted by SerialNumber.
	if len(dumped.ChromeOSDevices) != 2 {
		t.Fatalf("Expected 2 chromeos devices, got %d", len(dumped.ChromeOSDevices))
	}
	if dumped.ChromeOSDevices[0].SerialNumber != "CHROME-001" {
		t.Errorf("Expected CHROME-001 first, got %s", dumped.ChromeOSDevices[0].SerialNumber)
	}

	// Mobile devices sorted by SerialNumber; Os field mapped to OS.
	if len(dumped.MobileDevices) != 2 {
		t.Fatalf("Expected 2 mobile devices, got %d", len(dumped.MobileDevices))
	}
	if dumped.MobileDevices[0].SerialNumber != "MOB-001" {
		t.Errorf("Expected MOB-001 first, got %s", dumped.MobileDevices[0].SerialNumber)
	}
	if dumped.MobileDevices[0].OS != "Android 14" {
		t.Errorf("Expected OS 'Android 14', got %q", dumped.MobileDevices[0].OS)
	}
}

func TestGoogleDumpStoreEmpty(t *testing.T) {
	s := store.NewMemoryStore()
	dumped, err := DumpStore(s)
	if err != nil {
		t.Fatalf("DumpStore on empty store failed: %v", err)
	}
	if dumped == nil {
		t.Fatal("DumpStore returned nil config for empty store")
	}
	if len(dumped.Clients) != 0 {
		t.Errorf("Expected 0 clients on empty store, got %d", len(dumped.Clients))
	}
}