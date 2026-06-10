package model

type OrgUnit struct {
	Kind              string `json:"kind"`
	Etag              string `json:"etag,omitempty"`
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
	OrgUnitPath       string `json:"orgUnitPath"`
	OrgUnitId         string `json:"orgUnitId"`
	ParentOrgUnitPath string `json:"parentOrgUnitPath,omitempty"`
	ParentOrgUnitId   string `json:"parentOrgUnitId,omitempty"`
	BlockInheritance  bool   `json:"blockInheritance,omitempty"`
}