package model

type Token struct {
	Kind            string   `json:"kind"`
	Etag            string   `json:"etag,omitempty"`
	ClientId        string   `json:"clientId,omitempty"`
	DisplayText     string   `json:"displayText,omitempty"`
	Scopes          []string `json:"scopes,omitempty"`
	UserKey         string   `json:"userKey,omitempty"`
	Anonymous       bool     `json:"anonymous"`
	ApplicationName string   `json:"applicationName,omitempty"`
}

type ASP struct {
	Kind               string `json:"kind"`
	Etag               string `json:"etag,omitempty"`
	CodeId             int64  `json:"codeId"`
	CreationTime       string `json:"creationTime,omitempty"`
	Name               string `json:"name,omitempty"`
	UserKey            string `json:"userKey,omitempty"`
	LastTimeUsed       string `json:"lastTimeUsed,omitempty"`
	LastIp             string `json:"lastIp,omitempty"`
}

type VerificationCode struct {
	Kind                    string `json:"kind"`
	Etag                    string `json:"etag,omitempty"`
	UserId                  string `json:"userId,omitempty"`
	VerificationCode        string `json:"verificationCode,omitempty"`
	VerificationMethod      string `json:"verificationMethod,omitempty"`
	VerificationTimestamp   string `json:"verificationTimestamp,omitempty"`
}

type UserInvitation struct {
	Name         string `json:"name"`
	State        string `json:"state,omitempty"`
	InvitedEmail string `json:"invitedEmail,omitempty"`
	Mails        []Mail `json:"mails,omitempty"`
}

type Mail struct {
	Subject string `json:"subject,omitempty"`
	Body    string `json:"body,omitempty"`
	Date    string `json:"date,omitempty"`
}