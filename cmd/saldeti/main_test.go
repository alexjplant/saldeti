package main

import "testing"

func TestValidateMode(t *testing.T) {
	tests := []struct {
		name    string
		mode    string
		wantErr bool
	}{
		{name: "entra valid", mode: "entra", wantErr: false},
		{name: "google valid", mode: "google", wantErr: false},
		{name: "empty invalid", mode: "", wantErr: true},
		{name: "invalid mode", mode: "invalid", wantErr: true},
		{name: "both invalid", mode: "both", wantErr: true}, // "both" was a prior mode, now invalid (mutually exclusive)
		{name: "Entra case-sensitive invalid", mode: "Entra", wantErr: true},
		{name: "Google case-sensitive invalid", mode: "Google", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMode(tt.mode)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateMode(%q) error = %v, wantErr %v", tt.mode, err, tt.wantErr)
			}
		})
	}
}
