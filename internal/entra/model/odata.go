package model

type ExpandOption struct {
	Property string   // e.g. "manager"
	Select   []string // e.g. ["userPrincipalName"] from ($select=userPrincipalName,displayName); nil if no nested $select
}

type ListOptions struct {
	Filter        string
	Select        []string
	Top           int
	OrderBy       string
	Count         bool
	Search        string
	Skip          int
	ExpandOptions []ExpandOption
}

type ListResponse struct {
	Context  string `json:"@odata.context"`
	Count    *int   `json:"@odata.count,omitempty"`
	NextLink string `json:"@odata.nextLink,omitempty"`
	Value    any    `json:"value"`
}

type SingleResponse struct {
	Context string `json:"@odata.context"`
	Value   any    `json:"value"`
}
