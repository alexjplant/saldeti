package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/gin-gonic/gin"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func GroupCreateHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			// Render empty form with defaults
			h.render(c, "templates/groups/form.html", gin.H{
				"ActiveNav":  "groups",
				"IsEdit":     false,
				"FormAction": "/ui/groups/new",
				"CancelURL":  "/ui/groups",
				"Form": map[string]interface{}{
					"DisplayName":      "",
					"Description":      "",
					"MailNickname":     "",
					"SecurityEnabled":  "true",
					"MailEnabled":      "false",
					"Unified":          "false",
					"Visibility":       "Public",
				},
			})
			return
		}

		// POST - handle form submission
		displayName := c.PostForm("displayName")

		// Validation
		if displayName == "" {
			h.render(c, "templates/groups/form.html", gin.H{
				"ActiveNav":  "groups",
				"IsEdit":     false,
				"FormAction": "/ui/groups/new",
				"CancelURL":  "/ui/groups",
				"Error":      "Display Name is required",
				"Form": map[string]interface{}{
					"DisplayName":     c.PostForm("displayName"),
					"Description":     c.PostForm("description"),
					"MailNickname":    c.PostForm("mailNickname"),
					"SecurityEnabled": c.PostForm("securityEnabled"),
					"MailEnabled":     c.PostForm("mailEnabled"),
					"Unified":         c.PostForm("unified"),
					"Visibility":      c.PostForm("visibility"),
				},
			})
			return
		}

		// Handle securityEnabled checkbox
		securityEnabled := c.PostForm("securityEnabled") == "true"

		// Handle mailEnabled checkbox
		mailEnabled := c.PostForm("mailEnabled") == "true"

		// Handle unified checkbox (GroupTypes)
		var groupTypes []string
		if c.PostForm("unified") == "true" {
			groupTypes = []string{"Unified"}
		}

		// Create group via SDK
		newGroup := models.NewGroup()
		newGroup.SetDisplayName(&displayName)
		newGroup.SetSecurityEnabled(&securityEnabled)
		newGroup.SetMailEnabled(&mailEnabled)

		if mailNickname := c.PostForm("mailNickname"); mailNickname != "" {
			newGroup.SetMailNickname(&mailNickname)
		}
		if description := c.PostForm("description"); description != "" {
			newGroup.SetDescription(&description)
		}
		if len(groupTypes) > 0 {
			newGroup.SetGroupTypes(groupTypes)
		}
		if visibility := c.PostForm("visibility"); visibility != "" {
			newGroup.SetVisibility(&visibility)
		}

		// Try SDK Post first
		created, err := h.client.Groups().Post(c.Request.Context(), newGroup, nil)
		
		// If SDK returns nil object without error, try manual HTTP request
		if created == nil && err == nil {
			// Manually make the HTTP request
			token, tokenErr := h.cred.GetToken(c.Request.Context(), policy.TokenRequestOptions{})
			if tokenErr != nil {
				err = fmt.Errorf("failed to get token for manual request: %w", tokenErr)
			} else {
				// Manually construct the JSON payload
				groupPayload := map[string]interface{}{
					"displayName":      displayName,
					"securityEnabled":  securityEnabled,
					"mailEnabled":      mailEnabled,
					"@odata.type":      "#microsoft.graph.group",
				}
				if mailNickname := c.PostForm("mailNickname"); mailNickname != "" {
					groupPayload["mailNickname"] = mailNickname
				}
				if description := c.PostForm("description"); description != "" {
					groupPayload["description"] = description
				}
				if len(groupTypes) > 0 {
					groupPayload["groupTypes"] = groupTypes
				}
				if visibility := c.PostForm("visibility"); visibility != "" {
					groupPayload["visibility"] = visibility
				}
				
				// Serialize to JSON
				groupJSON, marshalErr := json.Marshal(groupPayload)
				if marshalErr != nil {
					err = fmt.Errorf("failed to marshal group: %w", marshalErr)
				} else {
					// Create HTTP request
					req, reqErr := http.NewRequestWithContext(c.Request.Context(), "POST", h.client.GetAdapter().GetBaseUrl()+"/groups", bytes.NewBuffer(groupJSON))
					if reqErr != nil {
						err = fmt.Errorf("failed to create request: %w", reqErr)
					} else {
						req.Header.Set("Authorization", "Bearer "+token.Token)
						req.Header.Set("Content-Type", "application/json")
						
						// Make request
						resp, httpErr := httpClient.Do(req)
						if httpErr != nil {
							err = fmt.Errorf("HTTP request failed: %w", httpErr)
						} else {
							defer resp.Body.Close()
							
							if resp.StatusCode != http.StatusCreated {
								body, _ := io.ReadAll(resp.Body)
								err = fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
							} else {
								// Parse response to get ID
								var result map[string]interface{}
								if parseErr := json.NewDecoder(resp.Body).Decode(&result); parseErr != nil {
									err = fmt.Errorf("failed to decode response: %w", parseErr)
								} else {
									// Create a simple SDK-like object to return
									// We'll extract the ID and use it
									if id, ok := result["id"].(string); ok && id != "" {
										// Create a new group object with the ID
										manualGroup := models.NewGroup()
										manualGroup.SetId(&id)
										created = manualGroup
									} else {
										err = fmt.Errorf("response did not contain an ID")
									}
								}
							}
						}
					}
				}
			}
		}
		if err != nil {
			h.render(c, "templates/groups/form.html", gin.H{
				"ActiveNav":  "groups",
				"IsEdit":     false,
				"FormAction": "/ui/groups/new",
				"CancelURL":  "/ui/groups",
				"Error":      fmt.Sprintf("Failed to create group: %v", err),
				"Form": map[string]interface{}{
					"DisplayName":     c.PostForm("displayName"),
					"Description":     c.PostForm("description"),
					"MailNickname":    c.PostForm("mailNickname"),
					"SecurityEnabled": c.PostForm("securityEnabled"),
					"MailEnabled":     c.PostForm("mailEnabled"),
					"Unified":         c.PostForm("unified"),
					"Visibility":      c.PostForm("visibility"),
				},
			})
			return
		}

		// Check if created object is nil
		if created == nil {
			h.render(c, "templates/groups/form.html", gin.H{
				"ActiveNav":  "groups",
				"IsEdit":     false,
				"FormAction": "/ui/groups/new",
				"CancelURL":  "/ui/groups",
				"Error":      "Group was created but response was empty",
				"Form": map[string]interface{}{
					"DisplayName":     c.PostForm("displayName"),
					"Description":     c.PostForm("description"),
					"MailNickname":    c.PostForm("mailNickname"),
					"SecurityEnabled": c.PostForm("securityEnabled"),
					"MailEnabled":     c.PostForm("mailEnabled"),
					"Unified":         c.PostForm("unified"),
					"Visibility":      c.PostForm("visibility"),
				},
			})
			return
		}

		// Get group ID - try GetId() first, then fall back to additional data
		var groupID string
		if id := created.GetId(); id != nil {
			groupID = *id
		} else if additionalData := created.GetAdditionalData(); additionalData != nil {
			if id, ok := additionalData["id"].(string); ok {
				groupID = id
			}
		}

		if groupID == "" {
			h.render(c, "templates/groups/form.html", gin.H{
				"ActiveNav":  "groups",
				"IsEdit":     false,
				"FormAction": "/ui/groups/new",
				"CancelURL":  "/ui/groups",
				"Error":      "Group was created but ID was not returned in response",
				"Form": map[string]interface{}{
					"DisplayName":     c.PostForm("displayName"),
					"Description":     c.PostForm("description"),
					"MailNickname":    c.PostForm("mailNickname"),
					"SecurityEnabled": c.PostForm("securityEnabled"),
					"MailEnabled":     c.PostForm("mailEnabled"),
					"Unified":         c.PostForm("unified"),
					"Visibility":      c.PostForm("visibility"),
				},
			})
			return
		}

		// Success - redirect to group detail page
		SetFlash(c, FlashSuccess, "Group created successfully")
		c.Redirect(http.StatusFound, "/ui/groups/"+groupID)
	}
}
