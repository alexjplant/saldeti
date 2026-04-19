package seed

import (
	"github.com/saldeti/saldeti/internal/store"
)

// Seed seeds the store with default data for backward compatibility.
// It constructs a SeedConfig in Go code and delegates to SeedFromConfig.
func Seed(s store.Store) error {
	accountEnabled := true
	accountDisabled := false

	cfg := &SeedConfig{
		Clients: []SeedClient{
			{
				ClientID:     "sim-client-id",
				ClientSecret: "sim-client-secret",
				TenantID:     "sim-tenant-id",
			},
		},
		Users: []SeedUser{
			// Admin user (index 0)
			{
				Email:       "admin@saldeti.local",
				DisplayName: "Admin User",
				Password:    "Simulator123!",
				GivenName:   "Admin",
				Surname:     "User",
				JobTitle:    "Administrator",
				Department:  "IT",
				Enabled:     &accountEnabled,
				IsGuest:     false,
			},
			// Alice Smith (index 1)
			{
				Email:       "alice.smith@saldeti.local",
				DisplayName: "Alice Smith",
				Password:    "Simulator123!",
				GivenName:   "Alice",
				Surname:     "Smith",
				JobTitle:    "Software Engineer",
				Department:  "Engineering",
				Enabled:     &accountEnabled,
				IsGuest:     false,
			},
			// Bob Jones (index 2)
			{
				Email:       "bob.jones@saldeti.local",
				DisplayName: "Bob Jones",
				Password:    "Simulator123!",
				GivenName:   "Bob",
				Surname:     "Jones",
				JobTitle:    "Senior Engineer",
				Department:  "Engineering",
				Enabled:     &accountEnabled,
				IsGuest:     false,
			},
			// Charlie Brown (index 3)
			{
				Email:       "charlie.brown@saldeti.local",
				DisplayName: "Charlie Brown",
				Password:    "Simulator123!",
				GivenName:   "Charlie",
				Surname:     "Brown",
				JobTitle:    "Marketing Manager",
				Department:  "Marketing",
				Enabled:     &accountEnabled,
				IsGuest:     false,
			},
			// Diana Prince (index 4)
			{
				Email:       "diana.prince@saldeti.local",
				DisplayName: "Diana Prince",
				Password:    "Simulator123!",
				GivenName:   "Diana",
				Surname:     "Prince",
				JobTitle:    "HR Director",
				Department:  "HR",
				Enabled:     &accountEnabled,
				IsGuest:     false,
			},
			// Eve Wilson (index 5)
			{
				Email:       "eve.wilson@saldeti.local",
				DisplayName: "Eve Wilson",
				Password:    "Simulator123!",
				GivenName:   "Eve",
				Surname:     "Wilson",
				JobTitle:    "Tech Lead",
				Department:  "Engineering",
				Enabled:     &accountEnabled,
				IsGuest:     false,
			},
			// Frank Miller (index 6)
			{
				Email:       "frank.miller@saldeti.local",
				DisplayName: "Frank Miller",
				Password:    "Simulator123!",
				GivenName:   "Frank",
				Surname:     "Miller",
				JobTitle:    "CFO",
				Department:  "Finance",
				Enabled:     &accountEnabled,
				IsGuest:     false,
			},
			// Grace Lee (index 7, disabled)
			{
				Email:       "grace.lee@saldeti.local",
				DisplayName: "Grace Lee",
				Password:    "Simulator123!",
				GivenName:   "Grace",
				Surname:     "Lee",
				JobTitle:    "Intern",
				Department:  "Engineering",
				Enabled:     &accountDisabled,
				IsGuest:     false,
			},
			// Henry Taylor (index 8)
			{
				Email:       "henry.taylor@saldeti.local",
				DisplayName: "Henry Taylor",
				Password:    "Simulator123!",
				GivenName:   "Henry",
				Surname:     "Taylor",
				JobTitle:    "Account Executive",
				Department:  "Sales",
				Enabled:     &accountEnabled,
				IsGuest:     false,
			},
			// Ivan Guest (index 9, guest)
			{
				Email:       "ivan.guest@external.com",
				DisplayName: "Ivan Guest",
				Password:    "Simulator123!",
				GivenName:   "Ivan",
				Surname:     "Guest",
				Enabled:     &accountEnabled,
				IsGuest:     true,
			},
			// Julia Roberts (index 10)
			{
				Email:       "julia.roberts@saldeti.local",
				DisplayName: "Julia Roberts",
				Password:    "Simulator123!",
				GivenName:   "Julia",
				Surname:     "Roberts",
				JobTitle:    "Content Writer",
				Department:  "Marketing",
				Enabled:     &accountEnabled,
				IsGuest:     false,
			},
		},
		Groups: []SeedGroup{
			{
				DisplayName:  "Engineering Team",
				Description:  "Engineering department",
				MailNickname: "engineeringteam",
				Visibility:   "Public",
			},
			{
				DisplayName:  "Marketing Team",
				Description:  "Marketing department",
				MailNickname: "marketingteam",
				Visibility:   "Public",
			},
			{
				DisplayName:  "All Staff",
				Description:  "All employees",
				MailNickname: "allstaff",
				Visibility:   "Public",
			},
			{
				DisplayName:  "Leadership",
				Description:  "Leadership team",
				MailNickname: "leadership",
				Visibility:   "Private",
			},
			{
				DisplayName:  "Project Alpha",
				Description:  "Project Alpha team",
				MailNickname: "projectalpha",
				Visibility:   "Public",
				GroupTypes:   []string{"Unified"},
			},
		},
		Memberships: []SeedMembership{
			// Engineering Team: alice (1), bob (2), eve (5), grace (7)
			{UserIndex: intPtr(1), GroupIndex: intPtr(0)},
			{UserIndex: intPtr(2), GroupIndex: intPtr(0)},
			{UserIndex: intPtr(5), GroupIndex: intPtr(0)},
			{UserIndex: intPtr(7), GroupIndex: intPtr(0)},
			// Marketing Team: charlie (3), julia (10)
			{UserIndex: intPtr(3), GroupIndex: intPtr(1)},
			{UserIndex: intPtr(10), GroupIndex: intPtr(1)},
			// All Staff: alice (1), bob (2), charlie (3), diana (4), eve (5), frank (6), henry (8), julia (10)
			{UserIndex: intPtr(1), GroupIndex: intPtr(2)},
			{UserIndex: intPtr(2), GroupIndex: intPtr(2)},
			{UserIndex: intPtr(3), GroupIndex: intPtr(2)},
			{UserIndex: intPtr(4), GroupIndex: intPtr(2)},
			{UserIndex: intPtr(5), GroupIndex: intPtr(2)},
			{UserIndex: intPtr(6), GroupIndex: intPtr(2)},
			{UserIndex: intPtr(8), GroupIndex: intPtr(2)},
			{UserIndex: intPtr(10), GroupIndex: intPtr(2)},
			// All Staff nested: Engineering Team (0), Marketing Team (1)
			{MemberGroupIndex: intPtr(0), GroupIndex: intPtr(2)},
			{MemberGroupIndex: intPtr(1), GroupIndex: intPtr(2)},
			// Leadership: diana (4), frank (6)
			{UserIndex: intPtr(4), GroupIndex: intPtr(3)},
			{UserIndex: intPtr(6), GroupIndex: intPtr(3)},
			// Project Alpha: alice (1), charlie (3), eve (5)
			{UserIndex: intPtr(1), GroupIndex: intPtr(4)},
			{UserIndex: intPtr(3), GroupIndex: intPtr(4)},
			{UserIndex: intPtr(5), GroupIndex: intPtr(4)},
		},
		Managers: []SeedManager{
			// frank (6) → manager: admin (0)
			{UserIndex: 6, ManagerIndex: 0},
			// diana (4) → manager: admin (0)
			{UserIndex: 4, ManagerIndex: 0},
			// eve (5) → manager: frank (6)
			{UserIndex: 5, ManagerIndex: 6},
			// alice (1) → manager: eve (5)
			{UserIndex: 1, ManagerIndex: 5},
			// bob (2) → manager: eve (5)
			{UserIndex: 2, ManagerIndex: 5},
		},
	}

	return SeedFromConfig(s, cfg)
}

func intPtr(i int) *int {
	return &i
}
