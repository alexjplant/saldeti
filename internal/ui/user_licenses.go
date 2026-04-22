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
		skuID := c.PostForm("skuId")
		if skuID == "" {
			SetFlash(c, FlashDanger, "No SKU selected")
			c.Redirect(http.StatusFound, "/ui/users/"+id)
			return
		}

		// Call POST /v1.0/users/{id}/assignLicense
		token, err := h.cred.GetToken(c.Request.Context(), policy.TokenRequestOptions{
			Scopes: []string{"https://graph.microsoft.com/.default"},
		})
		if err != nil {
			SetFlash(c, FlashDanger, "Failed to authenticate: "+err.Error())
			c.Redirect(http.StatusFound, "/ui/users/"+id)
			return
		}

		payload := map[string]interface{}{
			"addLicenses": []map[string]interface{}{
				{"skuId": skuID},
			},
			"removeLicenses": []interface{}{},
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequestWithContext(c.Request.Context(), "POST", h.baseURL+"/v1.0/users/"+id+"/assignLicense", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			SetFlash(c, FlashDanger, "Failed to assign license: "+err.Error())
			c.Redirect(http.StatusFound, "/ui/users/"+id)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			SetFlash(c, FlashDanger, fmt.Sprintf("Failed to assign license (%d): %s", resp.StatusCode, string(respBody)))
		} else {
			SetFlash(c, FlashSuccess, "License assigned successfully")
		}

		c.Redirect(http.StatusFound, "/ui/users/"+id)
	}
}

// UserRemoveLicenseHandler handles POST /ui/users/:id/licenses/:skuId/remove
func UserRemoveLicenseHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		skuID := c.Param("skuId")
		if skuID == "" {
			SetFlash(c, FlashDanger, "No SKU specified")
			c.Redirect(http.StatusFound, "/ui/users/"+id)
			return
		}

		token, err := h.cred.GetToken(c.Request.Context(), policy.TokenRequestOptions{
			Scopes: []string{"https://graph.microsoft.com/.default"},
		})
		if err != nil {
			SetFlash(c, FlashDanger, "Failed to authenticate: "+err.Error())
			c.Redirect(http.StatusFound, "/ui/users/"+id)
			return
		}

		payload := map[string]interface{}{
			"addLicenses":    []interface{}{},
			"removeLicenses": []map[string]interface{}{{"skuId": skuID}},
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequestWithContext(c.Request.Context(), "POST", h.baseURL+"/v1.0/users/"+id+"/assignLicense", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			SetFlash(c, FlashDanger, "Failed to remove license: "+err.Error())
			c.Redirect(http.StatusFound, "/ui/users/"+id)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			SetFlash(c, FlashDanger, fmt.Sprintf("Failed to remove license (%d): %s", resp.StatusCode, string(respBody)))
		} else {
			SetFlash(c, FlashSuccess, "License removed successfully")
		}

		c.Redirect(http.StatusFound, "/ui/users/"+id)
	}
}
