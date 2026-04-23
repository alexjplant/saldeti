package model

import "time"

type User struct {
	ODataType                  string          `json:"@odata.type,omitempty"`
	ID                         string          `json:"id,omitempty"`
	DisplayName                string          `json:"displayName,omitempty"`
	GivenName                  string          `json:"givenName,omitempty"`
	Surname                    string          `json:"surname,omitempty"`
	UserPrincipalName          string          `json:"userPrincipalName,omitempty"`
	Mail                       string          `json:"mail,omitempty"`
	MailNickname               string          `json:"mailNickname,omitempty"`
	JobTitle                   string          `json:"jobTitle,omitempty"`
	Department                 string          `json:"department,omitempty"`
	OfficeLocation             string          `json:"officeLocation,omitempty"`
	MobilePhone                string          `json:"mobilePhone,omitempty"`
	BusinessPhones             []string        `json:"businessPhones"`
	AccountEnabled             *bool           `json:"accountEnabled,omitempty"`
	CreatedDateTime            *time.Time      `json:"createdDateTime,omitempty"`
	DeletedDateTime            *time.Time      `json:"deletedDateTime,omitempty"`
	LastPasswordChangeDateTime *time.Time      `json:"lastPasswordChangeDateTime,omitempty"`
	PasswordPolicies           string          `json:"passwordPolicies,omitempty"`
	UsageLocation              string          `json:"usageLocation,omitempty"`
	UserType                   string          `json:"userType,omitempty"`
	OnPremisesImmutableId      string          `json:"onPremisesImmutableId,omitempty"`
	OnPremisesLastSyncDateTime *time.Time      `json:"onPremisesLastSyncDateTime,omitempty"`
	OnPremisesSyncEnabled      *bool           `json:"onPremisesSyncEnabled,omitempty"`
	PreferredLanguage          string          `json:"preferredLanguage,omitempty"`
	AgeGroup                   string          `json:"ageGroup,omitempty"`
	ConsentProvidedForMinor    string          `json:"consentProvidedForMinor,omitempty"`
	LegalAgeGroupClassification string         `json:"legalAgeGroupClassification,omitempty"`
	PasswordProfile            *PasswordProfile `json:"passwordProfile,omitempty"`
	AssignedLicenses           []AssignedLicense `json:"assignedLicenses,omitempty"`
	ModifiedAt                 time.Time        `json:"-"`
}

type PasswordProfile struct {
	ODataType                             string `json:"@odata.type,omitempty"`
	Password                              string `json:"-"`
	ForceChangePasswordNextSignIn         *bool  `json:"forceChangePasswordNextSignIn,omitempty"`
	ForceChangePasswordNextSignInWithMfa  *bool  `json:"forceChangePasswordNextSignInWithMfa,omitempty"`
}