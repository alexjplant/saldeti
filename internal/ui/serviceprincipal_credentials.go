package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/gin-gonic/gin"
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
			h.handleSPCredentialResponse(c, id, FlashDanger, fmt.Sprintf("Failed to get token: %v", err))
			return
		}

		payload := map[string]interface{}{
			"passwordCredential": map[string]interface{}{
				"displayName": displayName,
			},
		}
		payloadJSON, _ := json.Marshal(payload)

		req, _ := http.NewRequestWithContext(c.Request.Context(), "POST",
			h.baseURL+"/v1.0/servicePrincipals/"+id+"/addPassword",
			bytes.NewBuffer(payloadJSON))
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			h.handleSPCredentialResponse(c, id, FlashDanger, fmt.Sprintf("Failed to add password: %v", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			h.handleSPCredentialResponse(c, id, FlashDanger, fmt.Sprintf("Failed to add password: %s", string(body)))
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
			h.handleSPCredentialResponse(c, id, FlashDanger, fmt.Sprintf("Failed to get token: %v", err))
			return
		}

		payload := map[string]interface{}{
			"keyId": keyID,
		}
		payloadJSON, _ := json.Marshal(payload)

		req, _ := http.NewRequestWithContext(c.Request.Context(), "POST",
			h.baseURL+"/v1.0/servicePrincipals/"+id+"/removePassword",
			bytes.NewBuffer(payloadJSON))
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			h.handleSPCredentialResponse(c, id, FlashDanger, fmt.Sprintf("Failed to remove password: %v", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			h.handleSPCredentialResponse(c, id, FlashDanger, fmt.Sprintf("Failed to remove password: %s", string(body)))
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
			h.handleSPCredentialResponse(c, id, FlashDanger, fmt.Sprintf("Failed to get token: %v", err))
			return
		}

		certBase64, err := generateSelfSignedCert(displayName)
		if err != nil {
			h.handleSPCredentialResponse(c, id, FlashDanger, fmt.Sprintf("Failed to generate certificate: %v", err))
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
		payloadJSON, _ := json.Marshal(payload)

		req, _ := http.NewRequestWithContext(c.Request.Context(), "POST",
			h.baseURL+"/v1.0/servicePrincipals/"+id+"/addKey",
			bytes.NewBuffer(payloadJSON))
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			h.handleSPCredentialResponse(c, id, FlashDanger, fmt.Sprintf("Failed to add key: %v", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			h.handleSPCredentialResponse(c, id, FlashDanger, fmt.Sprintf("Failed to add key: %s", string(body)))
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
			h.handleSPCredentialResponse(c, id, FlashDanger, fmt.Sprintf("Failed to get token: %v", err))
			return
		}

		payload := map[string]interface{}{
			"keyId": keyID,
		}
		payloadJSON, _ := json.Marshal(payload)

		req, _ := http.NewRequestWithContext(c.Request.Context(), "POST",
			h.baseURL+"/v1.0/servicePrincipals/"+id+"/removeKey",
			bytes.NewBuffer(payloadJSON))
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			h.handleSPCredentialResponse(c, id, FlashDanger, fmt.Sprintf("Failed to remove key: %v", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			h.handleSPCredentialResponse(c, id, FlashDanger, fmt.Sprintf("Failed to remove key: %s", string(body)))
			return
		}

		h.handleSPCredentialResponse(c, id, FlashSuccess, "Key credential removed successfully")
	}
}
