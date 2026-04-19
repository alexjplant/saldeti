package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Set a test signing key for all tests
	SetSigningKey([]byte("test-signing-key-32-bytes-long"))
	m.Run()
}

func TestMintToken(t *testing.T) {
	tenantID := "test-tenant"
	clientID := "test-client"
	subject := "test-subject"
	scopes := []string{"User.Read", "User.Read.All"}
	roles := []string{"Application"}
	lifetime := time.Hour

	token, err := MintToken(tenantID, clientID, subject, scopes, roles, lifetime)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the token
	claims, err := ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, tenantID, claims.TenantID)
	assert.Equal(t, clientID, claims.ClientID)
	assert.Equal(t, subject, claims.Subject)
	assert.Equal(t, scopes, claims.Scopes)
	assert.Equal(t, roles, claims.Roles)
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
}

func TestValidateToken_Invalid(t *testing.T) {
	// Test with invalid token
	_, err := ValidateToken("invalid.token.string")
	assert.Error(t, err)

	// Test with expired token
	token, err := MintToken("tenant", "client", "subject", []string{"scope"}, []string{"role"}, -time.Hour)
	require.NoError(t, err)
	
	_, err = ValidateToken(token)
	assert.Error(t, err)
}

func TestValidateToken_Tampered(t *testing.T) {
	token, err := MintToken("tenant", "client", "subject", []string{"scope"}, []string{"role"}, time.Hour)
	require.NoError(t, err)

	// Tamper with the token
	tampered := token[:len(token)-5] + "xxxxx"
	_, err = ValidateToken(tampered)
	assert.Error(t, err)
}