package model

type CloudIdentityGroup struct {
	Name        string            `json:"name"`
	GroupKey    *EntityKey        `json:"groupKey,omitempty"`
	Parent      string            `json:"parent,omitempty"`
	DisplayName string            `json:"displayName,omitempty"`
	Description string            `json:"description,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	CreateTime  string            `json:"createTime,omitempty"`
	UpdateTime  string            `json:"updateTime,omitempty"`
}

type EntityKey struct {
	ID        string `json:"id"`
	Namespace string `json:"namespace,omitempty"`
}

type Membership struct {
	Name       string           `json:"name"`
	MemberKey  *EntityKey       `json:"memberKey,omitempty"`
	Roles      []MembershipRole `json:"roles,omitempty"`
	CreateTime string           `json:"createTime,omitempty"`
	UpdateTime string           `json:"updateTime,omitempty"`
}

type MembershipRole struct {
	Name        string       `json:"name"`
	ExpiryDetail *ExpiryDetail `json:"expiryDetail,omitempty"`
}

type ExpiryDetail struct {
	ExpireTime string `json:"expireTime,omitempty"`
}

type ModifyMembershipRolesRequest struct {
	AddRoles    []MembershipRole `json:"addRoles,omitempty"`
	RemoveRoles []string         `json:"removeRoles,omitempty"`
}

type SecuritySettings struct {
	Name                     string `json:"name,omitempty"`
	WhoCanJoin               string `json:"whoCanJoin,omitempty"`
	WhoCanViewMembership     string `json:"whoCanViewMembership,omitempty"`
	WhoCanDiscoverGroup      string `json:"whoCanDiscoverGroup,omitempty"`
	WhoCanModerateMembers    string `json:"whoCanModerateMembers,omitempty"`
	WhoCanManage             string `json:"whoCanManage,omitempty"`
}

type MembershipGraph struct {
	AdjacencyList []MembershipAdjacency `json:"adjacencyList,omitempty"`
}

type MembershipAdjacency struct {
	Edge    *MembershipEdge `json:"edge,omitempty"`
	Members []Membership    `json:"members,omitempty"`
}

type MembershipEdge struct {
	SourceMember string `json:"sourceMember,omitempty"`
	TargetMember string `json:"targetMember,omitempty"`
	Role         string `json:"role,omitempty"`
}