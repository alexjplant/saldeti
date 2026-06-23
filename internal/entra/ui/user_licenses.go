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

// UserAddLicenseHandler handles POST /ui/users/:id/licenses/add
func UserAddLicenseHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ctx := c.Request.Context()
		skuID := c.PostForm("skuId")

		if skuID == "" {
			h.handleLicenseResponse(c, id, FlashDanger, "No SKU selected")
			return
		}

		// Call POST /v1.0/users/{id}/assignLicense
		token, err := h.cred.GetToken(ctx, policy.TokenRequestOptions{
			Scopes: []string{"https://graph.microsoft.com/.default"},
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to authenticate for license assignment")
			h.handleLicenseResponse(c, id, FlashDanger, "Failed to authenticate. Please try again.")
			return
		}

		payload := map[string]interface{}{
			"addLicenses": []map[string]interface{}{
				{"skuId": skuID},
			},
			"removeLicenses": []interface{}{},
		}
		body, err := json.Marshal(payload)
		if err != nil {
			h.handleLicenseResponse(c, id, FlashDanger, "Failed to prepare request")
			return
		}

		req, err := http.NewRequestWithContext(ctx, "POST", h.baseURL+"/v1.0/users/"+id+"/assignLicense", bytes.NewBuffer(body))
		if err != nil {
			h.handleLicenseResponse(c, id, FlashDanger, "Failed to create request")
			return
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Error().Err(err).Msg("Failed to assign license")
			h.handleLicenseResponse(c, id, FlashDanger, "Failed to assign license. Please try again.")
			return
		}
		defer resp.Body.Close() //nolint:errcheck // deferred close error not actionable

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			log.Error().Int("status", resp.StatusCode).Str("body", string(respBody)).Msg("Failed to assign license")
			h.handleLicenseResponse(c, id, FlashDanger, "Failed to assign license. Please try again.")
			return
		}

		h.handleLicenseResponse(c, id, FlashSuccess, "License assigned successfully")
	}
}

// UserRemoveLicenseHandler handles POST /ui/users/:id/licenses/:skuId/remove
func UserRemoveLicenseHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ctx := c.Request.Context()
		skuID := c.Param("skuId")

		if skuID == "" {
			h.handleLicenseResponse(c, id, FlashDanger, "No SKU specified")
			return
		}

		token, err := h.cred.GetToken(ctx, policy.TokenRequestOptions{
			Scopes: []string{"https://graph.microsoft.com/.default"},
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to authenticate for license removal")
			h.handleLicenseResponse(c, id, FlashDanger, "Failed to authenticate. Please try again.")
			return
		}

		payload := map[string]interface{}{
			"addLicenses":    []interface{}{},
			"removeLicenses": []string{skuID},
		}
		body, err := json.Marshal(payload)
		if err != nil {
			h.handleLicenseResponse(c, id, FlashDanger, "Failed to prepare request")
			return
		}

		req, err := http.NewRequestWithContext(ctx, "POST", h.baseURL+"/v1.0/users/"+id+"/assignLicense", bytes.NewBuffer(body))
		if err != nil {
			h.handleLicenseResponse(c, id, FlashDanger, "Failed to create request")
			return
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Error().Err(err).Msg("Failed to remove license")
			h.handleLicenseResponse(c, id, FlashDanger, "Failed to remove license. Please try again.")
			return
		}
		defer resp.Body.Close() //nolint:errcheck // deferred close error not actionable

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			log.Error().Int("status", resp.StatusCode).Str("body", string(respBody)).Msg("Failed to remove license")
			h.handleLicenseResponse(c, id, FlashDanger, "Failed to remove license. Please try again.")
			return
		}

		h.handleLicenseResponse(c, id, FlashSuccess, "License removed successfully")
	}
}
