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

func ApplicationCreateHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			h.render(c, "templates/applications/form.html", gin.H{
				"ActiveNav":  "applications",
				"IsEdit":     false,
				"FormAction": "/ui/applications/new",
				"CancelURL":  "/ui/applications",
				"Form": map[string]interface{}{
					"DisplayName":    "",
					"Description":    "",
					"SignInAudience": "AzureADandPersonalMicrosoftAccount",
				},
			})
			return
		}

		// POST - handle form submission
		displayName := c.PostForm("displayName")

		// Validation
		if displayName == "" {
			h.render(c, "templates/applications/form.html", gin.H{
				"ActiveNav":  "applications",
				"IsEdit":     false,
				"FormAction": "/ui/applications/new",
				"CancelURL":  "/ui/applications",
				"Error":      "Display Name is required",
				"Form": map[string]interface{}{
					"DisplayName":    c.PostForm("displayName"),
					"Description":    c.PostForm("description"),
					"SignInAudience": c.PostForm("signInAudience"),
				},
			})
			return
		}

		// Create application via SDK
		newApp := models.NewApplication()
		newApp.SetDisplayName(&displayName)

		if description := c.PostForm("description"); description != "" {
			newApp.SetDescription(&description)
		}
		if signInAudience := c.PostForm("signInAudience"); signInAudience != "" {
			newApp.SetSignInAudience(&signInAudience)
		}

		// Try SDK Post first
		created, err := h.client.Applications().Post(c.Request.Context(), newApp, nil)

		// If SDK returns nil object without error, try manual HTTP request
		if created == nil && err == nil {
			token, tokenErr := h.cred.GetToken(c.Request.Context(), policy.TokenRequestOptions{})
			if tokenErr != nil {
				err = fmt.Errorf("failed to get token for manual request: %w", tokenErr)
			} else {
				appPayload := map[string]interface{}{
					"displayName":    displayName,
					"@odata.type":    "#microsoft.graph.application",
				}
				if description := c.PostForm("description"); description != "" {
					appPayload["description"] = description
				}
				if signInAudience := c.PostForm("signInAudience"); signInAudience != "" {
					appPayload["signInAudience"] = signInAudience
				}

				appJSON, marshalErr := json.Marshal(appPayload)
				if marshalErr != nil {
					err = fmt.Errorf("failed to marshal application: %w", marshalErr)
				} else {
					req, reqErr := http.NewRequestWithContext(c.Request.Context(), "POST", h.client.GetAdapter().GetBaseUrl()+"/applications", bytes.NewBuffer(appJSON))
					if reqErr != nil {
						err = fmt.Errorf("failed to create request: %w", reqErr)
					} else {
						req.Header.Set("Authorization", "Bearer "+token.Token)
						req.Header.Set("Content-Type", "application/json")

						resp, httpErr := httpClient.Do(req)
						if httpErr != nil {
							err = fmt.Errorf("HTTP request failed: %w", httpErr)
						} else {
							defer resp.Body.Close()

							if resp.StatusCode != http.StatusCreated {
								body, _ := io.ReadAll(resp.Body)
								err = fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
							} else {
								var result map[string]interface{}
								if parseErr := json.NewDecoder(resp.Body).Decode(&result); parseErr != nil {
									err = fmt.Errorf("failed to decode response: %w", parseErr)
								} else {
									if id, ok := result["id"].(string); ok && id != "" {
										manualApp := models.NewApplication()
										manualApp.SetId(&id)
										created = manualApp
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
			h.render(c, "templates/applications/form.html", gin.H{
				"ActiveNav":  "applications",
				"IsEdit":     false,
				"FormAction": "/ui/applications/new",
				"CancelURL":  "/ui/applications",
				"Error":      fmt.Sprintf("Failed to create application: %v", err),
				"Form": map[string]interface{}{
					"DisplayName":    c.PostForm("displayName"),
					"Description":    c.PostForm("description"),
					"SignInAudience": c.PostForm("signInAudience"),
				},
			})
			return
		}

		if created == nil {
			h.render(c, "templates/applications/form.html", gin.H{
				"ActiveNav":  "applications",
				"IsEdit":     false,
				"FormAction": "/ui/applications/new",
				"CancelURL":  "/ui/applications",
				"Error":      "Application was created but response was empty",
				"Form": map[string]interface{}{
					"DisplayName":    c.PostForm("displayName"),
					"Description":    c.PostForm("description"),
					"SignInAudience": c.PostForm("signInAudience"),
				},
			})
			return
		}

		// Get application ID
		var appObjID string
		if id := created.GetId(); id != nil {
			appObjID = *id
		} else if additionalData := created.GetAdditionalData(); additionalData != nil {
			if id, ok := additionalData["id"].(string); ok {
				appObjID = id
			}
		}

		if appObjID == "" {
			h.render(c, "templates/applications/form.html", gin.H{
				"ActiveNav":  "applications",
				"IsEdit":     false,
				"FormAction": "/ui/applications/new",
				"CancelURL":  "/ui/applications",
				"Error":      "Application was created but ID was not returned in response",
				"Form": map[string]interface{}{
					"DisplayName":    c.PostForm("displayName"),
					"Description":    c.PostForm("description"),
					"SignInAudience": c.PostForm("signInAudience"),
				},
			})
			return
		}

		// Success - redirect to application detail page
		SetFlash(c, FlashSuccess, "Application created successfully")
		c.Redirect(http.StatusFound, "/ui/applications/"+appObjID)
	}
}
