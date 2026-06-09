package model

type ListOptions struct {
	MaxResults    int    `json:"maxResults,omitempty"`
	PageToken     string `json:"pageToken,omitempty"`
	Query         string `json:"query,omitempty"`
	OrderBy       string `json:"orderBy,omitempty"`
	SortOrder     string `json:"sortOrder,omitempty"`
	Customer      string `json:"customer,omitempty"`
	Domain        string `json:"domain,omitempty"`
	Projection    string `json:"projection,omitempty"`
	ViewType      string `json:"viewType,omitempty"`
}

type PagedResponse struct {
	Kind          string `json:"kind"`
	Etag          string `json:"etag,omitempty"`
	NextPageToken string `json:"nextPageToken,omitempty"`
}

type GoogleError struct {
	Error GoogleErrorDetail `json:"error"`
}

type GoogleErrorDetail struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Status  string            `json:"status,omitempty"`
	Errors  []GoogleErrorItem `json:"errors,omitempty"`
}

type GoogleErrorItem struct {
	Domain  string `json:"domain"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}