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

// UserSetManagerHandler handles POST /ui/users/:id/manager/set
func UserSetManagerHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		managerID := c.PostForm("managerId")

		if managerID == "" {
			h.handleManagerResponse(c, id, FlashDanger, "No manager selected")
			return
		}

		// Build request body: {"@odata.id": "baseURL/v1.0/users/managerID"}
		body := map[string]string{
			"@odata.id": h.baseURL + "/v1.0/users/" + managerID,
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			h.handleManagerResponse(c, id, FlashDanger, "Failed to prepare request")
			return
		}

		// Get auth token
		token, err := h.cred.GetToken(c.Request.Context(), policy.TokenRequestOptions{
			Scopes: []string{"https://graph.microsoft.com/.default"},
		})
		if err != nil {
			h.handleManagerResponse(c, id, FlashDanger, fmt.Sprintf("Auth failed: %v", err))
			return
		}

		// PUT /v1.0/users/{id}/manager/$ref
		req, err := http.NewRequestWithContext(c.Request.Context(), "PUT", h.baseURL+"/v1.0/users/"+id+"/manager/$ref", bytes.NewReader(jsonBody))
		if err != nil {
			h.handleManagerResponse(c, id, FlashDanger, "Failed to create request")
			return
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			h.handleManagerResponse(c, id, FlashDanger, fmt.Sprintf("Failed to set manager: %v", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			respBody, _ := io.ReadAll(resp.Body)
			h.handleManagerResponse(c, id, FlashDanger, fmt.Sprintf("Failed to set manager (%d): %s", resp.StatusCode, string(respBody)))
			return
		}

		h.handleManagerResponse(c, id, FlashSuccess, "Manager set successfully")
	}
}

// UserRemoveManagerHandler handles POST /ui/users/:id/manager/remove
func UserRemoveManagerHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Get auth token
		token, err := h.cred.GetToken(c.Request.Context(), policy.TokenRequestOptions{
			Scopes: []string{"https://graph.microsoft.com/.default"},
		})
		if err != nil {
			h.handleManagerResponse(c, id, FlashDanger, fmt.Sprintf("Auth failed: %v", err))
			return
		}

		// DELETE /v1.0/users/{id}/manager/$ref
		req, err := http.NewRequestWithContext(c.Request.Context(), "DELETE", h.baseURL+"/v1.0/users/"+id+"/manager/$ref", nil)
		if err != nil {
			h.handleManagerResponse(c, id, FlashDanger, "Failed to create request")
			return
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			h.handleManagerResponse(c, id, FlashDanger, fmt.Sprintf("Failed to remove manager: %v", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			respBody, _ := io.ReadAll(resp.Body)
			h.handleManagerResponse(c, id, FlashDanger, fmt.Sprintf("Failed to remove manager (%d): %s", resp.StatusCode, string(respBody)))
			return
		}

		h.handleManagerResponse(c, id, FlashSuccess, "Manager removed successfully")
	}
}
