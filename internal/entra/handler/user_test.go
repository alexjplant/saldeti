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

func TestListUsers(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create 5 test users
	for i := 1; i <= 5; i++ {
		accountEnabled := true
		user := model.User{
			DisplayName:       fmt.Sprintf("User %d", i),
			UserPrincipalName: fmt.Sprintf("user%d@example.com", i),
			Mail:              fmt.Sprintf("user%d@example.com", i),
			Department:        fmt.Sprintf("Dept %d", (i%3)+1),
			AccountEnabled:    &accountEnabled,
		}
		_, err := store.CreateUser(ctx, user)
		require.NoError(t, err)
	}

	server := httptest.NewServer(router)
	defer server.Close()

	// Mint a token
	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Test listing all users
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	assert.Equal(t, "https://graph.microsoft.com/v1.0/$metadata#users", listResp["@odata.context"])
	users := listResp["value"].([]any)
	assert.Len(t, users, 5)
}

func TestListUsersWithFilter(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create users with different departments
	departments := []string{"Engineering", "Sales", "Engineering", "Marketing", "Engineering"}
	for i, dept := range departments {
		accountEnabled := true
		user := model.User{
			DisplayName:       fmt.Sprintf("User %d", i+1),
			UserPrincipalName: fmt.Sprintf("user%d@example.com", i+1),
			Mail:              fmt.Sprintf("user%d@example.com", i+1),
			Department:        dept,
			AccountEnabled:    &accountEnabled,
		}
		_, err := store.CreateUser(ctx, user)
		require.NoError(t, err)
	}

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Filter by department
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$filter=department%20eq%20'Engineering'", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	users := listResp["value"].([]any)
	assert.Len(t, users, 3) // Should have 3 Engineering users
}

func TestListUsersWithSelect(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "test@example.com",
		Mail:              "test@example.com",
		Department:        "Engineering",
		JobTitle:          "Developer",
		AccountEnabled:    &accountEnabled,
	}
	_, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Select specific fields
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$select=displayName,id,department", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	users := listResp["value"].([]any)
	require.Len(t, users, 1)

	userMap := users[0].(map[string]any)
	assert.Contains(t, userMap, "displayName")
	assert.Contains(t, userMap, "id")
	assert.Contains(t, userMap, "department")
	assert.NotContains(t, userMap, "jobTitle") // Should not be included
	assert.NotContains(t, userMap, "mail")     // Should not be included
}

func TestListUsersWithTop(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create 10 users
	for i := 1; i <= 10; i++ {
		accountEnabled := true
		user := model.User{
			DisplayName:       fmt.Sprintf("User %d", i),
			UserPrincipalName: fmt.Sprintf("user%d@example.com", i),
			Mail:              fmt.Sprintf("user%d@example.com", i),
			AccountEnabled:    &accountEnabled,
		}
		_, err := store.CreateUser(ctx, user)
		require.NoError(t, err)
	}

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Get only 5 users
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$top=5", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	users := listResp["value"].([]any)
	assert.Len(t, users, 5)

	// Check for nextLink
	assert.Contains(t, listResp, "@odata.nextLink")
	nextLink := listResp["@odata.nextLink"].(string)
	// URL might be encoded, check for skip parameter
	assert.Contains(t, nextLink, "skip=5")
}

func TestListUsersWithCount(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create 3 users
	for i := 1; i <= 3; i++ {
		accountEnabled := true
		user := model.User{
			DisplayName:       fmt.Sprintf("User %d", i),
			UserPrincipalName: fmt.Sprintf("user%d@example.com", i),
			Mail:              fmt.Sprintf("user%d@example.com", i),
			AccountEnabled:    &accountEnabled,
		}
		_, err := store.CreateUser(ctx, user)
		require.NoError(t, err)
	}

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Request with count
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$count=true", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	assert.Contains(t, listResp, "@odata.count")
	count := int(listResp["@odata.count"].(float64))
	assert.Equal(t, 3, count)
}

func TestGetUserByID(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "test@example.com",
		Mail:              "test@example.com",
		Department:        "Engineering",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Get user by ID
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users/"+createdUser.ID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var userResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &userResp)
	require.NoError(t, err)

	assert.Equal(t, "https://graph.microsoft.com/v1.0/$metadata#users/$entity", userResp["@odata.context"])
	assert.Equal(t, createdUser.ID, userResp["id"])
	assert.Equal(t, "test@example.com", userResp["userPrincipalName"])
	assert.Equal(t, "Test User", userResp["displayName"])
}

func TestGetUserByUPN(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "test@example.com",
		Mail:              "test@example.com",
		Department:        "Engineering",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Get user by UPN
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users/test@example.com", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var userResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &userResp)
	require.NoError(t, err)

	assert.Equal(t, createdUser.ID, userResp["id"])
	assert.Equal(t, "test@example.com", userResp["userPrincipalName"])
}

func TestGetUserNotFound(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Get non-existent user
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users/nonexistent", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCreateUser(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.ReadWrite.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Create user
	userJSON := `{
		"displayName": "New User",
		"userPrincipalName": "newuser@example.com",
		"mail": "newuser@example.com",
		"department": "Engineering",
		"accountEnabled": true
	}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/users", strings.NewReader(userJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Location"), "/v1.0/users/")

	var userResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &userResp)
	require.NoError(t, err)

	assert.Equal(t, "https://graph.microsoft.com/v1.0/$metadata#users/$entity", userResp["@odata.context"])
	assert.Equal(t, "New User", userResp["displayName"])
	assert.Equal(t, "newuser@example.com", userResp["userPrincipalName"])
	assert.Contains(t, userResp, "id")
	assert.NotEmpty(t, userResp["id"])
}

func TestCreateUserDuplicateUPN(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create first user
	accountEnabled := true
	user := model.User{
		DisplayName:       "Existing User",
		UserPrincipalName: "existing@example.com",
		Mail:              "existing@example.com",
		AccountEnabled:    &accountEnabled,
	}
	_, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.ReadWrite.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Try to create user with same UPN
	userJSON := `{
		"displayName": "Duplicate User",
		"userPrincipalName": "existing@example.com",
		"mail": "duplicate@example.com"
	}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/users", strings.NewReader(userJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestCreateUserMissingFields(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.ReadWrite.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Try to create user without displayName
	userJSON := `{
		"userPrincipalName": "nofields@example.com"
	}`

	req, err := http.NewRequest("POST", server.URL+"/v1.0/users", strings.NewReader(userJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUpdateUser(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	accountEnabled := true
	user := model.User{
		DisplayName:       "Original User",
		UserPrincipalName: "original@example.com",
		Mail:              "original@example.com",
		Department:        "Sales",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.ReadWrite.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Update user
	patchJSON := `{
		"displayName": "Updated User",
		"department": "Engineering",
		"jobTitle": "Senior Developer"
	}`

	req, err := http.NewRequest("PATCH", server.URL+"/v1.0/users/"+createdUser.ID, strings.NewReader(patchJSON))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var userResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &userResp)
	require.NoError(t, err)

	assert.Equal(t, "Updated User", userResp["displayName"])
	assert.Equal(t, "Engineering", userResp["department"])
	assert.Equal(t, "Senior Developer", userResp["jobTitle"])
	assert.Equal(t, "original@example.com", userResp["userPrincipalName"]) // Should not change
}

func TestDeleteUser(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	accountEnabled := true
	user := model.User{
		DisplayName:       "User to Delete",
		UserPrincipalName: "delete@example.com",
		Mail:              "delete@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.ReadWrite.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Delete user
	req, err := http.NewRequest("DELETE", server.URL+"/v1.0/users/"+createdUser.ID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify user is deleted
	req, err = http.NewRequest("GET", server.URL+"/v1.0/users/"+createdUser.ID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestODataFilterStartsWith(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create users with different names
	names := []string{"Alice", "Bob", "Charlie", "Alex", "David"}
	for _, name := range names {
		accountEnabled := true
		user := model.User{
			DisplayName:       name,
			UserPrincipalName: fmt.Sprintf("%s@example.com", strings.ToLower(name)),
			Mail:              fmt.Sprintf("%s@example.com", strings.ToLower(name)),
			AccountEnabled:    &accountEnabled,
		}
		_, err := store.CreateUser(ctx, user)
		require.NoError(t, err)
	}

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Filter names starting with 'A'
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$filter=startswith(displayName,'A')", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	users := listResp["value"].([]any)
	assert.Len(t, users, 2) // Alice and Alex

	// Verify correct users
	userNames := make([]string, len(users))
	for i, u := range users {
		userMap := u.(map[string]any)
		userNames[i] = userMap["displayName"].(string)
	}

	assert.Contains(t, userNames, "Alice")
	assert.Contains(t, userNames, "Alex")
	assert.NotContains(t, userNames, "Bob")
	assert.NotContains(t, userNames, "Charlie")
	assert.NotContains(t, userNames, "David")
}

func TestODataFilterBoolean(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create users with different accountEnabled status
	users := []struct {
		name           string
		accountEnabled bool
	}{
		{"Enabled User 1", true},
		{"Enabled User 2", true},
		{"Disabled User 1", false},
		{"Disabled User 2", false},
		{"Enabled User 3", true},
	}

	for i, u := range users {
		accountEnabled := u.accountEnabled
		user := model.User{
			DisplayName:       u.name,
			UserPrincipalName: fmt.Sprintf("user%d@example.com", i+1),
			Mail:              fmt.Sprintf("user%d@example.com", i+1),
			AccountEnabled:    &accountEnabled,
		}
		_, err := store.CreateUser(ctx, user)
		require.NoError(t, err)
	}

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Filter enabled users
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$filter=accountEnabled%20eq%20true", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	usersResp := listResp["value"].([]any)
	assert.Len(t, usersResp, 3) // 3 enabled users
}

func TestODataOrderBy(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create users with different names out of order
	names := []string{"Charlie", "Alice", "Bob"}
	departments := []string{"Engineering", "Sales", "Marketing"}
	for i, name := range names {
		accountEnabled := true
		user := model.User{
			DisplayName:       name,
			UserPrincipalName: fmt.Sprintf("%s@example.com", strings.ToLower(name)),
			Mail:              fmt.Sprintf("%s@example.com", strings.ToLower(name)),
			Department:        departments[i],
			AccountEnabled:    &accountEnabled,
		}
		_, err := store.CreateUser(ctx, user)
		require.NoError(t, err)
	}

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Order by displayName
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$orderby=displayName", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	users := listResp["value"].([]any)
	require.Len(t, users, 3)

	// Verify alphabetical order: Alice, Bob, Charlie
	assert.Equal(t, "Alice", users[0].(map[string]any)["displayName"])
	assert.Equal(t, "Bob", users[1].(map[string]any)["displayName"])
	assert.Equal(t, "Charlie", users[2].(map[string]any)["displayName"])
}

func TestODataSearch(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create users with different names
	names := []string{"Alice Smith", "Bob Johnson", "Alice Cooper", "David Aliceson"}
	for i, name := range names {
		accountEnabled := true
		user := model.User{
			DisplayName:       name,
			UserPrincipalName: fmt.Sprintf("user%d@example.com", i+1),
			Mail:              fmt.Sprintf("user%d@example.com", i+1),
			AccountEnabled:    &accountEnabled,
		}
		_, err := store.CreateUser(ctx, user)
		require.NoError(t, err)
	}

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Search for "Ali" (case-insensitive)
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$search=Ali", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	users := listResp["value"].([]any)
	// Should match: Alice Smith, Alice Cooper, David Aliceson (contains "Alice")
	assert.Len(t, users, 3)

	// Verify correct users
	userNames := make([]string, len(users))
	for i, u := range users {
		userMap := u.(map[string]any)
		userNames[i] = userMap["displayName"].(string)
	}

	assert.Contains(t, userNames, "Alice Smith")
	assert.Contains(t, userNames, "Alice Cooper")
	assert.Contains(t, userNames, "David Aliceson")
	assert.NotContains(t, userNames, "Bob Johnson")
}

func TestODataInvalidFilter(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Send invalid filter syntax - using URL encoding for special characters
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$filter=invalid%20syntax%20%25%25", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should return 400 Bad Request, not 500 Internal Server Error
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Verify error response
	var errorResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &errorResp)
	require.NoError(t, err)

	assert.Contains(t, errorResp, "error")
	errorObj := errorResp["error"].(map[string]any)
	assert.Contains(t, errorObj, "code")
	assert.Contains(t, errorObj, "message")
	assert.Equal(t, "InvalidRequest", errorObj["code"])
}

func TestListUsersExpandManager(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create different users
	accountEnabled := true
	manager := model.User{
		DisplayName:       "Manager User",
		UserPrincipalName: "manager@example.com",
		Mail:              "manager@example.com",
		Department:        "Engineering",
		JobTitle:          "Director",
		AccountEnabled:    &accountEnabled,
	}
	_, err := store.CreateUser(ctx, manager)
	require.NoError(t, err)

	solo := model.User{
		DisplayName:       "Solo User",
		UserPrincipalName: "solo@example.com",
		Mail:              "solo@example.com",
		Department:        "Engineering",
		JobTitle:          "Engineer",
		AccountEnabled:    &accountEnabled,
	}
	_, err = store.CreateUser(ctx, solo)
	require.NoError(t, err)

	// Create a proper chain of management
	topManager := model.User{
		DisplayName:       "Top Manager",
		UserPrincipalName: "topmanager@example.com",
		Mail:              "topmanager@example.com",
		Department:        "Engineering",
		JobTitle:          "VP",
		AccountEnabled:    &accountEnabled,
	}
	createdTopManager, err := store.CreateUser(ctx, topManager)
	require.NoError(t, err)

	employee := model.User{
		DisplayName:       "Employee User",
		UserPrincipalName: "employee@example.com",
		Mail:              "employee@example.com",
		Department:        "Engineering",
		JobTitle:          "Developer",
		AccountEnabled:    &accountEnabled,
	}
	createdEmployee, err := store.CreateUser(ctx, employee)
	require.NoError(t, err)

	// Set topManager as manager of employee
	err = store.SetManager(ctx, createdEmployee.ID, createdTopManager.ID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Request with $expand=manager
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$expand=manager", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	users := listResp["value"].([]any)
	require.Len(t, users, 4)

	// Find employee and verify manager
	for _, u := range users {
		userMap := u.(map[string]any)
		if userMap["userPrincipalName"] == "employee@example.com" {
			// Employee should have a manager
			assert.Contains(t, userMap, "manager", "employee should have manager key")
			mgrMap, ok := userMap["manager"].(map[string]any)
			require.True(t, ok, "manager should be a map")
			assert.Equal(t, createdTopManager.ID, mgrMap["id"])
			assert.Equal(t, "Top Manager", mgrMap["displayName"])
			assert.Equal(t, "topmanager@example.com", mgrMap["userPrincipalName"])
		}
		if userMap["userPrincipalName"] == "topmanager@example.com" {
			// Top manager has no manager set - should NOT have manager key
			_, hasManager := userMap["manager"]
			assert.False(t, hasManager, "top manager should not have manager key at all")
		}
	}
}

func TestListUsersExpandManagerWithSelect(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	accountEnabled := true
	manager := model.User{
		DisplayName:       "Manager User",
		UserPrincipalName: "manager@example.com",
		Mail:              "manager@example.com",
		Department:        "Engineering",
		JobTitle:          "Director",
		AccountEnabled:    &accountEnabled,
	}
	createdManager, err := store.CreateUser(ctx, manager)
	require.NoError(t, err)

	employee := model.User{
		DisplayName:       "Employee User",
		UserPrincipalName: "employee@example.com",
		Mail:              "employee@example.com",
		Department:        "Engineering",
		JobTitle:          "Developer",
		AccountEnabled:    &accountEnabled,
	}
	createdEmployee, err := store.CreateUser(ctx, employee)
	require.NoError(t, err)

	err = store.SetManager(ctx, createdEmployee.ID, createdManager.ID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Request with $expand=manager($select=userPrincipalName)
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$expand=manager($select=userPrincipalName)", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	users := listResp["value"].([]any)
	require.Len(t, users, 2)

	// Find employee and verify manager has only selected fields
	for _, u := range users {
		userMap := u.(map[string]any)
		if userMap["userPrincipalName"] == "employee@example.com" {
			mgrMap, ok := userMap["manager"].(map[string]any)
			require.True(t, ok, "manager should be a map")
			// Should have userPrincipalName, id, @odata.type
			assert.Contains(t, mgrMap, "userPrincipalName", "manager should have userPrincipalName from $select")
			assert.Contains(t, mgrMap, "id", "manager should always have id")
			assert.Contains(t, mgrMap, "@odata.type", "manager should always have @odata.type")
			// Should NOT have displayName, mail, department, jobTitle
			assert.NotContains(t, mgrMap, "displayName", "manager should not have displayName when not selected")
			assert.NotContains(t, mgrMap, "mail", "manager should not have mail when not selected")
			assert.NotContains(t, mgrMap, "department", "manager should not have department when not selected")
			assert.NotContains(t, mgrMap, "jobTitle", "manager should not have jobTitle when not selected")
		}
	}
}

func TestGetUserExpandManager(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	accountEnabled := true
	manager := model.User{
		DisplayName:       "Manager User",
		UserPrincipalName: "manager@example.com",
		Mail:              "manager@example.com",
		Department:        "Engineering",
		JobTitle:          "Director",
		AccountEnabled:    &accountEnabled,
	}
	createdManager, err := store.CreateUser(ctx, manager)
	require.NoError(t, err)

	employee := model.User{
		DisplayName:       "Employee User",
		UserPrincipalName: "employee@example.com",
		Mail:              "employee@example.com",
		Department:        "Engineering",
		JobTitle:          "Developer",
		AccountEnabled:    &accountEnabled,
	}
	createdEmployee, err := store.CreateUser(ctx, employee)
	require.NoError(t, err)

	err = store.SetManager(ctx, createdEmployee.ID, createdManager.ID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Request single user with $expand=manager
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users/"+createdEmployee.ID+"?$expand=manager", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var userResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &userResp)
	require.NoError(t, err)

	// Should have manager object
	assert.Contains(t, userResp, "manager", "response should have manager key")
	mgrMap, ok := userResp["manager"].(map[string]any)
	require.True(t, ok, "manager should be a map")
	assert.Equal(t, createdManager.ID, mgrMap["id"])
	assert.Equal(t, "Manager User", mgrMap["displayName"])
	assert.Equal(t, "manager@example.com", mgrMap["userPrincipalName"])
	// Should include @odata.type
	assert.Equal(t, "#microsoft.graph.user", mgrMap["@odata.type"])
}

func TestGetUserExpandManagerByUPN(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	accountEnabled := true
	manager := model.User{
		DisplayName:       "Manager User",
		UserPrincipalName: "manager@example.com",
		Mail:              "manager@example.com",
		Department:        "Engineering",
		JobTitle:          "Director",
		AccountEnabled:    &accountEnabled,
	}
	createdManager, err := store.CreateUser(ctx, manager)
	require.NoError(t, err)

	employee := model.User{
		DisplayName:       "Employee User",
		UserPrincipalName: "employee@example.com",
		Mail:              "employee@example.com",
		Department:        "Engineering",
		JobTitle:          "Developer",
		AccountEnabled:    &accountEnabled,
	}
	createdEmployee, err := store.CreateUser(ctx, employee)
	require.NoError(t, err)

	err = store.SetManager(ctx, createdEmployee.ID, createdManager.ID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Request single user by UPN with $expand=manager
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users/employee@example.com?$expand=manager", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var userResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &userResp)
	require.NoError(t, err)

	// Should have manager object when fetched by UPN
	assert.Contains(t, userResp, "manager", "response should have manager key when fetched by UPN")
	mgrMap, ok := userResp["manager"].(map[string]any)
	require.True(t, ok, "manager should be a map")
	assert.Equal(t, createdManager.ID, mgrMap["id"])
	assert.Equal(t, "manager@example.com", mgrMap["userPrincipalName"])
	assert.Equal(t, "#microsoft.graph.user", mgrMap["@odata.type"])
}

func TestGetUserExpandManagerWithSelect(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	accountEnabled := true
	manager := model.User{
		DisplayName:       "Manager User",
		UserPrincipalName: "manager@example.com",
		Mail:              "manager@example.com",
		Department:        "Engineering",
		JobTitle:          "Director",
		AccountEnabled:    &accountEnabled,
	}
	createdManager, err := store.CreateUser(ctx, manager)
	require.NoError(t, err)

	employee := model.User{
		DisplayName:       "Employee User",
		UserPrincipalName: "employee@example.com",
		Mail:              "employee@example.com",
		Department:        "Engineering",
		JobTitle:          "Developer",
		AccountEnabled:    &accountEnabled,
	}
	createdEmployee, err := store.CreateUser(ctx, employee)
	require.NoError(t, err)

	err = store.SetManager(ctx, createdEmployee.ID, createdManager.ID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Request single user with $expand=manager($select=userPrincipalName,displayName)
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users/"+createdEmployee.ID+"?$expand=manager($select=userPrincipalName,displayName)", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var userResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &userResp)
	require.NoError(t, err)

	// Should have manager object
	mgrMap, ok := userResp["manager"].(map[string]any)
	require.True(t, ok, "manager should be a map")
	// Should have id, @odata.type, userPrincipalName, displayName
	assert.Contains(t, mgrMap, "id")
	assert.Contains(t, mgrMap, "@odata.type")
	assert.Contains(t, mgrMap, "userPrincipalName")
	assert.Contains(t, mgrMap, "displayName")
	// Should NOT have mail, department, jobTitle
	assert.NotContains(t, mgrMap, "mail")
	assert.NotContains(t, mgrMap, "department")
	assert.NotContains(t, mgrMap, "jobTitle")
}

func TestListUsersExpandManagerNoManager(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	accountEnabled := true
	user := model.User{
		DisplayName:       "Solo User",
		UserPrincipalName: "solo@example.com",
		Mail:              "solo@example.com",
		AccountEnabled:    &accountEnabled,
	}
	_, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Request with $expand=manager
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$expand=manager", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	users := listResp["value"].([]any)
	require.Len(t, users, 1)

	userMap := users[0].(map[string]any)
	// User without manager should NOT have "manager" key at all
	_, hasManager := userMap["manager"]
	assert.False(t, hasManager, "user without manager should not have manager key in response")
}

func TestListUsersExpandManagerWithUserSelect(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	accountEnabled := true
	manager := model.User{
		DisplayName:       "Manager User",
		UserPrincipalName: "manager@example.com",
		Mail:              "manager@example.com",
		Department:        "Engineering",
		JobTitle:          "Director",
		AccountEnabled:    &accountEnabled,
	}
	createdManager, err := store.CreateUser(ctx, manager)
	require.NoError(t, err)

	employee := model.User{
		DisplayName:       "Employee User",
		UserPrincipalName: "employee@example.com",
		Mail:              "employee@example.com",
		Department:        "Engineering",
		JobTitle:          "Developer",
		AccountEnabled:    &accountEnabled,
	}
	createdEmployee, err := store.CreateUser(ctx, employee)
	require.NoError(t, err)

	err = store.SetManager(ctx, createdEmployee.ID, createdManager.ID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Request with both $select on user and nested $select on manager
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$select=userPrincipalName&$expand=manager($select=userPrincipalName)", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	users := listResp["value"].([]any)
	require.Len(t, users, 2)

	for _, u := range users {
		userMap := u.(map[string]any)
		// User should only have id, userPrincipalName, @odata.context
		assert.Contains(t, userMap, "id")
		assert.Contains(t, userMap, "userPrincipalName")
		// Should NOT have displayName, mail, department, jobTitle on the user itself
		assert.NotContains(t, userMap, "displayName")
		assert.NotContains(t, userMap, "mail")

		// If this is the employee, check the manager
		if userMap["userPrincipalName"] == "employee@example.com" {
			mgrMap, ok := userMap["manager"].(map[string]any)
			require.True(t, ok)
			assert.Contains(t, mgrMap, "id")
			assert.Contains(t, mgrMap, "userPrincipalName")
			assert.Contains(t, mgrMap, "@odata.type")
			assert.NotContains(t, mgrMap, "displayName")
			assert.NotContains(t, mgrMap, "mail")
		}
	}
}

func TestListUsersExpandDirectReports(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	// Create a manager and 2 direct reports
	accountEnabled := true
	manager := model.User{
		DisplayName:       "Manager User",
		UserPrincipalName: "manager@example.com",
		Mail:              "manager@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdManager, err := store.CreateUser(ctx, manager)
	require.NoError(t, err)

	report1 := model.User{
		DisplayName:       "Report One",
		UserPrincipalName: "report1@example.com",
		Mail:              "report1@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdReport1, err := store.CreateUser(ctx, report1)
	require.NoError(t, err)

	report2 := model.User{
		DisplayName:       "Report Two",
		UserPrincipalName: "report2@example.com",
		Mail:              "report2@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdReport2, err := store.CreateUser(ctx, report2)
	require.NoError(t, err)

	// Set manager for both reports
	err = store.SetManager(ctx, createdReport1.ID, createdManager.ID)
	require.NoError(t, err)
	err = store.SetManager(ctx, createdReport2.ID, createdManager.ID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$expand=directReports", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	users := listResp["value"].([]any)
	// Find the manager and verify directReports
	for _, u := range users {
		userMap := u.(map[string]any)
		if userMap["userPrincipalName"] == "manager@example.com" {
			dr, ok := userMap["directReports"].([]any)
			require.True(t, ok, "manager should have directReports key")
			assert.Len(t, dr, 2, "manager should have 2 direct reports")
		}
		if userMap["userPrincipalName"] == "report1@example.com" || userMap["userPrincipalName"] == "report2@example.com" {
			// Reports should have empty or absent directReports
			drRaw, hasDR := userMap["directReports"]
			if hasDR {
				dr, ok := drRaw.([]any)
				if ok {
					assert.Len(t, dr, 0, "reports should have empty directReports")
				}
			}
		}
	}
}

func TestListUsersExpandMemberOf(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "testuser@example.com",
		Mail:              "testuser@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	mailEnabled := false
	securityEnabled := true
	group1 := model.Group{
		DisplayName:     "Group 1",
		MailEnabled:     &mailEnabled,
		SecurityEnabled: &securityEnabled,
		MailNickname:    "group1",
	}
	createdGroup1, err := store.CreateGroup(ctx, group1)
	require.NoError(t, err)

	group2 := model.Group{
		DisplayName:     "Group 2",
		MailEnabled:     &mailEnabled,
		SecurityEnabled: &securityEnabled,
		MailNickname:    "group2",
	}
	createdGroup2, err := store.CreateGroup(ctx, group2)
	require.NoError(t, err)

	err = store.AddMember(ctx, createdGroup1.ID, createdUser.ID, "user")
	require.NoError(t, err)
	err = store.AddMember(ctx, createdGroup2.ID, createdUser.ID, "user")
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All", "GroupMember.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$expand=memberOf", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	users := listResp["value"].([]any)
	for _, u := range users {
		userMap := u.(map[string]any)
		if userMap["id"] == createdUser.ID {
			memberOf, ok := userMap["memberOf"].([]any)
			require.True(t, ok, "user should have memberOf key")
			assert.Len(t, memberOf, 2, "user should be member of 2 groups")
		}
	}
}

func TestListUsersExpandDirectReportsWithSelect(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	accountEnabled := true
	manager := model.User{
		DisplayName:       "Manager User",
		UserPrincipalName: "manager@example.com",
		Mail:              "manager@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdManager, err := store.CreateUser(ctx, manager)
	require.NoError(t, err)

	report := model.User{
		DisplayName:       "Direct Report",
		UserPrincipalName: "report@example.com",
		Mail:              "report@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdReport, err := store.CreateUser(ctx, report)
	require.NoError(t, err)

	err = store.SetManager(ctx, createdReport.ID, createdManager.ID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$expand=directReports($select=id,displayName)", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	users := listResp["value"].([]any)
	for _, u := range users {
		userMap := u.(map[string]any)
		if userMap["id"] == createdManager.ID {
			dr, ok := userMap["directReports"].([]any)
			require.True(t, ok, "manager should have directReports key")
			require.Len(t, dr, 1, "manager should have 1 direct report")
			drMap := dr[0].(map[string]any)
			assert.Contains(t, drMap, "id")
			assert.Contains(t, drMap, "displayName")
			assert.Contains(t, drMap, "@odata.type")
			assert.NotContains(t, drMap, "mail")
			assert.NotContains(t, drMap, "userPrincipalName")
		}
	}
}

func TestGetUserExpandDirectReports(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	accountEnabled := true
	manager := model.User{
		DisplayName:       "Manager User",
		UserPrincipalName: "manager@example.com",
		Mail:              "manager@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdManager, err := store.CreateUser(ctx, manager)
	require.NoError(t, err)

	report := model.User{
		DisplayName:       "Direct Report",
		UserPrincipalName: "report@example.com",
		Mail:              "report@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdReport, err := store.CreateUser(ctx, report)
	require.NoError(t, err)

	// Set manager as the report's manager
	err = store.SetManager(ctx, createdReport.ID, createdManager.ID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	// Request the manager with $expand=directReports
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users/"+createdManager.ID+"?$expand=directReports", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var userResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &userResp)
	require.NoError(t, err)

	directReports, ok := userResp["directReports"].([]any)
	require.True(t, ok, "user should have directReports key when expanded")
	assert.Len(t, directReports, 1)
	dr := directReports[0].(map[string]any)
	assert.Equal(t, createdReport.ID, dr["id"])
	assert.Contains(t, dr, "@odata.type")
}

func TestGetUserExpandMemberOf(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "testuser@example.com",
		Mail:              "testuser@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	mailEnabled := false
	securityEnabled := true
	group := model.Group{
		DisplayName:     "Test Group",
		MailEnabled:     &mailEnabled,
		SecurityEnabled: &securityEnabled,
		MailNickname:    "testgroup",
	}
	createdGroup, err := store.CreateGroup(ctx, group)
	require.NoError(t, err)

	err = store.AddMember(ctx, createdGroup.ID, createdUser.ID, "user")
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All", "GroupMember.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL+"/v1.0/users/"+createdUser.ID+"?$expand=memberOf", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var userResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &userResp)
	require.NoError(t, err)

	memberOf, ok := userResp["memberOf"].([]any)
	require.True(t, ok, "user should have memberOf key when expanded")
	assert.Len(t, memberOf, 1)
	groupEntry := memberOf[0].(map[string]any)
	assert.Equal(t, createdGroup.ID, groupEntry["id"])
	assert.Contains(t, groupEntry, "@odata.type")
}

func TestListUsersExpandMemberOfWithSelect(t *testing.T) {
	store := store.NewMemoryStore()
	router := NewRouter(store)
	ctx := context.Background()

	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "testuser@example.com",
		Mail:              "testuser@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	mailEnabled := false
	securityEnabled := true
	group := model.Group{
		DisplayName:     "Test Group",
		MailEnabled:     &mailEnabled,
		SecurityEnabled: &securityEnabled,
		MailNickname:    "testgroup",
	}
	createdGroup, err := store.CreateGroup(ctx, group)
	require.NoError(t, err)

	err = store.AddMember(ctx, createdGroup.ID, createdUser.ID, "user")
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All", "GroupMember.Read.All"}, []string{"User"}, time.Hour, "", "")
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$expand=memberOf($select=id,displayName)", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	users := listResp["value"].([]any)
	for _, u := range users {
		userMap := u.(map[string]any)
		if userMap["id"] == createdUser.ID {
			memberOf, ok := userMap["memberOf"].([]any)
			require.True(t, ok, "user should have memberOf key")
			require.Len(t, memberOf, 1, "user should be member of 1 group")
			groupMap := memberOf[0].(map[string]any)
			assert.Contains(t, groupMap, "id")
			assert.Contains(t, groupMap, "displayName")
			assert.Contains(t, groupMap, "@odata.type")
			assert.NotContains(t, groupMap, "mailNickname")
			assert.NotContains(t, groupMap, "mail")
		}
	}
}
