//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	abs "github.com/microsoft/kiota-abstractions-go"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

type SimulatorCredential struct {
	tokenEndpoint string
	clientID      string
	clientSecret  string
	scope         string
}

func NewSimulatorCredential(serverURL, tenantID, clientID, clientSecret string) *SimulatorCredential {
	return &SimulatorCredential{
		tokenEndpoint: fmt.Sprintf("%s/%s/oauth2/v2.0/token", serverURL, tenantID),
		clientID:      clientID,
		clientSecret:  clientSecret,
		scope:         "https://graph.microsoft.com/.default",
	}
}

func (c *SimulatorCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("scope", c.scope)

	req, err := http.NewRequestWithContext(ctx, "POST", c.tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return azcore.AccessToken{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return azcore.AccessToken{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return azcore.AccessToken{}, fmt.Errorf("token request failed (%d): %s", resp.StatusCode, body)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return azcore.AccessToken{}, err
	}

	return azcore.AccessToken{
		Token:     tokenResp.AccessToken,
		ExpiresOn: time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}, nil
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
