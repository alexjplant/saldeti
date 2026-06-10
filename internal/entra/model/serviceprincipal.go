package model

import "time"

type ServicePrincipal struct {
	ODataType                          string               `json:"@odata.type,omitempty"`
	ID                                 string               `json:"id,omitempty"`
	AppID                              string               `json:"appId,omitempty"`
	DisplayName                        string               `json:"displayName,omitempty"`
	Description                        string               `json:"description,omitempty"`
	AppOwnerOrganizationID             string               `json:"appOwnerOrganizationId,omitempty"`
	ServicePrincipalNames              []string             `json:"servicePrincipalNames,omitempty"`
	ServicePrincipalType               string               `json:"servicePrincipalType,omitempty"`
	AccountEnabled                     *bool                `json:"accountEnabled,omitempty"`
	AppRoles                           []AppRole            `json:"appRoles,omitempty"`
	OAuth2PermissionScopes             []PermissionScope    `json:"oauth2PermissionScopes,omitempty"`
	PasswordCredentials                []PasswordCredential `json:"passwordCredentials,omitempty"`
	KeyCredentials                     []KeyCredential      `json:"keyCredentials,omitempty"`
	ReplyUrls                          []string             `json:"replyUrls,omitempty"`
	LogoutURL                          string               `json:"logoutUrl,omitempty"`
	Homepage                           string               `json:"homepage,omitempty"`
	LoginURL                           string               `json:"loginUrl,omitempty"`
	PreferredTokenSigningKeyThumbprint string               `json:"preferredTokenSigningKeyThumbprint,omitempty"`
	PreferredSingleSignOnMode          string               `json:"preferredSingleSignOnMode,omitempty"`
	SamlMetadataURL                    string               `json:"samlMetadataUrl,omitempty"`
	Tags                               []string              `json:"tags,omitempty"`
	Owners                             []DirectoryObjectRef  `json:"-"`
	VerifiedPublisher                  *VerifiedPublisher    `json:"verifiedPublisher,omitempty"`
	NotificationEmailAddresses         []string             `json:"notificationEmailAddresses,omitempty"`
	CreatedDateTime                    *time.Time           `json:"createdDateTime,omitempty"`
	DeletedDateTime                    *time.Time           `json:"deletedDateTime,omitempty"`
	ModifiedAt                         time.Time            `json:"-"`
}
