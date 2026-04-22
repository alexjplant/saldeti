package seed

import "github.com/saldeti/saldeti/internal/model"

// SeedConfig is the top-level structure for a seed JSON file.
type SeedConfig struct {
	Clients     []SeedClient      `json:"clients"`
	Users       []SeedUser        `json:"users,omitempty"`
	Groups      []SeedGroup       `json:"groups,omitempty"`
	Memberships []SeedMembership  `json:"memberships,omitempty"`
	Ownerships  []SeedOwnership   `json:"ownerships,omitempty"`
	Managers    []SeedManager     `json:"managers,omitempty"`
}

type SeedClient struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	TenantID     string `json:"tenant_id"`
}

type SeedUser struct {
	ID          string `json:"id,omitempty"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
	GivenName   string `json:"given_name,omitempty"`
	Surname     string `json:"surname,omitempty"`
	JobTitle    string `json:"job_title,omitempty"`
	Department        string                  `json:"department,omitempty"`
	Enabled           *bool                   `json:"enabled,omitempty"` // defaults to true if nil
	IsGuest           bool                    `json:"is_guest,omitempty"`
	ManagerUPN        string                  `json:"manager_upn,omitempty"`
	AssignedLicenses  []model.SeedLicense     `json:"assigned_licenses,omitempty"`
}

type SeedGroup struct {
	ID               string   `json:"id,omitempty"`
	DisplayName      string   `json:"display_name"`
	Description      string   `json:"description,omitempty"`
	MailNickname     string   `json:"mail_nickname,omitempty"`
	Visibility       string   `json:"visibility,omitempty"` // "Public" (default) or "Private"
	GroupTypes       []string `json:"group_types,omitempty"`
	MemberUPNs       []string `json:"member_upns,omitempty"`
	MemberGroupNames []string `json:"member_group_names,omitempty"`
	OwnerUPNs        []string `json:"owner_upns,omitempty"`
}

// SeedMembership adds an entity to a group.
// Use UserIndex for user membership, MemberGroupIndex for nested group membership.
type SeedMembership struct {
	UserIndex        *int `json:"user_index,omitempty"`
	GroupIndex       *int `json:"group_index,omitempty"`
	MemberGroupIndex *int `json:"member_group_index,omitempty"`
}

// SeedManager sets a manager for a user.
type SeedManager struct {
	UserIndex    int `json:"user_index"`
	ManagerIndex int `json:"manager_index"`
}

// SeedOwnership sets an owner for a group.
type SeedOwnership struct {
	UserIndex  int `json:"user_index"`
	GroupIndex int `json:"group_index"`
}
