package model

type Customer struct {
	Kind           string         `json:"kind"`
	ID             string         `json:"id"`
	Etag           string         `json:"etag,omitempty"`
	CustomerDomain string         `json:"customerDomain,omitempty"`
	PrimaryDomain  string         `json:"primaryDomain,omitempty"`
	CustomerName   string         `json:"customerName,omitempty"`
	PostalAddress  *PostalAddress `json:"postalAddress,omitempty"`
	Language       string         `json:"language,omitempty"`
	AdminCreated   bool           `json:"adminCreated"`
	PhoneNumber    string         `json:"phoneNumber,omitempty"`
	AlternateEmail string         `json:"alternateEmail,omitempty"`
}

type PostalAddress struct {
	ContactName      string `json:"contactName,omitempty"`
	OrganizationName string `json:"organizationName,omitempty"`
	AddressLine1     string `json:"addressLine1,omitempty"`
	AddressLine2     string `json:"addressLine2,omitempty"`
	AddressLine3     string `json:"addressLine3,omitempty"`
	Locality         string `json:"locality,omitempty"`
	Region           string `json:"region,omitempty"`
	PostalCode       string `json:"postalCode,omitempty"`
	CountryCode      string `json:"countryCode,omitempty"`
}
