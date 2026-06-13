package seed

// GoogleSeedConfig is the top-level structure for a Google Workspace seed JSON file.
//
//go:generate go run ../../cmd/genschema/main.go
type GoogleSeedConfig struct {
	Clients         []GoogleSeedClient         `json:"clients"`
	Users           []GoogleSeedUser           `json:"users,omitempty"`
	Groups          []GoogleSeedGroup          `json:"groups,omitempty"`
	OrgUnits        []GoogleSeedOrgUnit        `json:"org_units,omitempty"`
	Roles           []GoogleSeedRole           `json:"roles,omitempty"`
	RoleAssignments []GoogleSeedRoleAssignment `json:"role_assignments,omitempty"`
	Domains         []GoogleSeedDomain         `json:"domains,omitempty"`
	ChromeOSDevices []GoogleSeedChromeOSDevice `json:"chromeos_devices,omitempty"`
	MobileDevices   []GoogleSeedMobileDevice   `json:"mobile_devices,omitempty"`
}

type GoogleSeedClient struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type GoogleSeedUser struct {
	PrimaryEmail string   `json:"primary_email"`
	GivenName    string   `json:"given_name,omitempty"`
	FamilyName   string   `json:"family_name,omitempty"`
	Password     string   `json:"password,omitempty"`
	Suspended    bool     `json:"suspended,omitempty"`
	IsAdmin      bool     `json:"is_admin,omitempty"`
	OrgUnitPath  string   `json:"org_unit_path,omitempty"`
	Aliases      []string `json:"aliases,omitempty"`
}

type GoogleSeedGroup struct {
	Email        string   `json:"email"`
	Name         string   `json:"name,omitempty"`
	Description  string   `json:"description,omitempty"`
	MemberEmails []string `json:"member_emails,omitempty"`
	Aliases      []string `json:"aliases,omitempty"`
}

type GoogleSeedOrgUnit struct {
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
	ParentOrgUnitPath string `json:"parent_org_unit_path,omitempty"`
	BlockInheritance  bool   `json:"block_inheritance,omitempty"`
}

type GoogleSeedRole struct {
	RoleName        string   `json:"role_name"`
	RoleDescription string   `json:"role_description,omitempty"`
	Privileges      []string `json:"privileges,omitempty"`
}

type GoogleSeedRoleAssignment struct {
	AssignedToEmail string `json:"assigned_to_email"`
	RoleIDOrName   string `json:"role_id_or_name"`
}

type GoogleSeedDomain struct {
	DomainName string `json:"domain_name"`
	IsPrimary  bool   `json:"is_primary,omitempty"`
}

type GoogleSeedChromeOSDevice struct {
	SerialNumber  string `json:"serial_number"`
	AnnotatedUser string `json:"annotated_user,omitempty"`
	OrgUnitPath   string `json:"org_unit_path,omitempty"`
	Notes         string `json:"notes,omitempty"`
}

type GoogleSeedMobileDevice struct {
	SerialNumber string `json:"serial_number"`
	Model        string `json:"model,omitempty"`
	OS           string `json:"os,omitempty"`
	Status       string `json:"status,omitempty"`
}
