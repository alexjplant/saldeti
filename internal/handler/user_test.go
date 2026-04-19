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

	"github.com/saldeti/saldeti/internal/auth"
	"github.com/saldeti/saldeti/internal/model"
	"github.com/saldeti/saldeti/internal/store"
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
	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour)
	require.NoError(t, err)

	// Test listing all users
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	assert.Equal(t, "https://graph.microsoft.com/v1.0/$metadata#users", listResp["@odata.context"])
	users := listResp["value"].([]interface{})
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

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour)
	require.NoError(t, err)

	// Filter by department
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$filter=department%20eq%20'Engineering'", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	users := listResp["value"].([]interface{})
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

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour)
	require.NoError(t, err)

	// Select specific fields
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$select=displayName,id,department", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	users := listResp["value"].([]interface{})
	require.Len(t, users, 1)
	
	userMap := users[0].(map[string]interface{})
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

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour)
	require.NoError(t, err)

	// Get only 5 users
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$top=5", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	users := listResp["value"].([]interface{})
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

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour)
	require.NoError(t, err)

	// Request with count
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$count=true", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
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

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour)
	require.NoError(t, err)

	// Get user by ID
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users/"+createdUser.ID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var userResp map[string]interface{}
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

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour)
	require.NoError(t, err)

	// Get user by UPN
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users/test@example.com", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var userResp map[string]interface{}
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

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour)
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

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.ReadWrite.All"}, []string{"User"}, time.Hour)
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

	var userResp map[string]interface{}
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

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.ReadWrite.All"}, []string{"User"}, time.Hour)
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

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.ReadWrite.All"}, []string{"User"}, time.Hour)
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

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.ReadWrite.All"}, []string{"User"}, time.Hour)
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

	var userResp map[string]interface{}
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

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.ReadWrite.All"}, []string{"User"}, time.Hour)
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

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour)
	require.NoError(t, err)

	// Filter names starting with 'A'
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$filter=startswith(displayName,'A')", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	users := listResp["value"].([]interface{})
	assert.Len(t, users, 2) // Alice and Alex
	
	// Verify correct users
	userNames := make([]string, len(users))
	for i, u := range users {
		userMap := u.(map[string]interface{})
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
		name            string
		accountEnabled  bool
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

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour)
	require.NoError(t, err)

	// Filter enabled users
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$filter=accountEnabled%20eq%20true", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	usersResp := listResp["value"].([]interface{})
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

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour)
	require.NoError(t, err)

	// Order by displayName
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$orderby=displayName", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	users := listResp["value"].([]interface{})
	require.Len(t, users, 3)
	
	// Verify alphabetical order: Alice, Bob, Charlie
	assert.Equal(t, "Alice", users[0].(map[string]interface{})["displayName"])
	assert.Equal(t, "Bob", users[1].(map[string]interface{})["displayName"])
	assert.Equal(t, "Charlie", users[2].(map[string]interface{})["displayName"])
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

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour)
	require.NoError(t, err)

	// Search for "Ali" (case-insensitive)
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users?$search=Ali", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	users := listResp["value"].([]interface{})
	// Should match: Alice Smith, Alice Cooper, David Aliceson (contains "Alice")
	assert.Len(t, users, 3)
	
	// Verify correct users
	userNames := make([]string, len(users))
	for i, u := range users {
		userMap := u.(map[string]interface{})
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

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour)
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
	var errorResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &errorResp)
	require.NoError(t, err)
	
	assert.Contains(t, errorResp, "error")
	errorObj := errorResp["error"].(map[string]interface{})
	assert.Contains(t, errorObj, "code")
	assert.Contains(t, errorObj, "message")
	assert.Equal(t, "InvalidRequest", errorObj["code"])
}