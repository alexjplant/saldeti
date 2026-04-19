package model

type ServicePrincipal struct {
	ID          string `json:"id"`
	AppID       string `json:"appId"`
	DisplayName string `json:"displayName,omitempty"`
	ODataType   string `json:"@odata.type,omitempty"`
}
