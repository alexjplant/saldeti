package ui

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func ApplicationAddPasswordHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		displayName := c.PostForm("credentialDisplayName")
		if displayName == "" {
			displayName = "Generated Secret"
		}

		// Use raw HTTP to add password credential
		token, err := h.cred.GetToken(c.Request.Context(), policy.TokenRequestOptions{
			Scopes: []string{"https://graph.microsoft.com/.default"},
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to get token for adding password")
			h.handleCredentialResponse(c, id, FlashDanger, "Failed to authenticate. Please try again.")
			return
		}

		payload := map[string]interface{}{
			"passwordCredential": map[string]interface{}{
				"displayName": displayName,
			},
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal password payload")
			h.handleCredentialResponse(c, id, FlashDanger, "Failed to prepare request. Please try again.")
			return
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), "POST",
			h.baseURL+"/v1.0/applications/"+id+"/addPassword",
			bytes.NewBuffer(payloadJSON))
		if err != nil {
			log.Error().Err(err).Msg("Failed to create password request")
			h.handleCredentialResponse(c, id, FlashDanger, "Failed to create request. Please try again.")
			return
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Error().Err(err).Msg("Failed to add password")
			h.handleCredentialResponse(c, id, FlashDanger, "Failed to add password. Please try again.")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Error().Int("status", resp.StatusCode).Str("body", string(body)).Msg("Failed to add password")
			h.handleCredentialResponse(c, id, FlashDanger, "Failed to add password. Please try again.")
			return
		}

		h.handleCredentialResponse(c, id, FlashSuccess, "Password credential added successfully")
	}
}

func ApplicationRemovePasswordHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		keyID := c.Param("keyId")

		// Use raw HTTP to remove password credential
		token, err := h.cred.GetToken(c.Request.Context(), policy.TokenRequestOptions{
			Scopes: []string{"https://graph.microsoft.com/.default"},
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to get token for removing password")
			h.handleCredentialResponse(c, id, FlashDanger, "Failed to authenticate. Please try again.")
			return
		}

		payload := map[string]interface{}{
			"keyId": keyID,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal remove password payload")
			h.handleCredentialResponse(c, id, FlashDanger, "Failed to prepare request. Please try again.")
			return
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), "POST",
			h.baseURL+"/v1.0/applications/"+id+"/removePassword",
			bytes.NewBuffer(payloadJSON))
		if err != nil {
			log.Error().Err(err).Msg("Failed to create remove password request")
			h.handleCredentialResponse(c, id, FlashDanger, "Failed to create request. Please try again.")
			return
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Error().Err(err).Msg("Failed to remove password")
			h.handleCredentialResponse(c, id, FlashDanger, "Failed to remove password. Please try again.")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Error().Int("status", resp.StatusCode).Str("body", string(body)).Msg("Failed to remove password")
			h.handleCredentialResponse(c, id, FlashDanger, "Failed to remove password. Please try again.")
			return
		}

		h.handleCredentialResponse(c, id, FlashSuccess, "Password credential removed successfully")
	}
}

func ApplicationAddKeyHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		displayName := c.PostForm("keyDisplayName")
		keyType := c.PostForm("keyType")
		keyUsage := c.PostForm("keyUsage")

		if displayName == "" {
			displayName = "Generated Key"
		}
		if keyType == "" {
			keyType = "AsymmetricX509Cert"
		}
		if keyUsage == "" {
			keyUsage = "Verify"
		}

		// Use raw HTTP to add key credential
		token, err := h.cred.GetToken(c.Request.Context(), policy.TokenRequestOptions{
			Scopes: []string{"https://graph.microsoft.com/.default"},
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to get token for adding key")
			h.handleCredentialResponse(c, id, FlashDanger, "Failed to authenticate. Please try again.")
			return
		}

		// Generate a self-signed certificate for key
		certBase64, err := generateSelfSignedCert(displayName)
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate certificate")
			h.handleCredentialResponse(c, id, FlashDanger, "Failed to generate certificate. Please try again.")
			return
		}

		payload := map[string]interface{}{
			"keyCredential": map[string]interface{}{
				"displayName": displayName,
				"type":        keyType,
				"usage":       keyUsage,
				"key":         certBase64,
			},
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal key payload")
			h.handleCredentialResponse(c, id, FlashDanger, "Failed to prepare request. Please try again.")
			return
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), "POST",
			h.baseURL+"/v1.0/applications/"+id+"/addKey",
			bytes.NewBuffer(payloadJSON))
		if err != nil {
			log.Error().Err(err).Msg("Failed to create key request")
			h.handleCredentialResponse(c, id, FlashDanger, "Failed to create request. Please try again.")
			return
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Error().Err(err).Msg("Failed to add key")
			h.handleCredentialResponse(c, id, FlashDanger, "Failed to add key. Please try again.")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			log.Error().Int("status", resp.StatusCode).Str("body", string(body)).Msg("Failed to add key")
			h.handleCredentialResponse(c, id, FlashDanger, "Failed to add key. Please try again.")
			return
		}

		h.handleCredentialResponse(c, id, FlashSuccess, "Key credential added successfully")
	}
}

func ApplicationRemoveKeyHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		keyID := c.Param("keyId")

		// Use raw HTTP to remove key credential
		token, err := h.cred.GetToken(c.Request.Context(), policy.TokenRequestOptions{
			Scopes: []string{"https://graph.microsoft.com/.default"},
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to get token for removing key")
			h.handleCredentialResponse(c, id, FlashDanger, "Failed to authenticate. Please try again.")
			return
		}

		payload := map[string]interface{}{
			"keyId": keyID,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal remove key payload")
			h.handleCredentialResponse(c, id, FlashDanger, "Failed to prepare request. Please try again.")
			return
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), "POST",
			h.baseURL+"/v1.0/applications/"+id+"/removeKey",
			bytes.NewBuffer(payloadJSON))
		if err != nil {
			log.Error().Err(err).Msg("Failed to create remove key request")
			h.handleCredentialResponse(c, id, FlashDanger, "Failed to create request. Please try again.")
			return
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Error().Err(err).Msg("Failed to remove key")
			h.handleCredentialResponse(c, id, FlashDanger, "Failed to remove key. Please try again.")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Error().Int("status", resp.StatusCode).Str("body", string(body)).Msg("Failed to remove key")
			h.handleCredentialResponse(c, id, FlashDanger, "Failed to remove key. Please try again.")
			return
		}

		h.handleCredentialResponse(c, id, FlashSuccess, "Key credential removed successfully")
	}
}

// generateSelfSignedCert generates a minimal base64-encoded self-signed certificate
// This is used for the addKey endpoint which requires a key property
func generateSelfSignedCert(displayName string) (string, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", err
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: displayName},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(certDER), nil
}
