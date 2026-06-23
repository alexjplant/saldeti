package model

type Role struct {
	Etag             string          `json:"etag,omitempty"`
	RoleId           string          `json:"roleId,omitempty"`
	RoleName         string          `json:"roleName"`
	RoleDescription  string          `json:"roleDescription,omitempty"`
	RolePrivileges   []RolePrivilege `json:"rolePrivileges,omitempty"`
	IsSystemRole     bool            `json:"isSystemRole"`
	IsSuperAdminRole bool            `json:"isSuperAdminRole"`
	Kind             string          `json:"kind,omitempty"`
}

type RolePrivilege struct {
	PrivilegeName string `json:"privilegeName"`
	ServiceId     string `json:"serviceId"`
}

type RoleAssignment struct {
	Kind             string `json:"kind"`
	Etag             string `json:"etag,omitempty"`
	RoleAssignmentId string `json:"roleAssignmentId,omitempty"`
	RoleId           string `json:"roleId"`
	AssignedTo       string `json:"assignedTo"`
	AssigneeType     string `json:"assigneeType,omitempty"`
	ScopeType        string `json:"scopeType,omitempty"`
	OrgUnitId        string `json:"orgUnitId,omitempty"`
}

type Privilege struct {
	Kind            string      `json:"kind"`
	Etag            string      `json:"etag,omitempty"`
	ServiceName     string      `json:"serviceName"`
	PrivilegeName   string      `json:"privilegeName"`
	ServiceId       string      `json:"serviceId"`
	IsOuScopable    bool        `json:"isOuScopable"`
	ChildPrivileges []Privilege `json:"childPrivileges,omitempty"`
}
