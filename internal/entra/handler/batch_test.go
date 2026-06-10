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
	requests := make([]map[string]interface{}, 21)
	for i := 0; i < 21; i++ {
		requests[i] = map[string]interface{}{
			"id":     fmt.Sprintf("%d", i+1),
			"method": "GET",
			"url":    "/v1.0/me",
		}
	}
	batchBody, _ := json.Marshal(map[string]interface{}{
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

	var errResp map[string]interface{}
	err = json.Unmarshal(body, &errResp)
	require.NoError(t, err)

	errObj := errResp["error"].(map[string]interface{})
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
	batchBody, _ := json.Marshal(map[string]interface{}{
		"requests": []map[string]interface{}{
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

	var batchResp map[string]interface{}
	err = json.Unmarshal(body, &batchResp)
	require.NoError(t, err)

	responses := batchResp["responses"].([]interface{})
	assert.Len(t, responses, 1)

	subResp := responses[0].(map[string]interface{})
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

	batchBody, _ := json.Marshal(map[string]interface{}{
		"requests": []map[string]interface{}{
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

	var batchResp map[string]interface{}
	err = json.Unmarshal(body, &batchResp)
	require.NoError(t, err)

	responses := batchResp["responses"].([]interface{})
	assert.Len(t, responses, 1)

	subResp := responses[0].(map[string]interface{})
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
