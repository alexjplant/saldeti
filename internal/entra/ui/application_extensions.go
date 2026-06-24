package ui

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/saldeti/saldeti/internal/entra/model"
)

func ApplicationCreateExtensionHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		name := c.PostForm("name")
		dataType := c.PostForm("dataType")
		targetObjects := c.PostForm("targetObjects")

		if name == "" {
			h.handleExtensionsResponse(c, id, FlashDanger, "Extension property name is required")
			return
		}
		if !model.ValidExtensionDataTypes[dataType] {
			h.handleExtensionsResponse(c, id, FlashDanger, "Invalid dataType")
			return
		}

		var targets []string
		if targetObjects != "" {
			targets = strings.Split(targetObjects, ",")
			for i := range targets {
				targets[i] = strings.TrimSpace(targets[i])
			}
		}

		token, err := h.cred.GetToken(c.Request.Context(), policy.TokenRequestOptions{
			Scopes: []string{"https://graph.microsoft.com/.default"},
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to get token for creating extension property")
			h.handleExtensionsResponse(c, id, FlashDanger, "Failed to authenticate. Please try again.")
			return
		}

		payload := map[string]any{
			"name":          name,
			"dataType":      dataType,
			"targetObjects": targets,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal extension property payload")
			h.handleExtensionsResponse(c, id, FlashDanger, "Failed to prepare request. Please try again.")
			return
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), "POST",
			h.baseURL+"/v1.0/applications/"+id+"/extensionProperties",
			bytes.NewBuffer(payloadJSON))
		if err != nil {
			log.Error().Err(err).Msg("Failed to create extension property request")
			h.handleExtensionsResponse(c, id, FlashDanger, "Failed to create request. Please try again.")
			return
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create extension property")
			h.handleExtensionsResponse(c, id, FlashDanger, "Failed to create extension property. Please try again.")
			return
		}
		defer resp.Body.Close() //nolint:errcheck // deferred close error not actionable

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			log.Error().Int("status", resp.StatusCode).Str("body", string(body)).Msg("Failed to create extension property")
			h.handleExtensionsResponse(c, id, FlashDanger, "Failed to create extension property. Please try again.")
			return
		}

		h.handleExtensionsResponse(c, id, FlashSuccess, "Extension property created successfully")
	}
}

func ApplicationDeleteExtensionHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		extID := c.Param("extId")

		token, err := h.cred.GetToken(c.Request.Context(), policy.TokenRequestOptions{
			Scopes: []string{"https://graph.microsoft.com/.default"},
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to get token for deleting extension property")
			h.handleExtensionsResponse(c, id, FlashDanger, "Failed to authenticate. Please try again.")
			return
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), "DELETE",
			h.baseURL+"/v1.0/applications/"+id+"/extensionProperties/"+extID,
			nil)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create delete extension property request")
			h.handleExtensionsResponse(c, id, FlashDanger, "Failed to create request. Please try again.")
			return
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Error().Err(err).Msg("Failed to delete extension property")
			h.handleExtensionsResponse(c, id, FlashDanger, "Failed to delete extension property. Please try again.")
			return
		}
		defer resp.Body.Close() //nolint:errcheck // deferred close error not actionable

		if resp.StatusCode != http.StatusNoContent {
			body, _ := io.ReadAll(resp.Body)
			log.Error().Int("status", resp.StatusCode).Str("body", string(body)).Msg("Failed to delete extension property")
			h.handleExtensionsResponse(c, id, FlashDanger, "Failed to delete extension property. Please try again.")
			return
		}

		h.handleExtensionsResponse(c, id, FlashSuccess, "Extension property deleted successfully")
	}
}
