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
			h.handleLicenseResponse(c, id, FlashDanger, "Failed to authenticate: "+err.Error())
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
			h.handleLicenseResponse(c, id, FlashDanger, "Failed to assign license: "+err.Error())
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			h.handleLicenseResponse(c, id, FlashDanger, fmt.Sprintf("Failed to assign license (%d): %s", resp.StatusCode, string(respBody)))
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
			h.handleLicenseResponse(c, id, FlashDanger, "Failed to authenticate: "+err.Error())
			return
		}

		payload := map[string]interface{}{
			"addLicenses":    []interface{}{},
			"removeLicenses": []map[string]interface{}{{"skuId": skuID}},
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
			h.handleLicenseResponse(c, id, FlashDanger, "Failed to remove license: "+err.Error())
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			h.handleLicenseResponse(c, id, FlashDanger, fmt.Sprintf("Failed to remove license (%d): %s", resp.StatusCode, string(respBody)))
			return
		}

		h.handleLicenseResponse(c, id, FlashSuccess, "License removed successfully")
	}
}
