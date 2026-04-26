package handler

import (
	"context"
	"encoding/json"
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

// createTestAppWithRole creates a test app with an app role and returns (spID, roleID)
func createTestAppWithRole(t *testing.T, st store.Store, ctx context.Context) (string, string) {
	t.Helper()
	roleEnabled := true
	roleID := "00000000-0000-0000-0000-000000000001"
	app := model.Application{
		DisplayName: "Test App With Role",
		AppRoles: []model.AppRole{
			{
				ID:                 roleID,
				AllowedMemberTypes: []string{"User", "Application"},
				Description:        "Test Role",
				DisplayName:        "Test Role",
				IsEnabled:          &roleEnabled,
				Value:              "TestRole",
			},
		},
	}
	createdApp, err := st.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Get the auto-created SP
	sp, err := st.GetServicePrincipalByAppID(ctx, createdApp.AppID)
	require.NoError(t, err)

	return sp.ID, roleID
}

func TestListUserAppRoleAssignments(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create app with role
	spID, roleID := createTestAppWithRole(t, st, ctx)

	// Create user
	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "testuser@example.com",
		Mail:              "testuser@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := st.CreateUser(ctx, user)
	require.NoError(t, err)

	// Create assignment via store
	assignment, err := st.CreateAppRoleAssignment(ctx, spID, createdUser.ID, roleID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"AppRoleAssignment.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// GET /v1.0/users/{userID}/appRoleAssignments
	req, err := http.NewRequest("GET", server.URL+"/v1.0/users/"+createdUser.ID+"/appRoleAssignments", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 200, value has 1 item with correct fields
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	assignments := listResp["value"].([]interface{})
	assert.Len(t, assignments, 1)

	assignmentMap := assignments[0].(map[string]interface{})
	assert.Equal(t, assignment.ID, assignmentMap["id"])
	assert.Equal(t, createdUser.ID, assignmentMap["principalId"])
	assert.Equal(t, spID, assignmentMap["resourceId"])
	assert.Equal(t, roleID, assignmentMap["appRoleId"])
}

func TestCreateUserAppRoleAssignment(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create app with role
	spID, roleID := createTestAppWithRole(t, st, ctx)

	// Create user
	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "testuser@example.com",
		Mail:              "testuser@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := st.CreateUser(ctx, user)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"AppRoleAssignment.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// POST /v1.0/users/{userID}/appRoleAssignments with body
	body := `{"principalId":"` + createdUser.ID + `","resourceId":"` + spID + `","appRoleId":"` + roleID + `"}`
	req, err := http.NewRequest("POST", server.URL+"/v1.0/users/"+createdUser.ID+"/appRoleAssignments", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 201, response has id, principalId == userID, resourceId == spID, appRoleId == roleID
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var assignmentResp map[string]interface{}
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(respBody, &assignmentResp)
	require.NoError(t, err)

	assert.Contains(t, assignmentResp, "id")
	assert.NotEmpty(t, assignmentResp["id"])
	assert.Equal(t, createdUser.ID, assignmentResp["principalId"])
	assert.Equal(t, spID, assignmentResp["resourceId"])
	assert.Equal(t, roleID, assignmentResp["appRoleId"])
}

func TestDeleteUserAppRoleAssignment(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create app with role
	spID, roleID := createTestAppWithRole(t, st, ctx)

	// Create user
	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "testuser@example.com",
		Mail:              "testuser@example.com",
		AccountEnabled:    &accountEnabled,
	}
	createdUser, err := st.CreateUser(ctx, user)
	require.NoError(t, err)

	// Create assignment
	assignment, err := st.CreateAppRoleAssignment(ctx, spID, createdUser.ID, roleID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"AppRoleAssignment.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// DELETE /v1.0/users/{userID}/appRoleAssignments/{assignment.ID}
	req, err := http.NewRequest("DELETE", server.URL+"/v1.0/users/"+createdUser.ID+"/appRoleAssignments/"+assignment.ID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 204
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestListGroupAppRoleAssignments(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create app with role
	spID, roleID := createTestAppWithRole(t, st, ctx)

	// Create group
	mailEnabled := false
	securityEnabled := true
	group := model.Group{
		DisplayName:       "Test Group",
		MailNickname:      "testgroup",
		MailEnabled:       &mailEnabled,
		SecurityEnabled:   &securityEnabled,
		GroupTypes:        []string{"Unified"},
	}
	createdGroup, err := st.CreateGroup(ctx, group)
	require.NoError(t, err)

	// Create assignment via store
	assignment, err := st.CreateAppRoleAssignment(ctx, spID, createdGroup.ID, roleID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"AppRoleAssignment.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// GET /v1.0/groups/{groupID}/appRoleAssignments
	req, err := http.NewRequest("GET", server.URL+"/v1.0/groups/"+createdGroup.ID+"/appRoleAssignments", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 200, value has 1 item
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &listResp)
	require.NoError(t, err)

	assignments := listResp["value"].([]interface{})
	assert.Len(t, assignments, 1)

	assignmentMap := assignments[0].(map[string]interface{})
	assert.Equal(t, assignment.ID, assignmentMap["id"])
	assert.Equal(t, createdGroup.ID, assignmentMap["principalId"])
	assert.Equal(t, spID, assignmentMap["resourceId"])
	assert.Equal(t, roleID, assignmentMap["appRoleId"])
}

func TestCreateGroupAppRoleAssignment(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create app with role
	spID, roleID := createTestAppWithRole(t, st, ctx)

	// Create group
	mailEnabled := false
	securityEnabled := true
	group := model.Group{
		DisplayName:       "Test Group",
		MailNickname:      "testgroup",
		MailEnabled:       &mailEnabled,
		SecurityEnabled:   &securityEnabled,
		GroupTypes:        []string{"Unified"},
	}
	createdGroup, err := st.CreateGroup(ctx, group)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"AppRoleAssignment.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// POST /v1.0/groups/{groupID}/appRoleAssignments with body
	body := `{"principalId":"` + createdGroup.ID + `","resourceId":"` + spID + `","appRoleId":"` + roleID + `"}`
	req, err := http.NewRequest("POST", server.URL+"/v1.0/groups/"+createdGroup.ID+"/appRoleAssignments", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 201
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var assignmentResp map[string]interface{}
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(respBody, &assignmentResp)
	require.NoError(t, err)

	assert.Contains(t, assignmentResp, "id")
	assert.NotEmpty(t, assignmentResp["id"])
	assert.Equal(t, createdGroup.ID, assignmentResp["principalId"])
	assert.Equal(t, spID, assignmentResp["resourceId"])
	assert.Equal(t, roleID, assignmentResp["appRoleId"])
}

func TestDeleteGroupAppRoleAssignment(t *testing.T) {
	st := store.NewMemoryStore()
	router := NewRouter(st)
	ctx := context.Background()

	// Create app with role
	spID, roleID := createTestAppWithRole(t, st, ctx)

	// Create group
	mailEnabled := false
	securityEnabled := true
	group := model.Group{
		DisplayName:       "Test Group",
		MailNickname:      "testgroup",
		MailEnabled:       &mailEnabled,
		SecurityEnabled:   &securityEnabled,
		GroupTypes:        []string{"Unified"},
	}
	createdGroup, err := st.CreateGroup(ctx, group)
	require.NoError(t, err)

	// Create assignment
	assignment, err := st.CreateAppRoleAssignment(ctx, spID, createdGroup.ID, roleID)
	require.NoError(t, err)

	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"AppRoleAssignment.ReadWrite.All"}, []string{"Application"}, time.Hour, "", "")
	require.NoError(t, err)

	// DELETE /v1.0/groups/{groupID}/appRoleAssignments/{assignment.ID}
	req, err := http.NewRequest("DELETE", server.URL+"/v1.0/groups/"+createdGroup.ID+"/appRoleAssignments/"+assignment.ID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 204
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
