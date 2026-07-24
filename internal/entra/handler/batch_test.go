package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/entra/auth"
	"github.com/saldeti/saldeti/internal/entra/model"
	"github.com/saldeti/saldeti/internal/entra/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchHandler_ExceedsLimit(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Build 21 sub-requests
	requests := make([]map[string]any, 21)
	for i := 0; i < 21; i++ {
		requests[i] = map[string]any{
			"id":     fmt.Sprintf("%d", i+1),
			"method": "GET",
			"url":    "/v1.0/me",
		}
	}
	batchBody, _ := json.Marshal(map[string]any{
		"requests": requests,
	})

	req, err := http.NewRequest("POST", server.URL+"/v1.0/$batch", strings.NewReader(string(batchBody)))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var errResp map[string]any
	err = json.Unmarshal(body, &errResp)
	require.NoError(t, err)

	errObj := errResp["error"].(map[string]any)
	assert.Equal(t, "BadRequest", errObj["code"])
	assert.Contains(t, errObj["message"], "exceeds maximum of 20")
}

func TestBatchHandler_NoContentResponse(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create a user to delete
	accountEnabled := true
	user := model.User{
		DisplayName:       "Batch Delete Test",
		UserPrincipalName: "batchdelete@example.com",
		Mail:              "batchdelete@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.ReadWrite.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Build batch with a DELETE sub-request
	batchBody, _ := json.Marshal(map[string]any{
		"requests": []map[string]any{
			{
				"id":     "1",
				"method": "DELETE",
				"url":    "/v1.0/users/" + createdUser.ID,
			},
		},
	})

	req, err := http.NewRequest("POST", server.URL+"/v1.0/$batch", strings.NewReader(string(batchBody)))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var batchResp map[string]any
	err = json.Unmarshal(body, &batchResp)
	require.NoError(t, err)

	responses := batchResp["responses"].([]any)
	require.Len(t, responses, 1)

	subResp := responses[0].(map[string]any)
	assert.Equal(t, "1", subResp["id"])
	// Status should be 204 (float64 from JSON unmarshal)
	assert.Equal(t, float64(204), subResp["status"])
	// Body should be nil/empty for 204 No Content
	assert.Nil(t, subResp["body"])
}

func TestBatchHandler_RelativeURLNoPrefix(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	accountEnabled := true
	user := model.User{
		DisplayName:       "Batch Relative Test",
		UserPrincipalName: "batchrelative@example.com",
		Mail:              "batchrelative@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.ReadWrite.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	batchBody, _ := json.Marshal(map[string]any{
		"requests": []map[string]any{
			{
				"id":     "1",
				"method": "DELETE",
				"url":    "/users/" + createdUser.ID,
			},
		},
	})

	req, err := http.NewRequest("POST", server.URL+"/v1.0/$batch", strings.NewReader(string(batchBody)))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var batchResp map[string]any
	err = json.Unmarshal(body, &batchResp)
	require.NoError(t, err)

	responses := batchResp["responses"].([]any)
	assert.Len(t, responses, 1)

	subResp := responses[0].(map[string]any)
	assert.Equal(t, "1", subResp["id"])
	assert.Equal(t, float64(204), subResp["status"])
	assert.Nil(t, subResp["body"])
}

func TestNormalizeBatchURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/me", "/v1.0/me"},
		{"/groups/abc/members/$ref", "/v1.0/groups/abc/members/$ref"},
		{"/v1.0/me", "/v1.0/me"},
		{"/v1.0/users/123", "/v1.0/users/123"},
		{"/beta/me", "/beta/me"},
		{"me", "/v1.0/me"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, normalizeBatchURL(tt.input), "normalizeBatchURL(%q)", tt.input)
	}
}

func TestBatchHandler_MixedSuccessFailure(t *testing.T) {
	s := store.NewMemoryStore()
	router := NewRouter(s)
	ctx := context.Background()

	accountEnabled := true
	user := model.User{
		DisplayName:       "Mixed Test User",
		UserPrincipalName: "mixed@example.com",
		Mail:              "mixed@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.ReadWrite.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	batchBody, _ := json.Marshal(map[string]any{
		"requests": []map[string]any{
			{
				"id":     "1",
				"method": "GET",
				"url":    "/v1.0/users/" + createdUser.ID,
			},
			{
				"id":     "2",
				"method": "GET",
				"url":    "/v1.0/users/nonexistent-id-12345",
			},
			{
				"id":     "3",
				"method": "DELETE",
				"url":    "/v1.0/users/nonexistent-id-67890",
			},
		},
	})

	req, err := http.NewRequest("POST", server.URL+"/v1.0/$batch", strings.NewReader(string(batchBody)))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var batchResp map[string]any
	err = json.Unmarshal(body, &batchResp)
	require.NoError(t, err)

	responses := batchResp["responses"].([]any)
	require.Len(t, responses, 3)

	// First request: 200 OK
	subResp1 := responses[0].(map[string]any)
	assert.Equal(t, "1", subResp1["id"])
	assert.Equal(t, float64(200), subResp1["status"])

	// Second request: 404 Not Found
	subResp2 := responses[1].(map[string]any)
	assert.Equal(t, "2", subResp2["id"])
	assert.Equal(t, float64(404), subResp2["status"])

	// Third request: 404 Not Found
	subResp3 := responses[2].(map[string]any)
	assert.Equal(t, "3", subResp3["id"])
	assert.Equal(t, float64(404), subResp3["status"])
}

func TestBatchHandler_MultiRequestBatch(t *testing.T) {
	s := store.NewMemoryStore()
	router := NewRouter(s)
	ctx := context.Background()

	accountEnabled := true
	createdUser, err := s.CreateUser(ctx, model.User{
		DisplayName:       "Multi User",
		UserPrincipalName: "multi@example.com",
		Mail:              "multi@example.com",
		AccountEnabled:    &accountEnabled,
	})
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.ReadWrite.All", "Group.ReadWrite.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	batchBody, _ := json.Marshal(map[string]any{
		"requests": []map[string]any{
			{
				"id":     "1",
				"method": "GET",
				"url":    "/v1.0/users",
			},
			{
				"id":     "2",
				"method": "POST",
				"url":    "/v1.0/groups",
				"body": map[string]any{
					"displayName": "Batch Test Group",
				},
			},
			{
				"id":     "3",
				"method": "GET",
				"url":    "/v1.0/groups",
			},
			{
				"id":     "4",
				"method": "GET",
				"url":    "/v1.0/subscribedSkus",
			},
			{
				"id":     "5",
				"method": "GET",
				"url":    "/v1.0/users/" + createdUser.ID,
			},
		},
	})

	req, err := http.NewRequest("POST", server.URL+"/v1.0/$batch", strings.NewReader(string(batchBody)))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var batchResp map[string]any
	err = json.Unmarshal(body, &batchResp)
	require.NoError(t, err)

	responses := batchResp["responses"].([]any)
	require.Len(t, responses, 5)

	// Request 1: GET /users -> 200
	assert.Equal(t, float64(200), responses[0].(map[string]any)["status"])
	// Request 2: POST /groups -> 201
	assert.Equal(t, float64(201), responses[1].(map[string]any)["status"])
	// Request 3: GET /groups -> 200
	assert.Equal(t, float64(200), responses[2].(map[string]any)["status"])
	// Request 4: GET /subscribedSkus -> 200
	assert.Equal(t, float64(200), responses[3].(map[string]any)["status"])
	// Request 5: GET /users/{id} -> 200
	assert.Equal(t, float64(200), responses[4].(map[string]any)["status"])
}

func TestBatchHandler_CreateAndGetInBatch(t *testing.T) {
	s := store.NewMemoryStore()
	router := NewRouter(s)
	ctx := context.Background()

	accountEnabled := true
	createdUser, err := s.CreateUser(ctx, model.User{
		DisplayName:       "Existing User",
		UserPrincipalName: "existing@example.com",
		Mail:              "existing@example.com",
		AccountEnabled:    &accountEnabled,
	})
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.ReadWrite.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	batchBody, _ := json.Marshal(map[string]any{
		"requests": []map[string]any{
			{
				"id":     "1",
				"method": "POST",
				"url":    "/v1.0/users",
				"body": map[string]any{
					"displayName":       "Created In Batch",
					"userPrincipalName": "batchcreated@example.com",
					"mail":              "batchcreated@example.com",
					"accountEnabled":    true,
				},
			},
			{
				"id":     "2",
				"method": "GET",
				"url":    "/v1.0/users/" + createdUser.ID,
			},
		},
	})

	req, err := http.NewRequest("POST", server.URL+"/v1.0/$batch", strings.NewReader(string(batchBody)))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var batchResp map[string]any
	err = json.Unmarshal(body, &batchResp)
	require.NoError(t, err)

	responses := batchResp["responses"].([]any)
	require.Len(t, responses, 2)

	// Request 1: POST create user -> 201
	subResp1 := responses[0].(map[string]any)
	assert.Equal(t, "1", subResp1["id"])
	assert.Equal(t, float64(201), subResp1["status"])
	assert.NotNil(t, subResp1["body"])

	// Request 2: GET existing user -> 200
	subResp2 := responses[1].(map[string]any)
	assert.Equal(t, "2", subResp2["id"])
	assert.Equal(t, float64(200), subResp2["status"])
}

func TestBatchHandler_InvalidJSON(t *testing.T) {
	s := store.NewMemoryStore()
	router := NewRouter(s)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.ReadWrite.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("POST", server.URL+"/v1.0/$batch", strings.NewReader("{not valid json}"))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var errResp map[string]any
	err = json.Unmarshal(body, &errResp)
	require.NoError(t, err)

	errObj := errResp["error"].(map[string]any)
	assert.Equal(t, "BadRequest", errObj["code"])
}

func TestBatchHandler_EmptyRequests(t *testing.T) {
	s := store.NewMemoryStore()
	router := NewRouter(s)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.ReadWrite.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	batchBody, _ := json.Marshal(map[string]any{
		"requests": []map[string]any{},
	})

	req, err := http.NewRequest("POST", server.URL+"/v1.0/$batch", strings.NewReader(string(batchBody)))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var batchResp map[string]any
	err = json.Unmarshal(body, &batchResp)
	require.NoError(t, err)

	responses := batchResp["responses"].([]any)
	assert.Len(t, responses, 0)
}

func TestBatchHandler_NonExistentURL(t *testing.T) {
	s := store.NewMemoryStore()
	router := NewRouter(s)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.ReadWrite.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	batchBody, _ := json.Marshal(map[string]any{
		"requests": []map[string]any{
			{
				"id":     "1",
				"method": "GET",
				"url":    "/v1.0/thisEndpointDoesNotExist",
			},
		},
	})

	req, err := http.NewRequest("POST", server.URL+"/v1.0/$batch", strings.NewReader(string(batchBody)))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var batchResp map[string]any
	err = json.Unmarshal(body, &batchResp)
	require.NoError(t, err)

	responses := batchResp["responses"].([]any)
	assert.Len(t, responses, 1)

	subResp := responses[0].(map[string]any)
	assert.Equal(t, "1", subResp["id"])
	assert.Equal(t, float64(404), subResp["status"])
}

func TestBatchHandler_MarshalError(t *testing.T) {
	s := store.NewMemoryStore()
	router := NewRouter(s)

	// Construct a test context for the batch handler. The marshal-error path
	// is unreachable via HTTP because ShouldBindJSON decodes sub.Body into a
	// map[string]interface{} whose values are always JSON-representable Go
	// types. We therefore invoke the sub-request processing function directly
	// with a struct whose body contains a value that cannot be marshaled
	// (a channel).
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1.0/$batch", nil)

	requests := []BatchSubRequest{
		{
			ID:     "1",
			Method: "POST",
			URL:    "/v1.0/users",
			Body:   map[string]any{"bad": make(chan int)},
		},
	}

	responses := processBatchSubRequests(c, router, requests)
	require.Len(t, responses, 1)

	subResp := responses[0]
	assert.Equal(t, "1", subResp.ID)
	assert.Equal(t, http.StatusInternalServerError, subResp.Status)

	errObj, ok := subResp.Body["error"]
	require.True(t, ok, "error field should be present in body")

	errMap, ok := errObj.(gin.H)
	require.True(t, ok, "error should be a gin.H map")

	assert.Equal(t, "InternalError", errMap["code"])
	assert.Equal(t, "Failed to marshal request body", errMap["message"])
}
