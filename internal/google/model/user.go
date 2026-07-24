package model

type User struct {
	Kind                       string         `json:"kind"`
	ID                         string         `json:"id"`
	Etag                       string         `json:"etag,omitempty"`
	PrimaryEmail               string         `json:"primaryEmail"`
	Name                       *UserName      `json:"name,omitempty"`
	GivenName                  string         `json:"-"`
	FamilyName                 string         `json:"-"`
	DisplayName                string         `json:"-"`
	IsAdmin                    bool           `json:"isAdmin"`
	IsDelegatedAdmin           bool           `json:"isDelegatedAdmin"`
	IsMailboxSetup             bool           `json:"isMailboxSetup"`
	Emails                     []UserEmail    `json:"emails,omitempty"`
	Aliases                    []string       `json:"aliases,omitempty"`
	NonEditableAliases         []string       `json:"nonEditableAliases,omitempty"`
	CustomerID                 string         `json:"customerId,omitempty"`
	OrgUnitPath                string         `json:"orgUnitPath,omitempty"`
	OrgUnitId                  string         `json:"orgUnitId,omitempty"`
	ThumbnailPhotoUrl          string         `json:"thumbnailPhotoUrl,omitempty"`
	ThumbnailPhotoEtag         string         `json:"thumbnailPhotoEtag,omitempty"`
	Photos                     []UserPhoto    `json:"photos,omitempty"`
	Addresses                  []any          `json:"addresses,omitempty"`
	Relations                  []any          `json:"relations,omitempty"`
	ExternalIds                []any          `json:"externalIds,omitempty"`
	Organizations              []any          `json:"organizations,omitempty"`
	Telephones                 []any          `json:"telephones,omitempty"`
	Suspended                  bool           `json:"suspended"`
	SuspensionReason           string         `json:"suspensionReason,omitempty"`
	Archived                   bool           `json:"archived"`
	ChangePasswordAtNextLogin  bool           `json:"changePasswordAtNextLogin"`
	IPWhitelisted              bool           `json:"ipWhitelisted"`
	IncludeInGlobalAddressList bool           `json:"includeInGlobalAddressList"`
	RecoveryEmail              string         `json:"recoveryEmail,omitempty"`
	RecoveryPhone              string         `json:"recoveryPhone,omitempty"`
	LastLoginTime              string         `json:"lastLoginTime,omitempty"`
	CreationTime               string         `json:"creationTime,omitempty"`
	DeletionTime               string         `json:"deletionTime,omitempty"`
	IsEnforcedIn2Sv            bool           `json:"isEnforcedIn2Sv"`
	IsEnrolledIn2Sv            bool           `json:"isEnrolledIn2Sv"`
	CustomSchemas              map[string]any `json:"customSchemas,omitempty"`
	Languages                  []any          `json:"languages,omitempty"`
	SSHPublicKeys              []any          `json:"sshPublicKeys,omitempty"`
	Websites                   []any          `json:"websites,omitempty"`
	Keywords                   []any          `json:"keywords,omitempty"`
	Locations                  []UserLocation `json:"locations,omitempty"`
}

type UserName struct {
	GivenName  string `json:"givenName,omitempty"`
	FamilyName string `json:"familyName,omitempty"`
	FullName   string `json:"fullName,omitempty"`
}

type UserEmail struct {
	Address    string `json:"address,omitempty"`
	CustomType string `json:"customType,omitempty"`
	Primary    bool   `json:"primary"`
	Type       string `json:"type,omitempty"`
}

type UserPhoto struct {
	Kind         string `json:"kind,omitempty"`
	Etag         string `json:"etag,omitempty"`
	PrimaryEmail string `json:"primaryEmail,omitempty"`
	PhotoData    string `json:"photoData,omitempty"`
	MimeType     string `json:"mimeType,omitempty"`
	Height       int64  `json:"height,omitempty"`
	Width        int64  `json:"width,omitempty"`
}

type UserLocation struct {
	Area         string `json:"area,omitempty"`
	BuildingId   string `json:"buildingId,omitempty"`
	CustomType   string `json:"customType,omitempty"`
	DeskCode     string `json:"deskCode,omitempty"`
	FloorName    string `json:"floorName,omitempty"`
	FloorSection string `json:"floorSection,omitempty"`
	Type         string `json:"type,omitempty"`
}

type UserAlias struct {
	Kind         string `json:"kind,omitempty"`
	Etag         string `json:"etag,omitempty"`
	Id           string `json:"id,omitempty"`
	PrimaryEmail string `json:"primaryEmail,omitempty"`
	Alias        string `json:"alias,omitempty"`
}
