package model

type ListOptions struct {
    Filter   string
    Select   []string
    Top      int
    OrderBy  string
    Count    bool
    Search   string
    Skip     int
}

type ListResponse struct {
    Context     string        `json:"@odata.context"`
    Count       *int          `json:"@odata.count,omitempty"`
    NextLink    string        `json:"@odata.nextLink,omitempty"`
    Value       interface{}   `json:"value"`
}

type SingleResponse struct {
	Context string      `json:"@odata.context"`
	Value   interface{} `json:"value"`
}