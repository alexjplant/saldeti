package model

type Member struct {
	Kind             string `json:"kind"`
	ID               string `json:"id"`
	Etag             string `json:"etag,omitempty"`
	Email            string `json:"email"`
	Role             string `json:"role,omitempty"`
	Type             string `json:"type,omitempty"`
	Status           string `json:"status,omitempty"`
	DeliverySettings string `json:"delivery_settings,omitempty"`
}