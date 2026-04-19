package store

// Client represents a registered OAuth client.
type Client struct {
	ClientID     string
	ClientSecret string
	TenantID     string
}
