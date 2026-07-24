package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMintAndValidate(t *testing.T) {
	SetSigningKey([]byte("test-secret-key-that-is-32bytes!"))

	scopes := []string{"openid", "profile", "email", "https://www.googleapis.com/auth/admin.directory.user"}
	tokenString, err := MintToken("saldeti-google", "user@example.com", "user@example.com", scopes, time.Hour)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	claims, err := ValidateToken(tokenString)
	require.NoError(t, err)
	assert.Equal(t, "saldeti-google", claims.Iss)
	assert.Equal(t, "https://oauth2.googleapis.com/token", claims.Aud)
	assert.Equal(t, "user@example.com", claims.Sub)
	assert.Equal(t, "user@example.com", claims.Email)
	assert.Equal(t, strings.Join(scopes, " "), claims.Scope)
	assert.ElementsMatch(t, scopes, claims.Scopes)
}

func TestValidateInvalidToken(t *testing.T) {
	SetSigningKey([]byte("test-secret-key-that-is-32bytes!"))

	_, err := ValidateToken("this.is.not.a.valid.token")
	assert.Error(t, err)
}

func TestScopeFiltering(t *testing.T) {
	scopes := []string{
		"https://www.googleapis.com/auth/admin.directory.user",
		"https://www.googleapis.com/auth/admin.directory.group",
		"unknown_scope",
		"openid",
		"another_unknown",
	}

	result := FilterKnownScopes(scopes)
	require.Len(t, result, 3)
	assert.Contains(t, result, "https://www.googleapis.com/auth/admin.directory.user")
	assert.Contains(t, result, "https://www.googleapis.com/auth/admin.directory.group")
	assert.Contains(t, result, "openid")
	assert.NotContains(t, result, "unknown_scope")
	assert.NotContains(t, result, "another_unknown")
}

func TestRefreshTokenFlow(t *testing.T) {
	SetSigningKey([]byte("test-secret-key-that-is-32bytes!"))

	// Clear refresh tokens
	refreshTokensMutex.Lock()
	refreshTokens = make(map[string]refreshTokenClaims)
	refreshTokensMutex.Unlock()

	scopes := []string{"openid", "profile", "email"}

	// Step 1: Generate a refresh token
	rt, err := GenerateRefreshToken("test-client-id", "user@example.com", scopes)
	require.NoError(t, err)
	assert.NotEmpty(t, rt)

	// Step 2: Verify stored correctly
	refreshTokensMutex.RLock()
	entry, exists := refreshTokens[rt]
	refreshTokensMutex.RUnlock()
	require.True(t, exists)
	assert.Equal(t, "test-client-id", entry.ClientID)
	assert.Equal(t, "user@example.com", entry.Subject)
	assert.Equal(t, scopes, entry.Scopes)

	// Step 3: Simulate refresh — mint new access token using stored info
	finalScopes := entry.Scopes
	accessToken, err := MintToken("saldeti-google", entry.Subject, "", finalScopes, time.Hour)
	require.NoError(t, err)

	// Step 4: Validate the new access token
	claims, err := ValidateToken(accessToken)
	require.NoError(t, err)
	assert.Equal(t, "saldeti-google", claims.Iss)
	assert.Equal(t, "user@example.com", claims.Sub)
	assert.Equal(t, "https://oauth2.googleapis.com/token", claims.Aud)
	assert.ElementsMatch(t, scopes, claims.Scopes)
}

func TestExpiredToken(t *testing.T) {
	SetSigningKey([]byte("test-secret-key-that-is-32bytes!"))

	// Mint a token that expired 1 hour ago
	token, err := MintToken("saldeti-google", "user@example.com", "user@example.com", []string{"openid"}, -time.Hour)
	require.NoError(t, err)

	_, err = ValidateToken(token)
	assert.Error(t, err)
}
