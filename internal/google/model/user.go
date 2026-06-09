package model

type User struct {
	Kind                      string                 `json:"kind"`
	ID                        string                 `json:"id"`
	Etag                      string                 `json:"etag,omitempty"`
	PrimaryEmail              string                 `json:"primaryEmail"`
	Name                      *UserName              `json:"name,omitempty"`
	GivenName                 string                 `json:"givenName,omitempty"`
	FamilyName                string                 `json:"familyName,omitempty"`
	DisplayName               string                 `json:"displayName,omitempty"`
	IsAdmin                   bool                   `json:"isAdmin,omitempty"`
	IsDelegatedAdmin          bool                   `json:"isDelegatedAdmin,omitempty"`
	IsMailboxSetup            bool                   `json:"isMailboxSetup,omitempty"`
	Emails                    []UserEmail            `json:"emails,omitempty"`
	Aliases                   []string               `json:"aliases,omitempty"`
	NonEditableAliases        []string               `json:"nonEditableAliases,omitempty"`
	CustomerID                string                 `json:"customerId,omitempty"`
	OrgUnitPath               string                 `json:"orgUnitPath,omitempty"`
	OrgUnitId                 string                 `json:"orgUnitId,omitempty"`
	ThumbnailPhotoUrl         string                 `json:"thumbnailPhotoUrl,omitempty"`
	ThumbnailPhotoEtag        string                 `json:"thumbnailPhotoEtag,omitempty"`
	Photos                    []UserPhoto            `json:"photos,omitempty"`
	Addresses                 []interface{}          `json:"addresses,omitempty"`
	Relations                 []interface{}          `json:"relations,omitempty"`
	ExternalIds               []interface{}          `json:"externalIds,omitempty"`
	Organizations             []interface{}          `json:"organizations,omitempty"`
	Telephones                []interface{}          `json:"telephones,omitempty"`
	Suspended                 bool                   `json:"suspended,omitempty"`
	SuspensionReason          string                 `json:"suspensionReason,omitempty"`
	Archived                  bool                   `json:"archived,omitempty"`
	ChangePasswordAtNextLogin bool                   `json:"changePasswordAtNextLogin,omitempty"`
	IPWhitelisted             bool                   `json:"ipWhitelisted,omitempty"`
	IncludeInGlobalAddressList bool                  `json:"includeInGlobalAddressList,omitempty"`
	RecoveryEmail             string                 `json:"recoveryEmail,omitempty"`
	RecoveryPhone             string                 `json:"recoveryPhone,omitempty"`
	LastLoginTime             string                 `json:"lastLoginTime,omitempty"`
	CreationTime              string                 `json:"creationTime,omitempty"`
	DeletionTime              string                 `json:"deletionTime,omitempty"`
	IsEnforcedIn2Sv           bool                   `json:"isEnforcedIn2Sv,omitempty"`
	IsEnrolledIn2Sv           bool                   `json:"isEnrolledIn2Sv,omitempty"`
	Is2svEnrolled             bool                   `json:"is2svEnrolled,omitempty"`
	CustomSchemas             map[string]interface{} `json:"customSchemas,omitempty"`
	Languages                 []interface{}          `json:"languages,omitempty"`
	SSHPublicKeys             []interface{}          `json:"sshPublicKeys,omitempty"`
	Websites                  []interface{}          `json:"websites,omitempty"`
	Keywords                  []interface{}          `json:"keywords,omitempty"`
	Locations                 []UserLocation         `json:"locations,omitempty"`
}

type UserName struct {
	GivenName  string `json:"givenName,omitempty"`
	FamilyName string `json:"familyName,omitempty"`
	FullName   string `json:"fullName,omitempty"`
}

type UserEmail struct {
	Address   string `json:"address,omitempty"`
	CustomType string `json:"customType,omitempty"`
	Primary   bool   `json:"primary,omitempty"`
	Type      string `json:"type,omitempty"`
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
	Area          string `json:"area,omitempty"`
	BuildingId    string `json:"buildingId,omitempty"`
	CustomType    string `json:"customType,omitempty"`
	DeskCode      string `json:"deskCode,omitempty"`
	FloorName     string `json:"floorName,omitempty"`
	FloorSection  string `json:"floorSection,omitempty"`
	Type          string `json:"type,omitempty"`
}

type UserAlias struct {
	Kind         string `json:"kind,omitempty"`
	Etag         string `json:"etag,omitempty"`
	Id           string `json:"id,omitempty"`
	PrimaryEmail string `json:"primaryEmail,omitempty"`
	Alias        string `json:"alias,omitempty"`
}