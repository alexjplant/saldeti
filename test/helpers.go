//go:build e2e

package e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	absser "github.com/microsoft/kiota-abstractions-go/serialization"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	kiotaauth "github.com/microsoft/kiota-authentication-azure-go"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/saldeti/saldeti/internal/handler"
	"github.com/saldeti/saldeti/internal/store"
)

// httpTransport wraps an http.Client to implement policy.Transporter
type httpTransport struct {
	client *http.Client
}

func (t *httpTransport) Do(req *http.Request) (*http.Response, error) {
	return t.client.Do(req)
}

// TestServer wraps a test server with authentication helpers
type TestServer struct {
	Server    *httptest.Server
	Store     store.Store
	BaseURL   string
	SDKClient *msgraphsdk.GraphServiceClient
}

// setupTestServer creates a store, seeds data, creates router, and returns test server
func setupTestServer(t *testing.T) *TestServer {
	s := store.NewMemoryStore()

	// Register default client
	ctx := context.Background()
	err := s.RegisterClient(ctx, "sim-client-id", "sim-client-secret", "sim-tenant-id")
	if err != nil && err.Error() != "client already registered" {
		t.Fatalf("Failed to register client: %v", err)
	}

	router := handler.NewRouter(s)

	// Use TLS test server for HTTPS
	ts := httptest.NewTLSServer(router)

	// Create azidentity credential pointing at test server
	// Use test server's client transport which already trusts the test server's cert
	cred, err := azidentity.NewClientSecretCredential(
		"sim-tenant-id", "sim-client-id", "sim-client-secret",
		&azidentity.ClientSecretCredentialOptions{
			ClientOptions: azcore.ClientOptions{
				Cloud: cloud.Configuration{
					ActiveDirectoryAuthorityHost: ts.URL,
				},
				Transport: &httpTransport{client: ts.Client()},
			},
			DisableInstanceDiscovery: true,
		},
	)
	if err != nil {
		t.Fatalf("Failed to create credential: %v", err)
	}

	// Create Kiota auth provider
	authProvider, err := kiotaauth.NewAzureIdentityAuthenticationProvider(cred)
	if err != nil {
		t.Fatalf("Failed to create auth provider: %v", err)
	}

	// Use test server's client for SDK calls (trusts the test server's cert)
	adapter, err := msgraphsdk.NewGraphRequestAdapterWithParseNodeFactoryAndSerializationWriterFactoryAndHttpClient(
		authProvider,
		absser.DefaultParseNodeFactoryInstance,
		absser.DefaultSerializationWriterFactoryInstance,
		ts.Client(),
	)
	if err != nil {
		t.Fatalf("Failed to create SDK adapter: %v", err)
	}

	sdkClient := msgraphsdk.NewGraphServiceClient(adapter)
	sdkClient.GetAdapter().SetBaseUrl(ts.URL + "/v1.0")

	return &TestServer{
		Server:    ts,
		Store:     s,
		BaseURL:   ts.URL,
		SDKClient: sdkClient,
	}
}
