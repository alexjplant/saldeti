package seed

import (
	"github.com/saldeti/saldeti/internal/model"
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
				Email:            "admin@saldeti.local",
				DisplayName:      "Admin User",
				Password:         "Simulator123!",
				GivenName:        "Admin",
				Surname:          "User",
				JobTitle:         "Administrator",
				Department:       "IT",
				Enabled:          &accountEnabled,
				IsGuest:          false,
				AssignedLicenses: []model.SeedLicense{{SkuPartNumber: "SPE_E5"}},
			},
			// Alice Smith (index 1)
			{
				Email:            "alice.smith@saldeti.local",
				DisplayName:      "Alice Smith",
				Password:         "Simulator123!",
				GivenName:        "Alice",
				Surname:          "Smith",
				JobTitle:         "Software Engineer",
				Department:       "Engineering",
				Enabled:          &accountEnabled,
				IsGuest:          false,
				ManagerUPN:       "eve.wilson@saldeti.local",
				AssignedLicenses: []model.SeedLicense{{SkuPartNumber: "SPE_E3", DisabledPlans: []string{"MCOSTANDARD"}}},
			},
			// Bob Jones (index 2)
			{
				Email:            "bob.jones@saldeti.local",
				DisplayName:      "Bob Jones",
				Password:         "Simulator123!",
				GivenName:        "Bob",
				Surname:          "Jones",
				JobTitle:         "Senior Engineer",
				Department:       "Engineering",
				Enabled:          &accountEnabled,
				IsGuest:          false,
				ManagerUPN:       "eve.wilson@saldeti.local",
				AssignedLicenses: []model.SeedLicense{{SkuPartNumber: "SPE_E3"}},
			},
			// Charlie Brown (index 3)
			{
				Email:            "charlie.brown@saldeti.local",
				DisplayName:      "Charlie Brown",
				Password:         "Simulator123!",
				GivenName:        "Charlie",
				Surname:          "Brown",
				JobTitle:         "Marketing Manager",
				Department:       "Marketing",
				Enabled:          &accountEnabled,
				IsGuest:          false,
				AssignedLicenses: []model.SeedLicense{{SkuPartNumber: "SPE_E3"}},
			},
			// Diana Prince (index 4)
			{
				Email:            "diana.prince@saldeti.local",
				DisplayName:      "Diana Prince",
				Password:         "Simulator123!",
				GivenName:        "Diana",
				Surname:          "Prince",
				JobTitle:         "HR Director",
				Department:       "HR",
				Enabled:          &accountEnabled,
				IsGuest:          false,
				ManagerUPN:       "admin@saldeti.local",
				AssignedLicenses: []model.SeedLicense{{SkuPartNumber: "SPE_E3"}},
			},
			// Eve Wilson (index 5)
			{
				Email:            "eve.wilson@saldeti.local",
				DisplayName:      "Eve Wilson",
				Password:         "Simulator123!",
				GivenName:        "Eve",
				Surname:          "Wilson",
				JobTitle:         "Tech Lead",
				Department:       "Engineering",
				Enabled:          &accountEnabled,
				IsGuest:          false,
				ManagerUPN:       "frank.miller@saldeti.local",
				AssignedLicenses: []model.SeedLicense{{SkuPartNumber: "SPE_E5"}, {SkuPartNumber: "EMS"}},
			},
			// Frank Miller (index 6)
			{
				Email:            "frank.miller@saldeti.local",
				DisplayName:      "Frank Miller",
				Password:         "Simulator123!",
				GivenName:        "Frank",
				Surname:          "Miller",
				JobTitle:         "CFO",
				Department:       "Finance",
				Enabled:          &accountEnabled,
				IsGuest:          false,
				ManagerUPN:       "admin@saldeti.local",
				AssignedLicenses: []model.SeedLicense{{SkuPartNumber: "SPE_E5"}},
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
				Email:            "henry.taylor@saldeti.local",
				DisplayName:      "Henry Taylor",
				Password:         "Simulator123!",
				GivenName:        "Henry",
				Surname:          "Taylor",
				JobTitle:         "Account Executive",
				Department:       "Sales",
				Enabled:          &accountEnabled,
				IsGuest:          false,
				AssignedLicenses: []model.SeedLicense{{SkuPartNumber: "SPE_E3"}},
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
				Email:            "julia.roberts@saldeti.local",
				DisplayName:      "Julia Roberts",
				Password:         "Simulator123!",
				GivenName:        "Julia",
				Surname:          "Roberts",
				JobTitle:         "Content Writer",
				Department:       "Marketing",
				Enabled:          &accountEnabled,
				IsGuest:          false,
				AssignedLicenses: []model.SeedLicense{{SkuPartNumber: "SPE_E3"}},
			},
		},
		Groups: []SeedGroup{
			{
				DisplayName:  "Engineering Team",
				Description:  "Engineering department",
				MailNickname: "engineeringteam",
				Visibility:   "Public",
				MemberUPNs:   []string{"alice.smith@saldeti.local", "bob.jones@saldeti.local", "eve.wilson@saldeti.local", "grace.lee@saldeti.local"},
			},
			{
				DisplayName:  "Marketing Team",
				Description:  "Marketing department",
				MailNickname: "marketingteam",
				Visibility:   "Public",
				MemberUPNs:   []string{"charlie.brown@saldeti.local", "julia.roberts@saldeti.local"},
			},
			{
				DisplayName:       "All Staff",
				Description:       "All employees",
				MailNickname:      "allstaff",
				Visibility:        "Public",
				MemberUPNs:        []string{"alice.smith@saldeti.local", "bob.jones@saldeti.local", "charlie.brown@saldeti.local", "diana.prince@saldeti.local", "eve.wilson@saldeti.local", "frank.miller@saldeti.local", "henry.taylor@saldeti.local", "julia.roberts@saldeti.local"},
				MemberGroupNames: []string{"Engineering Team", "Marketing Team"},
			},
			{
				DisplayName:  "Leadership",
				Description:  "Leadership team",
				MailNickname: "leadership",
				Visibility:   "Private",
				MemberUPNs:   []string{"diana.prince@saldeti.local", "frank.miller@saldeti.local"},
			},
			{
				DisplayName:  "Project Alpha",
				Description:  "Project Alpha team",
				MailNickname: "projectalpha",
				Visibility:   "Public",
				GroupTypes:   []string{"Unified"},
				MemberUPNs:   []string{"alice.smith@saldeti.local", "charlie.brown@saldeti.local", "eve.wilson@saldeti.local"},
			},
		},
	}

	return SeedFromConfig(s, cfg)
}

func intPtr(i int) *int {
	return &i
}
