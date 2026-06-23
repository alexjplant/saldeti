package model

type Domain struct {
	Kind          string        `json:"kind"`
	Etag          string        `json:"etag,omitempty"`
	DomainName    string        `json:"domainName"`
	IsPrimary     bool          `json:"isPrimary"`
	Verified      bool          `json:"verified"`
	CreationTime  string        `json:"creationTime,omitempty"`
	DomainAliases []DomainAlias `json:"domainAliases,omitempty"`
}

type DomainAlias struct {
	Kind             string `json:"kind"`
	Etag             string `json:"etag,omitempty"`
	DomainAliasName  string `json:"domainAliasName"`
	ParentDomainName string `json:"parentDomainName,omitempty"`
	Verified         bool   `json:"verified"`
}
