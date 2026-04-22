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

func UserCreateHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			h.render(c, "templates/users/form.html", gin.H{
				"ActiveNav":  "users",
				"IsEdit":     false,
				"FormAction": "/ui/users/new",
				"CancelURL":  "/ui/users",
				"Form": map[string]interface{}{
					"DisplayName":        "",
					"UserPrincipalName":  "",
					"GivenName":          "",
					"Surname":            "",
					"Mail":               "",
					"MailNickname":       "",
					"JobTitle":           "",
					"Department":         "",
					"OfficeLocation":     "",
					"MobilePhone":        "",
					"AccountEnabled":     true,
				},
			})
			return
		}

		displayName := c.PostForm("displayName")
		upn := c.PostForm("userPrincipalName")
		if displayName == "" || upn == "" {
			h.render(c, "templates/users/form.html", gin.H{
				"ActiveNav":  "users",
				"IsEdit":     false,
				"FormAction": "/ui/users/new",
				"CancelURL":  "/ui/users",
				"Error":      "Display Name and User Principal Name are required",
				"Form": map[string]interface{}{
					"DisplayName":       c.PostForm("displayName"),
					"UserPrincipalName": c.PostForm("userPrincipalName"),
					"GivenName":         c.PostForm("givenName"),
					"Surname":           c.PostForm("surname"),
					"Mail":              c.PostForm("mail"),
					"MailNickname":      c.PostForm("mailNickname"),
					"JobTitle":          c.PostForm("jobTitle"),
					"Department":        c.PostForm("department"),
					"OfficeLocation":    c.PostForm("officeLocation"),
					"MobilePhone":       c.PostForm("mobilePhone"),
					"AccountEnabled":    c.PostForm("accountEnabled") == "true",
				},
			})
			return
		}

		accountEnabled := c.PostForm("accountEnabled") == "true"

		// Create user via SDK
		newUser := models.NewUser()
		newUser.SetDisplayName(&displayName)
		newUser.SetUserPrincipalName(&upn)
		newUser.SetAccountEnabled(&accountEnabled)

		if givenName := c.PostForm("givenName"); givenName != "" {
			newUser.SetGivenName(&givenName)
		}
		if surname := c.PostForm("surname"); surname != "" {
			newUser.SetSurname(&surname)
		}
		if mail := c.PostForm("mail"); mail != "" {
			newUser.SetMail(&mail)
		}
		if mailNickname := c.PostForm("mailNickname"); mailNickname != "" {
			newUser.SetMailNickname(&mailNickname)
		}
		if jobTitle := c.PostForm("jobTitle"); jobTitle != "" {
			newUser.SetJobTitle(&jobTitle)
		}
		if department := c.PostForm("department"); department != "" {
			newUser.SetDepartment(&department)
		}
		if officeLocation := c.PostForm("officeLocation"); officeLocation != "" {
			newUser.SetOfficeLocation(&officeLocation)
		}
		if mobilePhone := c.PostForm("mobilePhone"); mobilePhone != "" {
			newUser.SetMobilePhone(&mobilePhone)
		}

		// Try SDK Post first
		created, err := h.client.Users().Post(c.Request.Context(), newUser, nil)
		
		// If SDK returns nil object without error, try manual HTTP request
		if created == nil && err == nil {
			// Manually make the HTTP request
			token, tokenErr := h.cred.GetToken(c.Request.Context(), policy.TokenRequestOptions{})
			if tokenErr != nil {
				err = fmt.Errorf("failed to get token for manual request: %w", tokenErr)
			} else {
				// Manually construct the JSON payload
				userPayload := map[string]interface{}{
					"displayName":       displayName,
					"userPrincipalName": upn,
					"accountEnabled":   accountEnabled,
					"@odata.type":      "#microsoft.graph.user",
				}
				if givenName := c.PostForm("givenName"); givenName != "" {
					userPayload["givenName"] = givenName
				}
				if surname := c.PostForm("surname"); surname != "" {
					userPayload["surname"] = surname
				}
				if mail := c.PostForm("mail"); mail != "" {
					userPayload["mail"] = mail
				}
				if mailNickname := c.PostForm("mailNickname"); mailNickname != "" {
					userPayload["mailNickname"] = mailNickname
				}
				if jobTitle := c.PostForm("jobTitle"); jobTitle != "" {
					userPayload["jobTitle"] = jobTitle
				}
				if department := c.PostForm("department"); department != "" {
					userPayload["department"] = department
				}
				if officeLocation := c.PostForm("officeLocation"); officeLocation != "" {
					userPayload["officeLocation"] = officeLocation
				}
				if mobilePhone := c.PostForm("mobilePhone"); mobilePhone != "" {
					userPayload["mobilePhone"] = mobilePhone
				}
				
				// Serialize to JSON
				userJSON, marshalErr := json.Marshal(userPayload)
				if marshalErr != nil {
					err = fmt.Errorf("failed to marshal user: %w", marshalErr)
				} else {
					// Create HTTP request
					req, reqErr := http.NewRequestWithContext(c.Request.Context(), "POST", h.client.GetAdapter().GetBaseUrl()+"/users", bytes.NewBuffer(userJSON))
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
										// Create a new user object with the ID
										manualUser := models.NewUser()
										manualUser.SetId(&id)
										created = manualUser
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
			h.render(c, "templates/users/form.html", gin.H{
				"ActiveNav":  "users",
				"IsEdit":     false,
				"FormAction": "/ui/users/new",
				"CancelURL":  "/ui/users",
				"Error":      fmt.Sprintf("Failed to create user: %v", err),
				"Form": map[string]interface{}{
					"DisplayName":       c.PostForm("displayName"),
					"UserPrincipalName": c.PostForm("userPrincipalName"),
					"GivenName":         c.PostForm("givenName"),
					"Surname":           c.PostForm("surname"),
					"Mail":              c.PostForm("mail"),
					"MailNickname":      c.PostForm("mailNickname"),
					"JobTitle":          c.PostForm("jobTitle"),
					"Department":        c.PostForm("department"),
					"OfficeLocation":    c.PostForm("officeLocation"),
					"MobilePhone":       c.PostForm("mobilePhone"),
					"AccountEnabled":    c.PostForm("accountEnabled") == "true",
				},
			})
			return
		}

		// Check if created object is nil
		if created == nil {
			h.render(c, "templates/users/form.html", gin.H{
				"ActiveNav":  "users",
				"IsEdit":     false,
				"FormAction": "/ui/users/new",
				"CancelURL":  "/ui/users",
				"Error":      "User was created but response was empty",
				"Form": map[string]interface{}{
					"DisplayName":       c.PostForm("displayName"),
					"UserPrincipalName": c.PostForm("userPrincipalName"),
					"GivenName":         c.PostForm("givenName"),
					"Surname":           c.PostForm("surname"),
					"Mail":              c.PostForm("mail"),
					"MailNickname":      c.PostForm("mailNickname"),
					"JobTitle":          c.PostForm("jobTitle"),
					"Department":        c.PostForm("department"),
					"OfficeLocation":    c.PostForm("officeLocation"),
					"MobilePhone":       c.PostForm("mobilePhone"),
					"AccountEnabled":    c.PostForm("accountEnabled") == "true",
				},
			})
			return
		}

		// Get user ID - try GetId() first, then fall back to additional data
		var userID string
		if id := created.GetId(); id != nil {
			userID = *id
		} else if additionalData := created.GetAdditionalData(); additionalData != nil {
			if id, ok := additionalData["id"].(string); ok {
				userID = id
			}
		}

		if userID == "" {
			h.render(c, "templates/users/form.html", gin.H{
				"ActiveNav":  "users",
				"IsEdit":     false,
				"FormAction": "/ui/users/new",
				"CancelURL":  "/ui/users",
				"Error":      "User was created but ID was not returned in response",
				"Form": map[string]interface{}{
					"DisplayName":       c.PostForm("displayName"),
					"UserPrincipalName": c.PostForm("userPrincipalName"),
					"GivenName":         c.PostForm("givenName"),
					"Surname":           c.PostForm("surname"),
					"Mail":              c.PostForm("mail"),
					"MailNickname":      c.PostForm("mailNickname"),
					"JobTitle":          c.PostForm("jobTitle"),
					"Department":        c.PostForm("department"),
					"OfficeLocation":    c.PostForm("officeLocation"),
					"MobilePhone":       c.PostForm("mobilePhone"),
					"AccountEnabled":    c.PostForm("accountEnabled") == "true",
				},
			})
			return
		}

		SetFlash(c, FlashSuccess, "User created successfully")
		c.Redirect(http.StatusFound, "/ui/users/"+userID)
	}
}
