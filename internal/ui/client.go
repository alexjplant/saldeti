package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	absser "github.com/microsoft/kiota-abstractions-go/serialization"
	abs "github.com/microsoft/kiota-abstractions-go"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/gin-gonic/gin"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/saldeti/saldeti/internal/model"
)

func ptrString(s string) *string { return &s }
func ptrBool(b bool) *bool { return &b }
func ptrInt32(i int32) *int32 { return &i }

func newGraphClient(baseURL string, cred *SimulatorCredential) (*msgraphsdk.GraphServiceClient, error) {
	// Create Kiota authentication provider
	authProvider := NewKiotaAuthenticationProvider(cred)

	// Create a custom HTTP client
	customHTTPClient := &http.Client{}

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

// KiotaAuthenticationProvider adapts SimulatorCredential to work with Kiota's AuthenticationProvider interface
type KiotaAuthenticationProvider struct {
	cred *SimulatorCredential
}

// NewKiotaAuthenticationProvider creates a new Kiota authentication provider
func NewKiotaAuthenticationProvider(cred *SimulatorCredential) *KiotaAuthenticationProvider {
	return &KiotaAuthenticationProvider{
		cred: cred,
	}
}

// AuthenticateRequest implements the Kiota AuthenticationProvider interface
func (p *KiotaAuthenticationProvider) AuthenticateRequest(ctx context.Context, request *abs.RequestInformation, additionalAuthenticationContext map[string]interface{}) error {
	// Get token from the underlying credential
	token, err := p.cred.GetToken(ctx, policy.TokenRequestOptions{})
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	// Set authorization header
	if request.Headers == nil {
		request.Headers = abs.NewRequestHeaders()
	}
	request.Headers.Add("Authorization", "Bearer "+token.Token)

	return nil
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
	token, err := h.cred.GetToken(ctx, policy.TokenRequestOptions{})
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
	token, err := h.cred.GetToken(ctx, policy.TokenRequestOptions{})
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
