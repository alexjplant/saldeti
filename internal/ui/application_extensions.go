package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/model"
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
			h.handleExtensionsResponse(c, id, FlashDanger, fmt.Sprintf("Failed to get token: %v", err))
			return
		}

		payload := map[string]interface{}{
			"name":          name,
			"dataType":      dataType,
			"targetObjects": targets,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			h.handleExtensionsResponse(c, id, FlashDanger, fmt.Sprintf("Failed to marshal payload: %v", err))
			return
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), "POST",
			h.baseURL+"/v1.0/applications/"+id+"/extensionProperties",
			bytes.NewBuffer(payloadJSON))
		if err != nil {
			h.handleExtensionsResponse(c, id, FlashDanger, fmt.Sprintf("Failed to create request: %v", err))
			return
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			h.handleExtensionsResponse(c, id, FlashDanger, fmt.Sprintf("Failed to create extension property: %v", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			h.handleExtensionsResponse(c, id, FlashDanger, fmt.Sprintf("Failed to create extension property: %s", string(body)))
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
			h.handleExtensionsResponse(c, id, FlashDanger, fmt.Sprintf("Failed to get token: %v", err))
			return
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), "DELETE",
			h.baseURL+"/v1.0/applications/"+id+"/extensionProperties/"+extID,
			nil)
		if err != nil {
			h.handleExtensionsResponse(c, id, FlashDanger, fmt.Sprintf("Failed to create request: %v", err))
			return
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			h.handleExtensionsResponse(c, id, FlashDanger, fmt.Sprintf("Failed to delete extension property: %v", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			body, _ := io.ReadAll(resp.Body)
			h.handleExtensionsResponse(c, id, FlashDanger, fmt.Sprintf("Failed to delete extension property: %s", string(body)))
			return
		}

		h.handleExtensionsResponse(c, id, FlashSuccess, "Extension property deleted successfully")
	}
}
