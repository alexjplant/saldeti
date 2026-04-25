package model

import "time"

// Application represents an Azure AD application registration
type Application struct {
	ODataType                  string                `json:"@odata.type,omitempty"`
	ID                         string                `json:"id,omitempty"`
	AppID                      string                `json:"appId,omitempty"`
	DisplayName                string                `json:"displayName,omitempty"`
	Description                string                `json:"description,omitempty"`
	SignInAudience             string                `json:"signInAudience,omitempty"`
	IdentifierUris             []string              `json:"identifierUris,omitempty"`
	API                        *ApplicationAPI       `json:"api,omitempty"`
	Web                        *ApplicationWeb       `json:"web,omitempty"`
	Spa                        *ApplicationSpa       `json:"spa,omitempty"`
	PublicClient                *ApplicationPublicClient `json:"publicClient,omitempty"`
	RequiredResourceAccess     []RequiredResourceAccess `json:"requiredResourceAccess,omitempty"`
	PasswordCredentials        []PasswordCredential  `json:"passwordCredentials,omitempty"`
	KeyCredentials             []KeyCredential       `json:"keyCredentials,omitempty"`
	AppRoles                   []AppRole             `json:"appRoles,omitempty"`
	OptionalClaims             *OptionalClaims       `json:"optionalClaims,omitempty"`
	Owners                     []DirectoryObjectRef  `json:"-"`
	Tags                       []string              `json:"tags,omitempty"`
	IsDeviceOnlyAuthSupported  *bool                 `json:"isDeviceOnlyAuthSupported,omitempty"`
	IsFallbackPublicClient     *bool                 `json:"isFallbackPublicClient,omitempty"`
	CreatedDateTime            *time.Time            `json:"createdDateTime,omitempty"`
	DeletedDateTime            *time.Time            `json:"deletedDateTime,omitempty"`
	PublisherDomain            string                `json:"publisherDomain,omitempty"`
	VerifiedPublisher          *VerifiedPublisher    `json:"verifiedPublisher,omitempty"`
	Certification                       *Certification        `json:"certification,omitempty"`
	OAuth2RequirePostResponse           *bool                 `json:"oauth2RequirePostResponse,omitempty"`
	NativeAuthenticationApisEnabled     *string               `json:"nativeAuthenticationApisEnabled,omitempty"`
	GroupMembershipClaims               *string               `json:"groupMembershipClaims,omitempty"`
	TokenEncryptionKeyId                *string               `json:"tokenEncryptionKeyId,omitempty"`
	Notes                               *string               `json:"notes,omitempty"`
	Info                                *InformationalURL     `json:"info,omitempty"`
	ModifiedAt                          time.Time             `json:"-"`
}

// ApplicationAPI contains API settings for an application
type ApplicationAPI struct {
	AcceptMappedClaims         *bool               `json:"acceptMappedClaims,omitempty"`
	KnownClientApplications    []string            `json:"knownClientApplications,omitempty"`
	OAuth2PermissionScopes     []PermissionScope   `json:"oauth2PermissionScopes,omitempty"`
	PreAuthorizedApplications  []PreAuthorizedApplication `json:"preAuthorizedApplications,omitempty"`
	RequestedAccessTokenVersion *int               `json:"requestedAccessTokenVersion,omitempty"`
}

// ApplicationWeb contains web settings for an application
type ApplicationWeb struct {
	HomePageURL          string   `json:"homePageUrl,omitempty"`
	ImplicitGrantSettings *ImplicitGrantSettings `json:"implicitGrantSettings,omitempty"`
	LogoutURL            string   `json:"logoutUrl,omitempty"`
	RedirectUris         []string `json:"redirectUris,omitempty"`
}

// ApplicationSpa contains SPA settings for an application
type ApplicationSpa struct {
	RedirectUris []string `json:"redirectUris,omitempty"`
}

// ApplicationPublicClient contains public client settings for an application
type ApplicationPublicClient struct {
	RedirectUris []string `json:"redirectUris,omitempty"`
}

// RequiredResourceAccess represents a resource access requirement
type RequiredResourceAccess struct {
	ResourceAccess []ResourceAccess `json:"resourceAccess,omitempty"`
	ResourceAppID  string           `json:"resourceAppId,omitempty"`
}

// ResourceAccess represents a specific resource access permission
type ResourceAccess struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}

// OptionalClaims represents optional claims configuration
type OptionalClaims struct {
	AccessToken []OptionalClaim `json:"accessToken,omitempty"`
	IDToken     []OptionalClaim `json:"idToken,omitempty"`
	SamlToken   []OptionalClaim `json:"samlToken,omitempty"`
}

// OptionalClaim represents a single optional claim
type OptionalClaim struct {
	AdditionalProperties []string `json:"additionalProperties,omitempty"`
	Essential            *bool    `json:"essential,omitempty"`
	Name                 string   `json:"name,omitempty"`
	Source               string   `json:"source,omitempty"`
}

// VerifiedPublisher represents verified publisher information
type VerifiedPublisher struct {
	AddedDateTime       *time.Time `json:"addedDateTime,omitempty"`
	DisplayName         string     `json:"displayName,omitempty"`
	VerifiedPublisherID  string     `json:"verifiedPublisherId,omitempty"`
}

// Certification represents certification information
type Certification struct {
	CertificationDetails             string     `json:"certificationDetails,omitempty"`
	CertificationExpirationDateTime  *time.Time `json:"certificationExpirationDateTime,omitempty"`
	IsCertifiedByMicrosoft           *bool      `json:"isCertifiedByMicrosoft,omitempty"`
	IsPublisherAttested              *bool      `json:"isPublisherAttested,omitempty"`
	LastCertificateUpdateDateTime    *time.Time `json:"lastCertificateUpdateDateTime,omitempty"`
	NextCertificateUpdateDateTime    *time.Time `json:"nextCertificateUpdateDateTime,omitempty"`
}

// InformationalURL represents informational URLs for an application
type InformationalURL struct {
	LogoURL              string `json:"logoUrl,omitempty"`
	MarketingURL         string `json:"marketingUrl,omitempty"`
	PrivacyStatementURL  string `json:"privacyStatementUrl,omitempty"`
	SupportURL           string `json:"supportUrl,omitempty"`
	TermsOfServiceURL    string `json:"termsOfServiceUrl,omitempty"`
}

// PreAuthorizedApplication represents a pre-authorized application
type PreAuthorizedApplication struct {
	AppID           string   `json:"appId,omitempty"`
	PermissionIDs   []string `json:"permissionIds,omitempty"`
	DelegatedPermissionIDs []string `json:"delegatedPermissionIds,omitempty"`
}

// ImplicitGrantSettings represents implicit grant settings
type ImplicitGrantSettings struct {
	EnableAccessTokenIssuance *bool `json:"enableAccessTokenIssuance,omitempty"`
	EnableIDTokenIssuance     *bool `json:"enableIdTokenIssuance,omitempty"`
}
