package model

type Schema struct {
	Kind       string             `json:"kind"`
	Etag       string             `json:"etag,omitempty"`
	SchemaId   string             `json:"schemaId,omitempty"`
	SchemaName string             `json:"schemaName"`
	Fields     []SchemaFieldSpec  `json:"fields,omitempty"`
	DisplayName string            `json:"displayName,omitempty"`
}

type SchemaFieldSpec struct {
	FieldId            string              `json:"fieldId,omitempty"`
	FieldName          string              `json:"fieldName"`
	FieldType          string              `json:"fieldType"`
	FieldReadAccessType string             `json:"fieldReadAccessType,omitempty"`
	MultiValued        bool                `json:"multiValued,omitempty"`
	Kind               string              `json:"kind,omitempty"`
	Etag               string              `json:"etag,omitempty"`
	NumericIndexingSpec *NumericIndexingSpec `json:"numericIndexingSpec,omitempty"`
	DisplayName        string              `json:"displayName,omitempty"`
	Indexed            bool                `json:"indexed,omitempty"`
	Required           bool                `json:"required,omitempty"`
}

type NumericIndexingSpec struct {
	MinValue float64 `json:"minValue,omitempty"`
	MaxValue float64 `json:"maxValue,omitempty"`
}

type CalendarResource struct {
	Kind                  string           `json:"kind"`
	Etag                  string           `json:"etag,omitempty"`
	ResourceId            string           `json:"resourceId"`
	ResourceName          string           `json:"resourceName,omitempty"`
	ResourceType          string           `json:"resourceType,omitempty"`
	ResourceDescription   string           `json:"resourceDescription,omitempty"`
	ResourceCategory      string           `json:"resourceCategory,omitempty"`
	UserVisibleDescription string           `json:"userVisibleDescription,omitempty"`
	GeneratedResourceName string           `json:"generatedResourceName,omitempty"`
	BuildingId            string           `json:"buildingId,omitempty"`
	FloorName             string           `json:"floorName,omitempty"`
	FloorSection          string           `json:"floorSection,omitempty"`
	Capacity              int64            `json:"capacity,omitempty"`
	FeatureInstances      []FeatureInstance `json:"featureInstances,omitempty"`
}

type FeatureInstance struct {
	Feature *FeatureRef `json:"feature,omitempty"`
}

type FeatureRef struct {
	Name string `json:"name,omitempty"`
}

type Building struct {
	Kind          string          `json:"kind"`
	Etag          string          `json:"etag,omitempty"`
	BuildingId    string          `json:"buildingId"`
	BuildingName  string          `json:"buildingName,omitempty"`
	Description   string          `json:"description,omitempty"`
	Address       *Address        `json:"address,omitempty"`
	FloorNames    []string        `json:"floorNames,omitempty"`
	Coordinates   *Coordinates    `json:"coordinates,omitempty"`
	BuildingFloors []BuildingFloor `json:"buildingFloors,omitempty"`
}

type Address struct {
	Locality     string `json:"locality,omitempty"`
	Region       string `json:"region,omitempty"`
	PostalCode   string `json:"postalCode,omitempty"`
	CountryCode  string `json:"countryCode,omitempty"`
	AddressLine1 string `json:"addressLine1,omitempty"`
	AddressLine2 string `json:"addressLine2,omitempty"`
	AddressLine3 string `json:"addressLine3,omitempty"`
	LanguageCode string `json:"languageCode,omitempty"`
	FullAddress  string `json:"fullAddress,omitempty"`
}

type Coordinates struct {
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

type BuildingFloor struct {
	FloorName    string   `json:"floorName,omitempty"`
	FloorSection []string `json:"floorSection,omitempty"`
}

type Feature struct {
	Kind       string `json:"kind"`
	Etag       string `json:"etag,omitempty"`
	FeatureName string `json:"featureName"`
}

type GroupSettings struct {
	Kind                                  string `json:"kind"`
	Etag                                  string `json:"etag,omitempty"`
	Email                                 string `json:"email,omitempty"`
	Name                                  string `json:"name,omitempty"`
	Description                           string `json:"description,omitempty"`
	WhoCanJoin                            string `json:"whoCanJoin,omitempty"`
	WhoCanViewMembership                  string `json:"whoCanViewMembership,omitempty"`
	WhoCanViewGroup                       string `json:"whoCanViewGroup,omitempty"`
	WhoCanInvite                          string `json:"whoCanInvite,omitempty"`
	WhoCanAdd                             string `json:"whoCanAdd,omitempty"`
	WhoCanPostMessage                     string `json:"whoCanPostMessage,omitempty"`
	WhoCanPostAnnouncements               string `json:"whoCanPostAnnouncements,omitempty"`
	WhoCanReply                           string `json:"whoCanReply,omitempty"`
	WhoCanModerateMembers                 string `json:"whoCanModerateMembers,omitempty"`
	WhoCanModerateContent                 string `json:"whoCanModerateContent,omitempty"`
	WhoCanAssistContent                   string `json:"whoCanAssistContent,omitempty"`
	CustomReplyTo                         string `json:"customReplyTo,omitempty"`
	SendReplyToOwner                      bool   `json:"sendReplyToOwner,omitempty"`
	IncludeCustomFooter                   bool   `json:"includeCustomFooter,omitempty"`
	CustomFooterText                      string `json:"customFooterText,omitempty"`
	AllowExternalMembers                  bool   `json:"allowExternalMembers,omitempty"`
	MaxMessageBytes                       int64  `json:"maxMessageBytes,omitempty"`
	IsArchived                            bool   `json:"isArchived,omitempty"`
	ArchiveOnly                           bool   `json:"archiveOnly,omitempty"`
	MessageModerationLevel                string `json:"messageModerationLevel,omitempty"`
	SpamModerationLevel                   string `json:"spamModerationLevel,omitempty"`
	PrimaryLanguage                       string `json:"primaryLanguage,omitempty"`
	DefaultMessageDenyNotificationText    string `json:"defaultMessageDenyNotificationText,omitempty"`
	ShowInGroupDirectory                  bool   `json:"showInGroupDirectory,omitempty"`
	AllowGoogleCommunication              bool   `json:"allowGoogleCommunication,omitempty"`
	AutoAddNewUsersWithContact            bool   `json:"autoAddNewUsersWithContact,omitempty"`
	SendMessageDenyNotification           bool   `json:"sendMessageDenyNotification,omitempty"`
	DefaultSender                         string `json:"defaultSender,omitempty"`
	WhoCanContactOwner                    string `json:"whoCanContactOwner,omitempty"`
	WhoCanApproveMembers                  string `json:"whoCanApproveMembers,omitempty"`
	WhoCanBanUsers                        string `json:"whoCanBanUsers,omitempty"`
	WhoCanLeaveGroup                      string `json:"whoCanLeaveGroup,omitempty"`
	EnableCollaborativeInbox              bool   `json:"enableCollaborativeInbox,omitempty"`
	FavoriteRepliesOnTop                  bool   `json:"favoriteRepliesOnTop,omitempty"`
	WhoCanMarkFavoriteReplyOnAnyTopic     string `json:"whoCanMarkFavoriteReplyOnAnyTopic,omitempty"`
	WhoCanMarkNoResponseNeeded            string `json:"whoCanMarkNoResponseNeeded,omitempty"`
	WhoCanMarkDuplicate                   string `json:"whoCanMarkDuplicate,omitempty"`
	WhoCanTakeTopics                      string `json:"whoCanTakeTopics,omitempty"`
	WhoCanUnassignTopic                   string `json:"whoCanUnassignTopic,omitempty"`
	WhoCanUnmarkAnyTopic                  string `json:"whoCanUnmarkAnyTopic,omitempty"`
	WhoCanEnterFreeFormTags               string `json:"whoCanEnterFreeFormTags,omitempty"`
	WhoCanModifyTagsAndCategories         string `json:"whoCanModifyTagsAndCategories,omitempty"`
	WhoCanAssignTopics                    string `json:"whoCanAssignTopics,omitempty"`
}

type DataTransfer struct {
	Kind                      string                     `json:"kind"`
	Etag                      string                     `json:"etag,omitempty"`
	Id                        string                     `json:"id,omitempty"`
	OldOwnerUserId            string                     `json:"oldOwnerUserId,omitempty"`
	NewOwnerUserId            string                     `json:"newOwnerUserId,omitempty"`
	ApplicationDataTransfers  []ApplicationDataTransfer  `json:"applicationDataTransfers,omitempty"`
	RequestTime               string                     `json:"requestTime,omitempty"`
	OverallTransferStatusCode string                     `json:"overallTransferStatusCode,omitempty"`
}

type ApplicationDataTransfer struct {
	ApplicationId            string               `json:"applicationId,omitempty"`
	ApplicationTransferId    string               `json:"applicationTransferId,omitempty"`
	ApplicationTransferStatus string               `json:"applicationTransferStatus,omitempty"`
	ApplicationTransferParams map[string][]string `json:"applicationTransferParams,omitempty"`
}

type TransferApplication struct {
	Kind          string        `json:"kind"`
	Etag          string        `json:"etag,omitempty"`
	Id            string        `json:"id"`
	Name          string        `json:"name,omitempty"`
	TransferParams []TransferParam `json:"transferParams,omitempty"`
}

type TransferParam struct {
	Key   string   `json:"key,omitempty"`
	Value []string `json:"value,omitempty"`
}

type Subscription struct {
	Name                    string                 `json:"name"`
	Uid                     string                 `json:"uid,omitempty"`
	TargetResource          string                 `json:"targetResource,omitempty"`
	EventTypes              []string               `json:"eventTypes,omitempty"`
	PayloadOptions          *PayloadOptions        `json:"payloadOptions,omitempty"`
	NotificationEndpoint    *NotificationEndpoint  `json:"notificationEndpoint,omitempty"`
	State                   string                 `json:"state,omitempty"`
	ErrorType               string                 `json:"errorType,omitempty"`
	CreateTime              string                 `json:"createTime,omitempty"`
	UpdateTime              string                 `json:"updateTime,omitempty"`
	Authority               string                 `json:"authority,omitempty"`
	Etag                    string                 `json:"etag,omitempty"`
	Reconciling             bool                   `json:"reconciling,omitempty"`
	SuspensionReason        string                 `json:"suspensionReason,omitempty"`
	ExpireTime              string                 `json:"expireTime,omitempty"`
	Ttl                     string                 `json:"ttl,omitempty"`
}

type PayloadOptions struct {
	IncludeResource bool   `json:"includeResource,omitempty"`
	FieldMask       string `json:"fieldMask,omitempty"`
}

type NotificationEndpoint struct {
	Endpoint string `json:"endpoint,omitempty"`
}