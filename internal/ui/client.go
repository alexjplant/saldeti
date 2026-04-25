package ui

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	absser "github.com/microsoft/kiota-abstractions-go/serialization"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/gin-gonic/gin"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/saldeti/saldeti/internal/model"
	kiotaauth "github.com/microsoft/kiota-authentication-azure-go"
)

func ptrString(s string) *string { return &s }
func ptrBool(b bool) *bool { return &b }
func ptrInt32(i int32) *int32 { return &i }

func newInsecureHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // InsecureSkipVerify: acceptable for local simulator
			},
		},
	}
}

func newGraphClient(baseURL string, cred azcore.TokenCredential) (*msgraphsdk.GraphServiceClient, error) {
	// Create Kiota authentication provider using Azure SDK authentication
	authProvider, err := kiotaauth.NewAzureIdentityAuthenticationProvider(cred)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth provider: %w", err)
	}

	// Create a custom HTTP client that skips TLS verification for self-signed certs
	customHTTPClient := newInsecureHTTPClient()

	// Create a custom request adapter with custom HTTP client
	adapter, err := msgraphsdk.NewGraphRequestAdapterWithParseNodeFactoryAndSerializationWriterFactoryAndHttpClient(
		authProvider,
		absser.DefaultParseNodeFactoryInstance,
		absser.DefaultSerializationWriterFactoryInstance,
		customHTTPClient,
	)
	if err != nil {
		return nil, err
	}

	// Create SDK client with custom adapter
	client := msgraphsdk.NewGraphServiceClient(adapter)
	// Set base URL without trailing slash
	client.GetAdapter().SetBaseUrl(baseURL + "/v1.0")

	return client, nil
}

// Convert SDK user to model.User for templates
func sdkUserToModel(u models.Userable) model.User {
	m := model.User{}
	if v := u.GetId(); v != nil {
		m.ID = *v
	}
	if v := u.GetDisplayName(); v != nil {
		m.DisplayName = *v
	}
	if v := u.GetGivenName(); v != nil {
		m.GivenName = *v
	}
	if v := u.GetSurname(); v != nil {
		m.Surname = *v
	}
	if v := u.GetUserPrincipalName(); v != nil {
		m.UserPrincipalName = *v
	}
	if v := u.GetMail(); v != nil {
		m.Mail = *v
	}
	if v := u.GetMailNickname(); v != nil {
		m.MailNickname = *v
	}
	if v := u.GetJobTitle(); v != nil {
		m.JobTitle = *v
	}
	if v := u.GetDepartment(); v != nil {
		m.Department = *v
	}
	if v := u.GetOfficeLocation(); v != nil {
		m.OfficeLocation = *v
	}
	if v := u.GetMobilePhone(); v != nil {
		m.MobilePhone = *v
	}
	m.AccountEnabled = u.GetAccountEnabled()
	if v := u.GetUserType(); v != nil {
		m.UserType = *v
	}
	if v := u.GetCreatedDateTime(); v != nil {
		m.CreatedDateTime = v
	}

	// Extract assignedLicenses from SDK native method
	if licenses := u.GetAssignedLicenses(); licenses != nil {
		for _, lic := range licenses {
			al := model.AssignedLicense{}
			if skuId := lic.GetSkuId(); skuId != nil {
				al.SkuID = skuId.String()
				// Look up skuPartNumber from the static catalog
				if skuPN, found := model.FindSkuBySkuID(al.SkuID); found {
					al.SkuPartNumber = skuPN
				}
			}
			if disabledPlans := lic.GetDisabledPlans(); disabledPlans != nil {
				for _, plan := range disabledPlans {
					al.DisabledPlans = append(al.DisabledPlans, plan.String())
				}
			}
			m.AssignedLicenses = append(m.AssignedLicenses, al)
		}
	}

	return m
}

// Convert SDK directory object to model.DirectoryObject
func sdkDirObjToModel(d models.DirectoryObjectable) model.DirectoryObject {
	m := model.DirectoryObject{}
	if v := d.GetId(); v != nil {
		m.ID = *v
	}
	if v := d.GetOdataType(); v != nil {
		m.ODataType = *v
	}

	// Try to extract display name via type assertion (SDK creates concrete types via discriminator)
	if u, ok := d.(models.Userable); ok {
		if v := u.GetDisplayName(); v != nil {
			m.DisplayName = *v
		}
	} else if g, ok := d.(models.Groupable); ok {
		if v := g.GetDisplayName(); v != nil {
			m.DisplayName = *v
		}
	}

	// Fallback: try additional data
	if m.DisplayName == "" {
		if additionalData := d.GetAdditionalData(); additionalData != nil {
			if dn, ok := additionalData["displayName"]; ok && dn != nil {
				if s, ok := dn.(string); ok {
					m.DisplayName = s
				}
			}
		}
	}

	return m
}

// Convert SDK group to model.Group
func sdkGroupToModel(g models.Groupable) model.Group {
	m := model.Group{}
	if v := g.GetId(); v != nil {
		m.ID = *v
	}
	if v := g.GetDisplayName(); v != nil {
		m.DisplayName = *v
	}
	if v := g.GetDescription(); v != nil {
		m.Description = *v
	}
	if v := g.GetMailNickname(); v != nil {
		m.MailNickname = *v
	}
	if v := g.GetMail(); v != nil {
		m.Mail = *v
	}
	m.MailEnabled = g.GetMailEnabled()
	m.SecurityEnabled = g.GetSecurityEnabled()
	if v := g.GetVisibility(); v != nil {
		m.Visibility = *v
	}
	m.GroupTypes = g.GetGroupTypes()
	if v := g.GetCreatedDateTime(); v != nil {
		m.CreatedDateTime = v
	}
	return m
}

// Convert SDK group to GroupRow for list template
func sdkGroupToGroupRow(g models.Groupable, memberCount int) GroupRow {
	gr := GroupRow{
		Group:       sdkGroupToModel(g),
		MemberCount: memberCount,
	}
	// Determine type label
	gr.TypeLabel = "Security"
	for _, gt := range g.GetGroupTypes() {
		if gt == "Unified" {
			gr.TypeLabel = "Unified (M365)"
			break
		}
	}
	return gr
}

// Helper to safely dereference *string
func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Helper to build form map from gin context
func buildFormMap(c *gin.Context) map[string]string {
	return map[string]string{
		"displayName":       c.PostForm("displayName"),
		"givenName":         c.PostForm("givenName"),
		"surname":           c.PostForm("surname"),
		"userPrincipalName": c.PostForm("userPrincipalName"),
		"mail":              c.PostForm("mail"),
		"mailNickname":      c.PostForm("mailNickname"),
		"jobTitle":          c.PostForm("jobTitle"),
		"department":        c.PostForm("department"),
		"officeLocation":    c.PostForm("officeLocation"),
		"mobilePhone":       c.PostForm("mobilePhone"),
		"accountEnabled":    c.PostForm("accountEnabled"),
	}
}

// fetchDirectoryObjects performs a manual HTTP GET to fetch a list of directory objects
func (h *UIHandler) fetchDirectoryObjects(ctx context.Context, url string) ([]model.DirectoryObject, error) {
	token, err := h.cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://graph.microsoft.com/.default"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // No data is OK for some endpoints
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Value []model.DirectoryObject `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return result.Value, nil
}

// fetchDirectoryObject performs a manual HTTP GET to fetch a single directory object
func (h *UIHandler) fetchDirectoryObject(ctx context.Context, url string) (*model.DirectoryObject, error) {
	token, err := h.cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://graph.microsoft.com/.default"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	var obj model.DirectoryObject
	if err := json.NewDecoder(resp.Body).Decode(&obj); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &obj, nil
}

// fetchSubscribedSkus fetches the subscribed SKU catalog from the API
func (h *UIHandler) fetchSubscribedSkus(ctx context.Context) ([]model.SubscribedSku, error) {
	token, err := h.cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://graph.microsoft.com/.default"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", h.baseURL+"/v1.0/subscribedSkus", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Value []model.SubscribedSku `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return result.Value, nil
}

// fetchAppRoleAssignments performs a manual HTTP GET to fetch a list of app role assignments
func (h *UIHandler) fetchAppRoleAssignments(ctx context.Context, url string) ([]model.AppRoleAssignment, error) {
	token, err := h.cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://graph.microsoft.com/.default"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Value []model.AppRoleAssignment `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return result.Value, nil
}

// fetchOAuth2PermissionGrants performs a manual HTTP GET to fetch a list of OAuth2 permission grants
func (h *UIHandler) fetchOAuth2PermissionGrants(ctx context.Context, url string) ([]model.OAuth2PermissionGrant, error) {
	token, err := h.cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://graph.microsoft.com/.default"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Value []model.OAuth2PermissionGrant `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return result.Value, nil
}

// Convert SDK application to model.Application for templates
func sdkApplicationToModel(a models.Applicationable) model.Application {
	m := model.Application{}
	if v := a.GetId(); v != nil {
		m.ID = *v
	}
	if v := a.GetAppId(); v != nil {
		m.AppID = *v
	}
	if v := a.GetDisplayName(); v != nil {
		m.DisplayName = *v
	}
	if v := a.GetDescription(); v != nil {
		m.Description = *v
	}
	if v := a.GetSignInAudience(); v != nil {
		m.SignInAudience = *v
	}
	if v := a.GetIdentifierUris(); v != nil {
		m.IdentifierUris = v
	}
	if v := a.GetCreatedDateTime(); v != nil {
		m.CreatedDateTime = v
	}
	if v := a.GetPublisherDomain(); v != nil {
		m.PublisherDomain = *v
	}
	m.IsDeviceOnlyAuthSupported = a.GetIsDeviceOnlyAuthSupported()
	m.IsFallbackPublicClient = a.GetIsFallbackPublicClient()
	if v := a.GetTags(); v != nil {
		m.Tags = v
	}
	if creds := a.GetPasswordCredentials(); creds != nil {
		m.PasswordCredentials = sdkPasswordCredentialsToModel(creds)
	}
	if creds := a.GetKeyCredentials(); creds != nil {
		m.KeyCredentials = sdkKeyCredentialsToModel(creds)
	}
	if roles := a.GetAppRoles(); roles != nil {
		m.AppRoles = sdkAppRolesToModel(roles)
	}
	return m
}

// Convert SDK password credentials to model
func sdkPasswordCredentialsToModel(creds []models.PasswordCredentialable) []model.PasswordCredential {
	result := make([]model.PasswordCredential, 0, len(creds))
	for _, c := range creds {
		pc := model.PasswordCredential{}
		if v := c.GetKeyId(); v != nil {
			pc.KeyID = v.String()
		}
		if v := c.GetDisplayName(); v != nil {
			pc.DisplayName = *v
		}
		if v := c.GetHint(); v != nil {
			pc.Hint = *v
		}
		if v := c.GetSecretText(); v != nil {
			pc.SecretText = *v
		}
		if v := c.GetStartDateTime(); v != nil {
			pc.StartDateTime = v
		}
		if v := c.GetEndDateTime(); v != nil {
			pc.EndDateTime = v
		}
		result = append(result, pc)
	}
	return result
}

// Convert SDK key credentials to model
func sdkKeyCredentialsToModel(creds []models.KeyCredentialable) []model.KeyCredential {
	result := make([]model.KeyCredential, 0, len(creds))
	for _, c := range creds {
		kc := model.KeyCredential{}
		if v := c.GetKeyId(); v != nil {
			kc.KeyID = v.String()
		}
		if v := c.GetDisplayName(); v != nil {
			kc.DisplayName = *v
		}
		if v := c.GetTypeEscaped(); v != nil {
			kc.Type = *v
		}
		if v := c.GetUsage(); v != nil {
			kc.Usage = *v
		}
		if v := c.GetStartDateTime(); v != nil {
			kc.StartDateTime = v
		}
		if v := c.GetEndDateTime(); v != nil {
			kc.EndDateTime = v
		}
		result = append(result, kc)
	}
	return result
}

// Convert SDK app roles to model
func sdkAppRolesToModel(roles []models.AppRoleable) []model.AppRole {
	result := make([]model.AppRole, 0, len(roles))
	for _, r := range roles {
		ar := model.AppRole{}
		if v := r.GetId(); v != nil {
			ar.ID = v.String()
		}
		if v := r.GetAllowedMemberTypes(); v != nil {
			ar.AllowedMemberTypes = v
		}
		if v := r.GetDescription(); v != nil {
			ar.Description = *v
		}
		if v := r.GetDisplayName(); v != nil {
			ar.DisplayName = *v
		}
		ar.IsEnabled = r.GetIsEnabled()
		if v := r.GetOrigin(); v != nil {
			ar.Origin = *v
		}
		if v := r.GetValue(); v != nil {
			ar.Value = *v
		}
		result = append(result, ar)
	}
	return result
}

// Convert SDK service principal to model.ServicePrincipal for templates
func sdkServicePrincipalToModel(sp models.ServicePrincipalable) model.ServicePrincipal {
	m := model.ServicePrincipal{}
	if v := sp.GetId(); v != nil {
		m.ID = *v
	}
	if v := sp.GetAppId(); v != nil {
		m.AppID = *v
	}
	if v := sp.GetDisplayName(); v != nil {
		m.DisplayName = *v
	}
	if v := sp.GetDescription(); v != nil {
		m.Description = *v
	}
	if v := sp.GetServicePrincipalType(); v != nil {
		m.ServicePrincipalType = *v
	}
	m.AccountEnabled = sp.GetAccountEnabled()
	if v := sp.GetServicePrincipalNames(); v != nil {
		m.ServicePrincipalNames = v
	}
	if v := sp.GetReplyUrls(); v != nil {
		m.ReplyUrls = v
	}
	if v := sp.GetHomepage(); v != nil {
		m.Homepage = *v
	}
	if v := sp.GetLoginUrl(); v != nil {
		m.LoginURL = *v
	}
	if v := sp.GetTags(); v != nil {
		m.Tags = v
	}
	if creds := sp.GetPasswordCredentials(); creds != nil {
		m.PasswordCredentials = sdkPasswordCredentialsToModel(creds)
	}
	if creds := sp.GetKeyCredentials(); creds != nil {
		m.KeyCredentials = sdkKeyCredentialsToModel(creds)
	}
	return m
}

// fetchExtensionProperties performs a manual HTTP GET to fetch a list of extension properties
func (h *UIHandler) fetchExtensionProperties(ctx context.Context, url string) ([]model.ExtensionProperty, error) {
	token, err := h.cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://graph.microsoft.com/.default"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Value []model.ExtensionProperty `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return result.Value, nil
}
