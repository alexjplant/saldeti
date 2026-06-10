package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/saldeti/saldeti/internal/google/store"
)

var (
	signingKey         []byte
	refreshTokens      = make(map[string]refreshTokenClaims)
	refreshTokensMutex sync.RWMutex
)

var knownScopes = map[string]bool{
	"https://www.googleapis.com/auth/admin.directory.user":           true,
	"https://www.googleapis.com/auth/admin.directory.group":          true,
	"https://www.googleapis.com/auth/admin.directory.device.chromeos": true,
	"https://www.googleapis.com/auth/admin.directory.device.mobile":  true,
	"https://www.googleapis.com/auth/cloud-identity.devices.readonly": true,
	"https://www.googleapis.com/auth/cloud-identity.groups":          true,
	"https://www.googleapis.com/auth/admin.reports.usage.readonly":   true,
	"https://www.googleapis.com/auth/admin.reports.audit.readonly":   true,
	"https://www.googleapis.com/auth/admin.datatransfer":             true,
	"https://www.googleapis.com/auth/apps.groups.settings":           true,
	"https://www.googleapis.com/auth/cloud-platform":                 true,
	"openid":   true,
	"profile":  true,
	"email":    true,
}

type GoogleClaims struct {
	Iss    string   `json:"iss"`
	Aud    string   `json:"aud"`
	Sub    string   `json:"sub"`
	Email  string   `json:"email,omitempty"`
	Scope  string   `json:"scope"`
	Scopes []string `json:"-"`
	jwt.RegisteredClaims
}

type refreshTokenClaims struct {
	ClientID  string
	Subject   string
	Scopes    []string
	ExpiresAt time.Time
}

// SetSigningKey sets the JWT signing key for Google token generation.
func SetSigningKey(key []byte) {
	signingKey = key
	// Generate a short hash of the key for logging
	hash := sha256.Sum256(key)
	shortHash := hex.EncodeToString(hash[:])[:16]
	if key == nil || len(key) == 0 {
		log.Warn().Msg("JWT signing key is empty")
	} else if len(key) < 32 {
		log.Warn().Int("key_len", len(key)).Msg("JWT signing key is less than 32 bytes (insecure)")
	} else {
		log.Info().Str("hash", shortHash).Msg("JWT signing key configured")
	}
}

// MintToken creates a new JWT token with the given parameters.
func MintToken(iss, sub, email string, scopes []string, lifetime time.Duration) (string, error) {
	if signingKey == nil {
		return "", errors.New("JWT signing key not configured")
	}
	now := time.Now()
	scopeString := strings.Join(scopes, " ")
	claims := GoogleClaims{
		Iss:    iss,
		Aud:    "https://oauth2.googleapis.com/token",
		Sub:    sub,
		Email:  email,
		Scope:  scopeString,
		Scopes: scopes,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    iss,
			Audience:  jwt.ClaimStrings{"https://oauth2.googleapis.com/token"},
			Subject:   sub,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(lifetime)),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signingKey)
}

// ValidateToken validates a JWT token and returns the claims.
func ValidateToken(tokenString string) (*GoogleClaims, error) {
	if signingKey == nil {
		return nil, errors.New("JWT signing key not configured")
	}
	token, err := jwt.ParseWithClaims(tokenString, &GoogleClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return signingKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*GoogleClaims); ok && token.Valid {
		// Populate Scopes from Scope field
		claims.Scopes = strings.Split(claims.Scope, " ")
		// Trim empty strings
		var filteredScopes []string
		for _, scope := range claims.Scopes {
			if scope != "" {
				filteredScopes = append(filteredScopes, scope)
			}
		}
		claims.Scopes = filteredScopes
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// FilterKnownScopes filters out unknown scopes (Google-like behavior: silently filter).
func FilterKnownScopes(scopes []string) []string {
	var filtered []string
	for _, scope := range scopes {
		if knownScopes[scope] {
			filtered = append(filtered, scope)
		}
	}
	return filtered
}

// GenerateRefreshToken generates a cryptographically random refresh token.
func GenerateRefreshToken(clientID, subject string, scopes []string) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	tokenID := hex.EncodeToString(bytes)

	refreshTokensMutex.Lock()
	defer refreshTokensMutex.Unlock()

	// Store refresh token with 24h TTL
	refreshTokens[tokenID] = refreshTokenClaims{
		ClientID:  clientID,
		Subject:   subject,
		Scopes:    scopes,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	return tokenID, nil
}

// TokenHandler returns a gin handler for the Google OAuth2 token endpoint.
func TokenHandler(st store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost {
			c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
			return
		}

		if err := c.Request.ParseForm(); err != nil {
			writeTokenError(c, "invalid_request", "Failed to parse form data")
			return
		}

		grantType := c.Request.FormValue("grant_type")

		switch grantType {
		case "client_credentials":
			handleClientCredentialsGoogle(c, st)
		case "authorization_code":
			handleAuthCodeGoogle(c, st)
		case "refresh_token":
			handleRefreshTokenGoogle(c, st)
		default:
			writeTokenError(c, "unsupported_grant_type", fmt.Sprintf("Grant type '%s' not supported", grantType))
		}
	}
}

func handleClientCredentialsGoogle(c *gin.Context, st store.Store) {
	clientID := c.Request.FormValue("client_id")
	clientSecret := c.Request.FormValue("client_secret")
	scope := c.Request.FormValue("scope")

	if clientID == "" || clientSecret == "" {
		writeTokenError(c, "invalid_request", "client_id and client_secret are required")
		return
	}

	// Validate client credentials
	storedSecret, err := st.GetClient(c.Request.Context(), clientID)
	if err != nil || storedSecret != clientSecret {
		writeTokenError(c, "invalid_client", "Invalid client credentials")
		return
	}

	// Parse scopes
	scopes := []string{}
	if scope != "" {
		scopes = strings.Split(scope, " ")
	}

	// Filter unknown scopes (Google-like behavior: silently filter)
	scopes = FilterKnownScopes(scopes)

	// Default scope if empty
	if len(scopes) == 0 {
		scopes = []string{"openid", "profile", "email", "https://www.googleapis.com/auth/admin.directory.user"}
	}

	// Mint token
	token, err := MintToken("saldeti-google", clientID, "", scopes, time.Hour)
	if err != nil {
		writeTokenError(c, "server_error", "Failed to mint token")
		return
	}

	// Generate refresh token
	refreshToken, err := GenerateRefreshToken(clientID, clientID, scopes)
	if err != nil {
		writeTokenError(c, "server_error", "Failed to generate refresh token")
		return
	}

	writeTokenResponse(c, token, time.Hour, refreshToken)
}

func handleAuthCodeGoogle(c *gin.Context, st store.Store) {
	code := c.Request.FormValue("code")
	clientID := c.Request.FormValue("client_id")
	clientSecret := c.Request.FormValue("client_secret")
	scope := c.Request.FormValue("scope")

	if code == "" || clientID == "" || clientSecret == "" {
		writeTokenError(c, "invalid_request", "code, client_id, and client_secret are required")
		return
	}

	// Validate client credentials
	storedSecret, err := st.GetClient(c.Request.Context(), clientID)
	if err != nil || storedSecret != clientSecret {
		writeTokenError(c, "invalid_client", "Invalid client credentials")
		return
	}

	// Look up user - code is the user's email or ID
	user, err := st.GetUser(c.Request.Context(), code)
	if err != nil {
		writeTokenError(c, "invalid_grant", "Invalid authorization code")
		return
	}

	// Parse scopes
	scopes := []string{}
	if scope != "" {
		scopes = strings.Split(scope, " ")
	}

	// Filter unknown scopes
	scopes = FilterKnownScopes(scopes)

	// Default scope if empty
	if len(scopes) == 0 {
		scopes = []string{"openid", "profile", "email", "https://www.googleapis.com/auth/admin.directory.user"}
	}

	// Mint token with user as subject
	token, err := MintToken("saldeti-google", user.PrimaryEmail, user.PrimaryEmail, scopes, time.Hour)
	if err != nil {
		writeTokenError(c, "server_error", "Failed to mint token")
		return
	}

	// Generate refresh token
	refreshToken, err := GenerateRefreshToken(clientID, user.PrimaryEmail, scopes)
	if err != nil {
		writeTokenError(c, "server_error", "Failed to generate refresh token")
		return
	}

	writeTokenResponse(c, token, time.Hour, refreshToken)
}

func handleRefreshTokenGoogle(c *gin.Context, st store.Store) {
	refreshToken := c.Request.FormValue("refresh_token")
	clientID := c.Request.FormValue("client_id")
	scope := c.Request.FormValue("scope")

	if refreshToken == "" {
		writeTokenError(c, "invalid_request", "refresh_token is required")
		return
	}

	// Parse scopes
	scopes := []string{}
	if scope != "" {
		scopes = strings.Split(scope, " ")
	}

	// Filter unknown scopes
	scopes = FilterKnownScopes(scopes)

	// Look up and validate refresh token under lock, then release before calling GenerateRefreshToken
	refreshTokensMutex.Lock()
	entry, exists := refreshTokens[refreshToken]
	if !exists {
		refreshTokensMutex.Unlock()
		writeTokenError(c, "invalid_grant", "Invalid refresh token")
		return
	}

	// Validate token hasn't expired
	if time.Now().After(entry.ExpiresAt) {
		delete(refreshTokens, refreshToken)
		refreshTokensMutex.Unlock()
		writeTokenError(c, "invalid_grant", "Refresh token has expired")
		return
	}

	// Validate client matches if provided
	if clientID != "" && entry.ClientID != clientID {
		refreshTokensMutex.Unlock()
		writeTokenError(c, "invalid_grant", "Refresh token client mismatch")
		return
	}

	// Determine scopes: use provided scopes, or fall back to original token scopes
	finalScopes := scopes
	if len(finalScopes) == 0 {
		finalScopes = entry.Scopes
	}

	// Refresh token rotation: invalidate old refresh token
	delete(refreshTokens, refreshToken)
	refreshTokensMutex.Unlock()

	// Mint new access token (no lock needed)
	token, err := MintToken("saldeti-google", entry.Subject, "", finalScopes, time.Hour)
	if err != nil {
		writeTokenError(c, "server_error", "Failed to mint token")
		return
	}

	// Generate new refresh token (acquires its own lock internally)
	newRefreshToken, err := GenerateRefreshToken(entry.ClientID, entry.Subject, finalScopes)
	if err != nil {
		writeTokenError(c, "server_error", "Failed to generate refresh token")
		return
	}

	writeTokenResponse(c, token, time.Hour, newRefreshToken)
}

func writeTokenError(c *gin.Context, errorCode, errorDescription string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error":             errorCode,
		"error_description": errorDescription,
	})
}

func writeTokenResponse(c *gin.Context, accessToken string, lifetime time.Duration, refreshToken ...string) {
	resp := gin.H{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   int(lifetime.Seconds()),
	}
	if len(refreshToken) > 0 && refreshToken[0] != "" {
		resp["refresh_token"] = refreshToken[0]
	}
	c.JSON(http.StatusOK, resp)
}

// StartRefreshTokenCleanup starts a background goroutine that periodically
// removes expired refresh tokens from the in-memory store.
func StartRefreshTokenCleanup(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("Refresh token cleanup stopped")
				return
			case <-ticker.C:
				refreshTokensMutex.Lock()
				now := time.Now()
				evicted := 0
				for id, entry := range refreshTokens {
					if now.After(entry.ExpiresAt) {
						delete(refreshTokens, id)
						evicted++
					}
				}
				refreshTokensMutex.Unlock()
				if evicted > 0 {
					log.Debug().Int("evicted", evicted).Int("remaining", len(refreshTokens)).Msg("Refresh token cleanup")
				}
			}
		}
	}()
}