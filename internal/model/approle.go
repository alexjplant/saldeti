package model

import "time"

// AppRole represents an application role that can be assigned to users, groups, or service principals
type AppRole struct {
	ID                 string   `json:"id,omitempty"`
	AllowedMemberTypes []string `json:"allowedMemberTypes,omitempty"`
	Description        string   `json:"description,omitempty"`
	DisplayName        string   `json:"displayName,omitempty"`
	IsEnabled          *bool    `json:"isEnabled,omitempty"`
	Origin             string   `json:"origin,omitempty"`
	Value              string   `json:"value,omitempty"`
}

// OAuth2PermissionGrant represents an OAuth2 permission grant (delegated permission)
type OAuth2PermissionGrant struct {
	ODataType   string `json:"@odata.type,omitempty"`
	ID          string `json:"id,omitempty"`
	ClientID    string `json:"clientId,omitempty"`
	ConsentType string `json:"consentType,omitempty"` // "AllPrincipals" or "Principal"
	PrincipalID string `json:"principalId,omitempty"`
	ResourceID  string `json:"resourceId,omitempty"`
	Scope       string `json:"scope,omitempty"`
}

// AppRoleAssignment represents an assignment of an app role to a principal
type AppRoleAssignment struct {
	ODataType             string     `json:"@odata.type,omitempty"`
	ID                    string     `json:"id,omitempty"`
	AppRoleID             string     `json:"appRoleId,omitempty"`
	CreatedDateTime       *time.Time `json:"createdDateTime,omitempty"`
	PrincipalDisplayName  string     `json:"principalDisplayName,omitempty"`
	PrincipalID           string     `json:"principalId,omitempty"`
	PrincipalType         string     `json:"principalType,omitempty"`
	ResourceDisplayName   string     `json:"resourceDisplayName,omitempty"`
	ResourceID            string     `json:"resourceId,omitempty"`
}

// PasswordCredential represents a password credential for an application or service principal
type PasswordCredential struct {
	CustomKeyIdentifier string     `json:"customKeyIdentifier,omitempty"`
	DisplayName         string     `json:"displayName,omitempty"`
	KeyID               string     `json:"keyId,omitempty"`
	SecretText          string     `json:"secretText,omitempty"`
	Hint                string     `json:"hint,omitempty"`
	EndDateTime         *time.Time `json:"endDateTime,omitempty"`
	StartDateTime       *time.Time `json:"startDateTime,omitempty"`
}

// KeyCredential represents a key credential for an application or service principal
type KeyCredential struct {
	CustomKeyIdentifier string     `json:"customKeyIdentifier,omitempty"`
	DisplayName         string     `json:"displayName,omitempty"`
	KeyID               string     `json:"keyId,omitempty"`
	Type                string     `json:"type,omitempty"`
	Usage               string     `json:"usage,omitempty"`
	Key                 string     `json:"key,omitempty"`
	StartDateTime       *time.Time `json:"startDateTime,omitempty"`
	EndDateTime         *time.Time `json:"endDateTime,omitempty"`
}

// ExtensionProperty represents a directory extension property
type ExtensionProperty struct {
	ID            string   `json:"id,omitempty"`
	Name          string   `json:"name,omitempty"`
	DataType      string   `json:"dataType,omitempty"`
	TargetObjects []string `json:"targetObjects,omitempty"`
	AppDisplayName string  `json:"appDisplayName,omitempty"`
}

// ValidExtensionDataTypes defines the allowed values for ExtensionProperty.DataType
var ValidExtensionDataTypes = map[string]bool{
	"Binary":      true,
	"Boolean":     true,
	"DateTime":    true,
	"Integer":     true,
	"LargeString": true,
	"String":      true,
}

// PermissionScope represents an OAuth2 permission scope defined by a service principal
type PermissionScope struct {
	ID                       string `json:"id,omitempty"`
	IsEnabled                *bool  `json:"isEnabled,omitempty"`
	Type                     string `json:"type,omitempty"`
	UserConsentDescription   string `json:"userConsentDescription,omitempty"`
	UserConsentDisplayName   string `json:"userConsentDisplayName,omitempty"`
	AdminConsentDescription  string `json:"adminConsentDescription,omitempty"`
	AdminConsentDisplayName  string `json:"adminConsentDisplayName,omitempty"`
	Value                    string `json:"value,omitempty"`
}
