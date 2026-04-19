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
	"github.com/rs/zerolog/log"
	"github.com/saldeti/saldeti/internal/store"
)

var (
	signingKey          []byte
	refreshTokens       = make(map[string]refreshTokenClaims)
	refreshTokensMutex  sync.RWMutex
)

var knownScopes = map[string]bool{
	"User.Read.All":        true,
	"User.ReadWrite.All":   true,
	"Group.Read.All":       true,
	"Group.ReadWrite.All":  true,
	"Directory.Read.All":   true,
	"Directory.ReadWrite.All": true,
	"openid":               true,
	"profile":              true,
	"offline_access":       true,
	"User.Read":            true,
	"User.ReadWrite":       true,
}

type TokenClaims struct {
	TenantID string   `json:"tid"`
	ClientID string   `json:"appid"`
	Subject  string   `json:"sub"`
	Scopes   []string `json:"scp"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

type refreshTokenClaims struct {
	TenantID  string
	ClientID  string
	Subject   string
	Scopes    []string
	Roles     []string
	ExpiresAt time.Time
}

// SetSigningKey sets the JWT signing key
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

func MintToken(tenantID, clientID, subject string, scopes []string, roles []string, lifetime time.Duration) (string, error) {
	if signingKey == nil {
		return "", errors.New("JWT signing key not configured")
	}
	now := time.Now()
	claims := TokenClaims{
		TenantID: tenantID,
		ClientID: clientID,
		Subject:  subject,
		Scopes:   scopes,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://sts.windows.net/" + tenantID + "/",
			Audience:  jwt.ClaimStrings{"https://graph.microsoft.com"},
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(lifetime)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signingKey)
}

// FilterKnownScopes filters out unknown scopes (Entra-like behavior: silently filter)
func FilterKnownScopes(scopes []string) []string {
	var filtered []string
	for _, scope := range scopes {
		if knownScopes[scope] {
			filtered = append(filtered, scope)
		}
	}
	return filtered
}

// GenerateRefreshToken generates a cryptographically random refresh token
func GenerateRefreshToken(tenantID, clientID, subject string, scopes []string, roles []string) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	tokenID := hex.EncodeToString(bytes)

	refreshTokensMutex.Lock()
	defer refreshTokensMutex.Unlock()

	// Store refresh token with 24h TTL
	refreshTokens[tokenID] = refreshTokenClaims{
		TenantID:  tenantID,
		ClientID:  clientID,
		Subject:   subject,
		Scopes:    scopes,
		Roles:     roles,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	return tokenID, nil
}

func ValidateToken(tokenString string) (*TokenClaims, error) {
	if signingKey == nil {
		return nil, errors.New("JWT signing key not configured")
	}
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return signingKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func TokenHandler(store store.Store) gin.HandlerFunc {
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
		tenant := c.Param("tenant")

		switch grantType {
		case "client_credentials":
			handleClientCredentials(c, store, tenant)
		case "authorization_code":
			handleAuthorizationCode(c, store, tenant)
		case "refresh_token":
			handleRefreshToken(c, store, tenant)
		default:
			writeTokenError(c, "unsupported_grant_type", fmt.Sprintf("Grant type '%s' not supported", grantType))
		}
	}
}

func handleClientCredentials(c *gin.Context, store store.Store, tenant string) {
	clientID := c.Request.FormValue("client_id")
	clientSecret := c.Request.FormValue("client_secret")
	scope := c.Request.FormValue("scope")

	if clientID == "" || clientSecret == "" {
		writeTokenError(c, "invalid_request", "client_id and client_secret are required")
		return
	}

	// Validate client credentials
	_, storedSecret, storedTenantID, err := store.GetClient(c.Request.Context(), clientID)
	if err != nil || storedSecret != clientSecret || storedTenantID != tenant {
		writeTokenError(c, "invalid_client", "Invalid client credentials")
		return
	}

	// Parse scopes
	scopes := []string{}
	if scope != "" {
		scopes = strings.Split(scope, " ")
	}

	// Default scopes if none provided
	if len(scopes) == 0 {
		scopes = []string{"User.Read", "User.Read.All"}
	}

	// Filter unknown scopes (Entra-like behavior: silently filter)
	scopes = FilterKnownScopes(scopes)

	// Mint token
	token, err := MintToken(tenant, clientID, clientID, scopes, []string{"Application"}, time.Hour)
	if err != nil {
		writeTokenError(c, "server_error", "Failed to mint token")
		return
	}

	writeTokenResponse(c, token)
}

func handleAuthorizationCode(c *gin.Context, store store.Store, tenant string) {
	code := c.Request.FormValue("code")
	clientID := c.Request.FormValue("client_id")
	scope := c.Request.FormValue("scope")

	if code == "" {
		writeTokenError(c, "invalid_request", "code is required")
		return
	}

	// Validate client exists
	_, _, clientTenantID, err := store.GetClient(c.Request.Context(), clientID)
	if err != nil || clientTenantID != tenant {
		writeTokenError(c, "invalid_client", "Invalid client credentials")
		return
	}

	// Parse scopes
	scopes := []string{}
	if scope != "" {
		scopes = strings.Split(scope, " ")
	}

	// Default scopes if none provided
	if len(scopes) == 0 {
		scopes = []string{"User.Read"}
	}

	// Filter unknown scopes (Entra-like behavior: silently filter)
	scopes = FilterKnownScopes(scopes)

	// Look up user by UPN first, then by ID
	user, err := store.GetUserByUPN(c.Request.Context(), code)
	if err != nil {
		// Try looking up by ID
		user, err = store.GetUser(c.Request.Context(), code)
		if err != nil {
			writeTokenError(c, "invalid_grant", "Invalid authorization code")
			return
		}
	}

	subject := user.UserPrincipalName
	if subject == "" {
		subject = user.ID
	}

	// Mint token with real user subject
	token, err := MintToken(tenant, clientID, subject, scopes, []string{"User"}, time.Hour)
	if err != nil {
		writeTokenError(c, "server_error", "Failed to mint token")
		return
	}

	// Generate refresh token
	refreshToken, err := GenerateRefreshToken(tenant, clientID, subject, scopes, []string{"User"})
	if err != nil {
		writeTokenError(c, "server_error", "Failed to generate refresh token")
		return
	}

	writeTokenResponse(c, token, refreshToken)
}

func handleRefreshToken(c *gin.Context, store store.Store, tenant string) {
	refreshToken := c.Request.FormValue("refresh_token")
	clientID := c.Request.FormValue("client_id")
	scope := c.Request.FormValue("scope")

	if refreshToken == "" {
		writeTokenError(c, "invalid_request", "refresh_token is required")
		return
	}

	// Validate client exists
	_, _, clientTenantID, err := store.GetClient(c.Request.Context(), clientID)
	if err != nil || clientTenantID != tenant {
		writeTokenError(c, "invalid_client", "Invalid client credentials")
		return
	}

	// Parse scopes
	scopes := []string{}
	if scope != "" {
		scopes = strings.Split(scope, " ")
	}

	// Filter unknown scopes (Entra-like behavior: silently filter)
	scopes = FilterKnownScopes(scopes)

	// Look up and validate refresh token
	refreshTokensMutex.Lock()
	defer refreshTokensMutex.Unlock()

	claims, exists := refreshTokens[refreshToken]
	if !exists {
		writeTokenError(c, "invalid_grant", "Invalid refresh token")
		return
	}

	// Validate token hasn't expired
	if time.Now().After(claims.ExpiresAt) {
		delete(refreshTokens, refreshToken)
		writeTokenError(c, "invalid_grant", "Refresh token has expired")
		return
	}

	// Validate client matches
	if claims.ClientID != clientID || claims.TenantID != tenant {
		writeTokenError(c, "invalid_grant", "Refresh token client mismatch")
		return
	}

	// Determine scopes: use provided scopes, or fall back to original token scopes
	finalScopes := scopes
	if len(finalScopes) == 0 {
		finalScopes = claims.Scopes
	}

	// Refresh token rotation: invalidate old refresh token
	delete(refreshTokens, refreshToken)

	// Mint new access token
	token, err := MintToken(claims.TenantID, claims.ClientID, claims.Subject, finalScopes, claims.Roles, time.Hour)
	if err != nil {
		writeTokenError(c, "server_error", "Failed to mint token")
		return
	}

	// Generate new refresh token
	newRefreshToken, err := GenerateRefreshToken(claims.TenantID, claims.ClientID, claims.Subject, finalScopes, claims.Roles)
	if err != nil {
		writeTokenError(c, "server_error", "Failed to generate refresh token")
		return
	}

	writeTokenResponse(c, token, newRefreshToken)
}

func writeTokenError(c *gin.Context, errorCode, errorDescription string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error":             errorCode,
		"error_description": errorDescription,
	})
}

func writeTokenResponse(c *gin.Context, accessToken string, refreshToken ...string) {
	resp := gin.H{
		"token_type":     "Bearer",
		"expires_in":     3600,
		"ext_expires_in": 3600,
		"access_token":   accessToken,
	}
	if len(refreshToken) > 0 && refreshToken[0] != "" {
		resp["refresh_token"] = refreshToken[0]
	}
	c.JSON(http.StatusOK, resp)
}

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			writeAuthError(c, "InvalidAuthenticationToken", "Access token is missing")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			writeAuthError(c, "InvalidAuthenticationToken", "Invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := ValidateToken(tokenString)
		if err != nil {
			writeAuthError(c, "InvalidAuthenticationToken", "Invalid access token")
			c.Abort()
			return
		}

		// Add claims to context
		c.Set("claims", claims)
		c.Next()
	}
}

func writeAuthError(c *gin.Context, code, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
			"innerError": gin.H{
				"date": time.Now().Format(time.RFC3339),
			},
		},
	})
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
				for id, claims := range refreshTokens {
					if now.After(claims.ExpiresAt) {
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
