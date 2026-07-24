package model

type Activity struct {
	Kind        string          `json:"kind"`
	Etag        string          `json:"etag,omitempty"`
	OwnerDomain string          `json:"ownerDomain,omitempty"`
	IP          string          `json:"ip_address,omitempty"`
	Events      []ActivityEvent `json:"events,omitempty"`
	Actor       ActivityActor   `json:"actor,omitempty"`
	Id          *ActivityId     `json:"id,omitempty"`
}

type ActivityId struct {
	Time            string `json:"time,omitempty"`
	UniqueQualifier string `json:"uniqueQualifier,omitempty"`
	ApplicationName string `json:"applicationName,omitempty"`
	CustomerId      string `json:"customerId,omitempty"`
}

type ActivityActor struct {
	Email     string `json:"email,omitempty"`
	ProfileId string `json:"profileId,omitempty"`
	KeyType   string `json:"keyType,omitempty"`
}

type ActivityEvent struct {
	Type       string              `json:"type,omitempty"`
	Name       string              `json:"name,omitempty"`
	Parameters []ActivityParameter `json:"parameters,omitempty"`
}

type ActivityParameter struct {
	Name         string        `json:"name,omitempty"`
	Value        string        `json:"value,omitempty"`
	MultiValue   []string      `json:"multiValue,omitempty"`
	IntValue     int64         `json:"intValue,omitempty"`
	BoolValue    bool          `json:"boolValue,omitempty"`
	MessageValue *MessageValue `json:"messageValue,omitempty"`
}

type MessageValue struct {
	Parameter []ActivityParameter `json:"parameter,omitempty"`
}

type UsageReport struct {
	Kind       string                 `json:"kind"`
	Etag       string                 `json:"etag,omitempty"`
	Date       string                 `json:"date,omitempty"`
	Entity     *UsageReportEntity     `json:"entity,omitempty"`
	Parameters []UsageReportParameter `json:"parameters,omitempty"`
}

type UsageReportEntity struct {
	Type       string `json:"type,omitempty"`
	CustomerId string `json:"customerId,omitempty"`
	UserEmail  string `json:"userEmail,omitempty"`
	EntityId   string `json:"entityId,omitempty"`
}

type UsageReportParameter struct {
	Name          string `json:"name,omitempty"`
	Value         string `json:"value,omitempty"`
	IntValue      int64  `json:"intValue,omitempty"`
	BoolValue     bool   `json:"boolValue,omitempty"`
	DatetimeValue string `json:"datetimeValue,omitempty"`
	StringValue   string `json:"stringValue,omitempty"`
}
