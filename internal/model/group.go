package model

import "time"

type AssignedLicense struct {
    DisabledPlans []string `json:"disabledPlans"`
    SkuID         string   `json:"skuId"`
}

type Group struct {
    ODataType                   string            `json:"@odata.type,omitempty"`
    ID                          string            `json:"id,omitempty"`
    DisplayName                 string            `json:"displayName,omitempty"`
    Description                 string            `json:"description,omitempty"`
    MailNickname                string            `json:"mailNickname,omitempty"`
    Mail                        string            `json:"mail,omitempty"`
    MailEnabled                 *bool             `json:"mailEnabled,omitempty"`
    SecurityEnabled             *bool             `json:"securityEnabled,omitempty"`
    Visibility                  string            `json:"visibility,omitempty"`    // Public, Private, HiddenMembership
    GroupTypes                  []string          `json:"groupTypes,omitempty"`    // e.g. ["Unified"]
    MembershipRule              string            `json:"membershipRule,omitempty"`
    MembershipRuleProcessingState string          `json:"membershipRuleProcessingState,omitempty"`
    CreatedDateTime             *time.Time        `json:"createdDateTime,omitempty"`
    DeletedDateTime             *time.Time        `json:"deletedDateTime,omitempty"`
    IsAssignableToRole          *bool             `json:"isAssignableToRole,omitempty"`
    AssignedLicenses            []AssignedLicense `json:"assignedLicenses,omitempty"`
    PreferredLanguage           string            `json:"preferredLanguage,omitempty"`
    ProxyAddresses              []string          `json:"proxyAddresses,omitempty"`
    RenewedDateTime             *time.Time        `json:"renewedDateTime,omitempty"`
    Theme                       string            `json:"theme,omitempty"`
    UniqueName                  string            `json:"uniqueName,omitempty"`
    // Nested membership for inline creation (not serialized in GET responses)
    Members                     []DirectoryObjectRef `json:"-"`
    Owners                      []DirectoryObjectRef `json:"-"`
    ModifiedAt                  time.Time            `json:"-"`
}

type DirectoryObjectRef struct {
    ODataID string `json:"@odata.id,omitempty"`
}