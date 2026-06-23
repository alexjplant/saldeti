package ui

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/saldeti/saldeti/internal/google/model"
)

var errNotFound = fmt.Errorf("not found")

type GoogleClient struct {
	baseURL      string
	clientID     string
	clientSecret string
	httpClient   *http.Client
	mu           sync.Mutex
	token        string
	tokenExpiry  time.Time
}

func NewGoogleClient(baseURL, clientID, clientSecret string) *GoogleClient {
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &GoogleClient{
		baseURL:      baseURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // InsecureSkipVerify: acceptable for local simulator
				},
			},
		},
	}
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func (c *GoogleClient) getToken(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Until(c.tokenExpiry) > 5*time.Minute {
		return nil
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("scope", "https://www.googleapis.com/auth/admin.directory.user https://www.googleapis.com/auth/admin.directory.group https://www.googleapis.com/auth/admin.directory.device.chromeos https://www.googleapis.com/auth/admin.directory.device.mobile https://www.googleapis.com/auth/cloud-platform")

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/oauth2/v1/token", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	c.token = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

func (c *GoogleClient) doGet(ctx context.Context, path string, result interface{}) error {
	if err := c.getToken(ctx); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return errNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GET %s failed with status %d: %s", path, resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

func (c *GoogleClient) doPost(ctx context.Context, path string, body interface{}, result interface{}) error {
	if err := c.getToken(ctx); err != nil {
		return err
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("POST %s failed with status %d: %s", path, resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *GoogleClient) doPut(ctx context.Context, path string, body interface{}, result interface{}) error {
	if err := c.getToken(ctx); err != nil {
		return err
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", c.baseURL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("PUT %s failed with status %d: %s", path, resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *GoogleClient) doDelete(ctx context.Context, path string) error {
	if err := c.getToken(ctx); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("DELETE %s failed with status %d: %s", path, resp.StatusCode, string(body))
	}

	return nil
}

func (c *GoogleClient) GetUsers(ctx context.Context) ([]model.User, error) {
	var resp struct {
		Users []model.User `json:"users"`
	}
	if err := c.doGet(ctx, "/admin/directory/v1/users?maxResults=500", &resp); err != nil {
		return nil, err
	}
	if resp.Users == nil {
		return []model.User{}, nil
	}
	return resp.Users, nil
}

func (c *GoogleClient) GetUser(ctx context.Context, userKey string) (*model.User, error) {
	var user model.User
	if err := c.doGet(ctx, "/admin/directory/v1/users/"+userKey, &user); err != nil {
		if err == errNotFound {
			return nil, errNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (c *GoogleClient) CreateUser(ctx context.Context, user model.User) (*model.User, error) {
	var created model.User
	if err := c.doPost(ctx, "/admin/directory/v1/users", user, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

func (c *GoogleClient) UpdateUser(ctx context.Context, userKey string, user model.User) (*model.User, error) {
	var updated model.User
	if err := c.doPut(ctx, "/admin/directory/v1/users/"+userKey, user, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

func (c *GoogleClient) DeleteUser(ctx context.Context, userKey string) error {
	return c.doDelete(ctx, "/admin/directory/v1/users/"+userKey)
}

func (c *GoogleClient) GetGroups(ctx context.Context) ([]model.Group, error) {
	var resp struct {
		Groups []model.Group `json:"groups"`
	}
	if err := c.doGet(ctx, "/admin/directory/v1/groups?maxResults=500", &resp); err != nil {
		return nil, err
	}
	if resp.Groups == nil {
		return []model.Group{}, nil
	}
	return resp.Groups, nil
}

func (c *GoogleClient) GetGroup(ctx context.Context, groupKey string) (*model.Group, error) {
	var group model.Group
	if err := c.doGet(ctx, "/admin/directory/v1/groups/"+groupKey, &group); err != nil {
		if err == errNotFound {
			return nil, errNotFound
		}
		return nil, err
	}
	return &group, nil
}

func (c *GoogleClient) CreateGroup(ctx context.Context, group model.Group) (*model.Group, error) {
	var created model.Group
	if err := c.doPost(ctx, "/admin/directory/v1/groups", group, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

func (c *GoogleClient) UpdateGroup(ctx context.Context, groupKey string, group model.Group) (*model.Group, error) {
	var updated model.Group
	if err := c.doPut(ctx, "/admin/directory/v1/groups/"+groupKey, group, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

func (c *GoogleClient) DeleteGroup(ctx context.Context, groupKey string) error {
	return c.doDelete(ctx, "/admin/directory/v1/groups/"+groupKey)
}

func (c *GoogleClient) GetMembers(ctx context.Context, groupKey string) ([]model.Member, error) {
	var resp struct {
		Members []model.Member `json:"members"`
	}
	if err := c.doGet(ctx, "/admin/directory/v1/groups/"+groupKey+"/members", &resp); err != nil {
		return nil, err
	}
	if resp.Members == nil {
		return []model.Member{}, nil
	}
	return resp.Members, nil
}

func (c *GoogleClient) AddMember(ctx context.Context, groupKey string, member model.Member) (*model.Member, error) {
	var created model.Member
	if err := c.doPost(ctx, "/admin/directory/v1/groups/"+groupKey+"/members", member, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

func (c *GoogleClient) RemoveMember(ctx context.Context, groupKey, memberKey string) error {
	return c.doDelete(ctx, "/admin/directory/v1/groups/"+groupKey+"/members/"+memberKey)
}

func (c *GoogleClient) GetOrgUnits(ctx context.Context, customer string) ([]model.OrgUnit, error) {
	var resp struct {
		OrganizationUnits []model.OrgUnit `json:"organizationUnits"`
	}
	if err := c.doGet(ctx, "/admin/directory/v1/customer/"+customer+"/orgunits/", &resp); err != nil {
		return nil, err
	}
	if resp.OrganizationUnits == nil {
		return []model.OrgUnit{}, nil
	}
	return resp.OrganizationUnits, nil
}

func (c *GoogleClient) CreateOrgUnit(ctx context.Context, customer string, ou model.OrgUnit) (*model.OrgUnit, error) {
	var created model.OrgUnit
	if err := c.doPost(ctx, "/admin/directory/v1/customer/"+customer+"/orgunits", ou, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

func (c *GoogleClient) GetChromeOSDevices(ctx context.Context, customer string) ([]model.ChromeOSDevice, error) {
	var resp struct {
		ChromeOSDevices []model.ChromeOSDevice `json:"chromeosdevices"`
	}
	if err := c.doGet(ctx, "/admin/directory/v1/customer/"+customer+"/devices/chromeos?maxResults=500", &resp); err != nil {
		return nil, err
	}
	if resp.ChromeOSDevices == nil {
		return []model.ChromeOSDevice{}, nil
	}
	return resp.ChromeOSDevices, nil
}

func (c *GoogleClient) GetMobileDevices(ctx context.Context, customer string) ([]model.MobileDevice, error) {
	var resp struct {
		MobileDevices []model.MobileDevice `json:"mobiledevices"`
	}
	if err := c.doGet(ctx, "/admin/directory/v1/customer/"+customer+"/devices/mobile?maxResults=500", &resp); err != nil {
		return nil, err
	}
	if resp.MobileDevices == nil {
		return []model.MobileDevice{}, nil
	}
	return resp.MobileDevices, nil
}

func (c *GoogleClient) GetRoles(ctx context.Context, customer string) ([]model.Role, error) {
	var resp struct {
		Items []model.Role `json:"items"`
	}
	if err := c.doGet(ctx, "/admin/directory/v1/customer/"+customer+"/roles", &resp); err != nil {
		return nil, err
	}
	if resp.Items == nil {
		return []model.Role{}, nil
	}
	return resp.Items, nil
}

func (c *GoogleClient) GetDomains(ctx context.Context, customer string) ([]model.Domain, error) {
	var resp struct {
		Domains []model.Domain `json:"domains"`
	}
	if err := c.doGet(ctx, "/admin/directory/v1/customer/"+customer+"/domains", &resp); err != nil {
		return nil, err
	}
	if resp.Domains == nil {
		return []model.Domain{}, nil
	}
	return resp.Domains, nil
}
