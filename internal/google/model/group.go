package model

type Group struct {
	Kind               string   `json:"kind"`
	ID                 string   `json:"id"`
	Etag               string   `json:"etag,omitempty"`
	Email              string   `json:"email"`
	Name               string   `json:"name,omitempty"`
	Description        string   `json:"description,omitempty"`
	AdminCreated       bool     `json:"adminCreated"`
	DirectMembersCount string   `json:"directMembersCount,omitempty"`
	Aliases            []string `json:"aliases,omitempty"`
	NonEditableAliases []string `json:"nonEditableAliases,omitempty"`
	CustomerID         string   `json:"customerId,omitempty"`
}

type GroupAlias struct {
	Kind     string `json:"kind,omitempty"`
	Etag     string `json:"etag,omitempty"`
	Id       string `json:"id,omitempty"`
	GroupKey string `json:"groupKey,omitempty"`
	Alias    string `json:"alias,omitempty"`
}