package model

type DirectoryObject struct {
	ODataType      string                 `json:"@odata.type,omitempty"`
	ID             string                 `json:"id"`
	DisplayName    string                 `json:"displayName,omitempty"`
	AdditionalData map[string]interface{} `json:"-"`
}
