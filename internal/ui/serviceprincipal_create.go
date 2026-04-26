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
	"github.com/saldeti/saldeti/internal/model"
)

func SPCreateHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			// Fetch applications for dropdown
			appResult, _ := h.client.Applications().Get(c.Request.Context(), nil)
			var appRows []model.Application
			if appResult != nil {
				for _, a := range appResult.GetValue() {
					appRows = append(appRows, sdkApplicationToModel(a))
				}
			}

			h.render(c, "templates/serviceprincipals/form.html", gin.H{
				"ActiveNav":    "serviceprincipals",
				"IsEdit":       false,
				"FormAction":   "/ui/servicePrincipals/new",
				"CancelURL":    "/ui/servicePrincipals",
				"Applications": appRows,
				"Form": map[string]interface{}{
					"DisplayName": "",
					"AppId":       "",
					"Notes":       "",
				},
			})
			return
		}

		// POST - handle form submission
		displayName := c.PostForm("displayName")

		// Fetch applications list for dropdown (used in error re-renders)
		appResult, _ := h.client.Applications().Get(c.Request.Context(), nil)
		var appRows []model.Application
		if appResult != nil {
			for _, a := range appResult.GetValue() {
				appRows = append(appRows, sdkApplicationToModel(a))
			}
		}

		// Validation
		if displayName == "" {
			h.render(c, "templates/serviceprincipals/form.html", gin.H{
				"ActiveNav":    "serviceprincipals",
				"IsEdit":       false,
				"FormAction":   "/ui/servicePrincipals/new",
				"CancelURL":    "/ui/servicePrincipals",
				"Error":        "Display Name is required",
				"Applications": appRows,
				"Form": map[string]interface{}{
					"DisplayName": c.PostForm("displayName"),
					"AppId":       c.PostForm("appId"),
					"Notes":       c.PostForm("notes"),
				},
			})
			return
		}

		// Validate appId is provided from dropdown
		appId := c.PostForm("appId")
		if appId == "" {
			h.render(c, "templates/serviceprincipals/form.html", gin.H{
				"ActiveNav":    "serviceprincipals",
				"IsEdit":       false,
				"FormAction":   "/ui/servicePrincipals/new",
				"CancelURL":    "/ui/servicePrincipals",
				"Error":        "Application is required",
				"Applications": appRows,
				"Form": map[string]interface{}{
					"DisplayName": c.PostForm("displayName"),
					"AppId":       "",
					"Notes":       c.PostForm("notes"),
				},
			})
			return
		}

		// Create service principal via SDK
		newSP := models.NewServicePrincipal()
		newSP.SetDisplayName(&displayName)
		newSP.SetAppId(&appId)

		if notes := c.PostForm("notes"); notes != "" {
			newSP.SetDescription(&notes)
		}

		// Try SDK Post first
		created, err := h.client.ServicePrincipals().Post(c.Request.Context(), newSP, nil)

		// If SDK returns nil object without error, try manual HTTP request
		if created == nil && err == nil {
			token, tokenErr := h.cred.GetToken(c.Request.Context(), policy.TokenRequestOptions{})
			if tokenErr != nil {
				err = fmt.Errorf("failed to get token for manual request: %w", tokenErr)
			} else {
				spPayload := map[string]interface{}{
					"displayName": displayName,
					"@odata.type": "#microsoft.graph.servicePrincipal",
				}
				spPayload["appId"] = appId
				if notes := c.PostForm("notes"); notes != "" {
					spPayload["description"] = notes
				}

				spJSON, marshalErr := json.Marshal(spPayload)
				if marshalErr != nil {
					err = fmt.Errorf("failed to marshal service principal: %w", marshalErr)
				} else {
					req, reqErr := http.NewRequestWithContext(c.Request.Context(), "POST", h.baseURL+"/v1.0/servicePrincipals", bytes.NewBuffer(spJSON))
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
										manualSP := models.NewServicePrincipal()
										manualSP.SetId(&id)
										created = manualSP
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
			h.render(c, "templates/serviceprincipals/form.html", gin.H{
				"ActiveNav":    "serviceprincipals",
				"IsEdit":       false,
				"FormAction":   "/ui/servicePrincipals/new",
				"CancelURL":    "/ui/servicePrincipals",
				"Error":        fmt.Sprintf("Failed to create service principal: %v", err),
				"Applications": appRows,
				"Form": map[string]interface{}{
					"DisplayName": c.PostForm("displayName"),
					"AppId":       c.PostForm("appId"),
					"Notes":       c.PostForm("notes"),
				},
			})
			return
		}

		if created == nil {
			h.render(c, "templates/serviceprincipals/form.html", gin.H{
				"ActiveNav":    "serviceprincipals",
				"IsEdit":       false,
				"FormAction":   "/ui/servicePrincipals/new",
				"CancelURL":    "/ui/servicePrincipals",
				"Error":        "Service principal was created but response was empty",
				"Applications": appRows,
				"Form": map[string]interface{}{
					"DisplayName": c.PostForm("displayName"),
					"AppId":       c.PostForm("appId"),
					"Notes":       c.PostForm("notes"),
				},
			})
			return
		}

		// Get service principal ID
		var spObjID string
		if id := created.GetId(); id != nil {
			spObjID = *id
		} else if additionalData := created.GetAdditionalData(); additionalData != nil {
			if id, ok := additionalData["id"].(string); ok {
				spObjID = id
			}
		}

		if spObjID == "" {
			h.render(c, "templates/serviceprincipals/form.html", gin.H{
				"ActiveNav":    "serviceprincipals",
				"IsEdit":       false,
				"FormAction":   "/ui/servicePrincipals/new",
				"CancelURL":    "/ui/servicePrincipals",
				"Error":        "Service principal was created but ID was not returned in response",
				"Applications": appRows,
				"Form": map[string]interface{}{
					"DisplayName": c.PostForm("displayName"),
					"AppId":       c.PostForm("appId"),
					"Notes":       c.PostForm("notes"),
				},
			})
			return
		}

		// Success - redirect to service principal detail page
		SetFlash(c, FlashSuccess, "Service principal created successfully")
		c.Redirect(http.StatusFound, "/ui/servicePrincipals/"+spObjID)
	}
}
