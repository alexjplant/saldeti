//go:build e2e

package e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	absser "github.com/microsoft/kiota-abstractions-go/serialization"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/saldeti/saldeti/internal/handler"
	"github.com/saldeti/saldeti/internal/store"
)

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

	ts := httptest.NewServer(router)

	// Create SDK client with custom HTTP client for httptest.Server compatibility
	cred := NewSimulatorCredential(ts.URL, "sim-tenant-id", "sim-client-id", "sim-client-secret")
	
	// Create Kiota authentication provider
	authProvider := NewKiotaAuthenticationProvider(cred)
	
	// Create a custom HTTP client that works with httptest.Server
	customHTTPClient := &http.Client{}
	
	// Create a custom request adapter with the custom HTTP client
	adapter, err := msgraphsdk.NewGraphRequestAdapterWithParseNodeFactoryAndSerializationWriterFactoryAndHttpClient(
		authProvider,
		absser.DefaultParseNodeFactoryInstance,
		absser.DefaultSerializationWriterFactoryInstance,
		customHTTPClient,
	)
	if err != nil {
		t.Fatalf("Failed to create SDK adapter: %v", err)
	}
	
	// Create SDK client with custom adapter
	sdkClient := msgraphsdk.NewGraphServiceClient(adapter)
	// Set base URL without trailing slash
	sdkClient.GetAdapter().SetBaseUrl(ts.URL + "/v1.0")

	return &TestServer{
		Server:    ts,
		Store:     s,
		BaseURL:   ts.URL,
		SDKClient: sdkClient,
	}
}
