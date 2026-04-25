package ui

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func SPAddPasswordHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		displayName := c.PostForm("credentialDisplayName")
		if displayName == "" {
			displayName = "Generated Secret"
		}

		token, err := h.cred.GetToken(c.Request.Context(), policy.TokenRequestOptions{
			Scopes: []string{"https://graph.microsoft.com/.default"},
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to get token for adding SP password")
			h.handleSPCredentialResponse(c, id, FlashDanger, "Failed to authenticate. Please try again.")
			return
		}

		payload := map[string]interface{}{
			"passwordCredential": map[string]interface{}{
				"displayName": displayName,
			},
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal SP password payload")
			h.handleSPCredentialResponse(c, id, FlashDanger, "Failed to prepare request. Please try again.")
			return
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), "POST",
			h.baseURL+"/v1.0/servicePrincipals/"+id+"/addPassword",
			bytes.NewBuffer(payloadJSON))
		if err != nil {
			log.Error().Err(err).Msg("Failed to create SP password request")
			h.handleSPCredentialResponse(c, id, FlashDanger, "Failed to create request. Please try again.")
			return
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Error().Err(err).Msg("Failed to add SP password")
			h.handleSPCredentialResponse(c, id, FlashDanger, "Failed to add password. Please try again.")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Error().Int("status", resp.StatusCode).Str("body", string(body)).Msg("Failed to add SP password")
			h.handleSPCredentialResponse(c, id, FlashDanger, "Failed to add password. Please try again.")
			return
		}

		h.handleSPCredentialResponse(c, id, FlashSuccess, "Password credential added successfully")
	}
}

func SPRemovePasswordHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		keyID := c.Param("keyId")

		token, err := h.cred.GetToken(c.Request.Context(), policy.TokenRequestOptions{
			Scopes: []string{"https://graph.microsoft.com/.default"},
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to get token for removing SP password")
			h.handleSPCredentialResponse(c, id, FlashDanger, "Failed to authenticate. Please try again.")
			return
		}

		payload := map[string]interface{}{
			"keyId": keyID,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal remove SP password payload")
			h.handleSPCredentialResponse(c, id, FlashDanger, "Failed to prepare request. Please try again.")
			return
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), "POST",
			h.baseURL+"/v1.0/servicePrincipals/"+id+"/removePassword",
			bytes.NewBuffer(payloadJSON))
		if err != nil {
			log.Error().Err(err).Msg("Failed to create remove SP password request")
			h.handleSPCredentialResponse(c, id, FlashDanger, "Failed to create request. Please try again.")
			return
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Error().Err(err).Msg("Failed to remove SP password")
			h.handleSPCredentialResponse(c, id, FlashDanger, "Failed to remove password. Please try again.")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Error().Int("status", resp.StatusCode).Str("body", string(body)).Msg("Failed to remove SP password")
			h.handleSPCredentialResponse(c, id, FlashDanger, "Failed to remove password. Please try again.")
			return
		}

		h.handleSPCredentialResponse(c, id, FlashSuccess, "Password credential removed successfully")
	}
}

func SPAddKeyHandler(h *UIHandler) gin.HandlerFunc {
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

		token, err := h.cred.GetToken(c.Request.Context(), policy.TokenRequestOptions{
			Scopes: []string{"https://graph.microsoft.com/.default"},
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to get token for adding SP key")
			h.handleSPCredentialResponse(c, id, FlashDanger, "Failed to authenticate. Please try again.")
			return
		}

		certBase64, err := generateSelfSignedCert(displayName)
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate SP certificate")
			h.handleSPCredentialResponse(c, id, FlashDanger, "Failed to generate certificate. Please try again.")
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
			log.Error().Err(err).Msg("Failed to marshal SP key payload")
			h.handleSPCredentialResponse(c, id, FlashDanger, "Failed to prepare request. Please try again.")
			return
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), "POST",
			h.baseURL+"/v1.0/servicePrincipals/"+id+"/addKey",
			bytes.NewBuffer(payloadJSON))
		if err != nil {
			log.Error().Err(err).Msg("Failed to create SP key request")
			h.handleSPCredentialResponse(c, id, FlashDanger, "Failed to create request. Please try again.")
			return
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Error().Err(err).Msg("Failed to add SP key")
			h.handleSPCredentialResponse(c, id, FlashDanger, "Failed to add key. Please try again.")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			log.Error().Int("status", resp.StatusCode).Str("body", string(body)).Msg("Failed to add SP key")
			h.handleSPCredentialResponse(c, id, FlashDanger, "Failed to add key. Please try again.")
			return
		}

		h.handleSPCredentialResponse(c, id, FlashSuccess, "Key credential added successfully")
	}
}

func SPRemoveKeyHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		keyID := c.Param("keyId")

		token, err := h.cred.GetToken(c.Request.Context(), policy.TokenRequestOptions{
			Scopes: []string{"https://graph.microsoft.com/.default"},
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to get token for removing SP key")
			h.handleSPCredentialResponse(c, id, FlashDanger, "Failed to authenticate. Please try again.")
			return
		}

		payload := map[string]interface{}{
			"keyId": keyID,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal remove SP key payload")
			h.handleSPCredentialResponse(c, id, FlashDanger, "Failed to prepare request. Please try again.")
			return
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), "POST",
			h.baseURL+"/v1.0/servicePrincipals/"+id+"/removeKey",
			bytes.NewBuffer(payloadJSON))
		if err != nil {
			log.Error().Err(err).Msg("Failed to create remove SP key request")
			h.handleSPCredentialResponse(c, id, FlashDanger, "Failed to create request. Please try again.")
			return
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Error().Err(err).Msg("Failed to remove SP key")
			h.handleSPCredentialResponse(c, id, FlashDanger, "Failed to remove key. Please try again.")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Error().Int("status", resp.StatusCode).Str("body", string(body)).Msg("Failed to remove SP key")
			h.handleSPCredentialResponse(c, id, FlashDanger, "Failed to remove key. Please try again.")
			return
		}

		h.handleSPCredentialResponse(c, id, FlashSuccess, "Key credential removed successfully")
	}
}
