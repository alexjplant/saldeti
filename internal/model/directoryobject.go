package model

type DirectoryObject struct {
	ODataType      string                 `json:"@odata.type,omitempty"`
	ID             string                 `json:"id,omitempty"`
	DisplayName    string                 `json:"displayName,omitempty"`
	AdditionalData map[string]interface{} `json:"-"`
}