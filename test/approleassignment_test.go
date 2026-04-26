//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createAppWithRole creates an application with an app role via raw HTTP.
// Returns the auto-created SP ID and the app role ID.
func createAppWithRole(t *testing.T, ts *TestServer, appDisplayName, roleValue string) (spID string, roleID string) {
	t.Helper()

	// Create application with app role via raw HTTP
	roleID = uuid.New().String()
	appBody := map[string]interface{}{
		"displayName": appDisplayName,
		"appRoles": []interface{}{
			map[string]interface{}{
				"id":                 roleID,
				"allowedMemberTypes": []string{"User", "Application"},
				"description":        "Test app role",
				"displayName":        "Test Role",
				"isEnabled":          true,
				"origin":             "Application",
				"value":              roleValue,
			},
		},
	}
	appJSON, err := json.Marshal(appBody)
	require.NoError(t, err)

	resp := authedPost(t, ts, ts.BaseURL+"/v1.0/applications", appJSON)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	appResp := readJSON(t, resp)
	appId := appResp["appId"].(string)

	// Get the auto-created service principal
	sp := getAutoCreatedSP(t, ts, appId)
	spID = *sp.GetId()

	return spID, roleID
}

// TestAppRoleAssignment_CreateAndGetForUser tests creating an app role assignment for a user
func TestAppRoleAssignment_CreateAndGetForUser(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	// 1. Create a user
	user := createTestUserSDK(t, tss, "Role User", "roleuser@saldeti.local")
	userID := *user.GetId()

	// 2. Create an application with an app role (auto-creates SP)
	spID, roleID := createAppWithRole(t, tss, "Role Test App", "Test.Role")

	// 3. Create app role assignment via raw HTTP POST to /users/{id}/appRoleAssignments
	assignBody := map[string]interface{}{
		"principalId": userID,
		"resourceId":  spID,
		"appRoleId":   roleID,
	}
	assignJSON, err := json.Marshal(assignBody)
	require.NoError(t, err)

	resp := authedPost(t, tss,
		fmt.Sprintf("%s/v1.0/users/%s/appRoleAssignments", tss.BaseURL, userID),
		assignJSON,
	)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	assignResp := readJSON(t, resp)
	assignmentID := assignResp["id"].(string)
	assert.NotEmpty(t, assignmentID)
	assert.Equal(t, roleID, assignResp["appRoleId"])
	assert.Equal(t, userID, assignResp["principalId"])
	assert.Equal(t, spID, assignResp["resourceId"])

	// 4. List app role assignments for the user
	listResp := authedGet(t, tss,
		fmt.Sprintf("%s/v1.0/users/%s/appRoleAssignments", tss.BaseURL, userID),
	)
	defer listResp.Body.Close()
	require.Equal(t, http.StatusOK, listResp.StatusCode)

	listResult := readJSON(t, listResp)
	values := listResult["value"].([]interface{})
	assert.GreaterOrEqual(t, len(values), 1, "Expected at least 1 app role assignment")

	// 5. Delete the assignment
	deleteResp := authedDelete(t, tss,
		fmt.Sprintf("%s/v1.0/users/%s/appRoleAssignments/%s", tss.BaseURL, userID, assignmentID),
	)
	defer deleteResp.Body.Close()
	assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

	// 6. Verify assignment is gone
	listResp2 := authedGet(t, tss,
		fmt.Sprintf("%s/v1.0/users/%s/appRoleAssignments", tss.BaseURL, userID),
	)
	defer listResp2.Body.Close()
	listResult2 := readJSON(t, listResp2)
	values2 := listResult2["value"].([]interface{})
	assert.Empty(t, values2, "Expected 0 app role assignments after deletion")
}

// TestAppRoleAssignment_SPAppRoleAssignments tests assigning a role TO a service principal
// (the SP is the principal receiving a role from a resource SP)
func TestAppRoleAssignment_SPAppRoleAssignments(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	// 1. Create resource application with an app role (auto-creates resource SP)
	resourceSPID, roleID := createAppWithRole(t, tss, "SP Resource App", "SP.Resource.Role")

	// 2. Create principal application (auto-creates principal SP)
	principalApp := createTestApplicationSDK(t, tss, "SP Principal App")
	principalAppId := *principalApp.GetAppId()
	principalSP := getAutoCreatedSP(t, tss, principalAppId)
	principalSPID := *principalSP.GetId()

	// 3. Assign the role from resource SP to principal SP
	assignBody := map[string]interface{}{
		"principalId": principalSPID,
		"resourceId":  resourceSPID,
		"appRoleId":   roleID,
	}
	assignJSON, err := json.Marshal(assignBody)
	require.NoError(t, err)

	resp := authedPost(t, tss,
		fmt.Sprintf("%s/v1.0/servicePrincipals/%s/appRoleAssignments", tss.BaseURL, principalSPID),
		assignJSON,
	)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	assignResp := readJSON(t, resp)
	assignmentID := assignResp["id"].(string)
	assert.NotEmpty(t, assignmentID)

	// 4. List app role assignments for the principal SP (should show the assignment)
	listResp := authedGet(t, tss,
		fmt.Sprintf("%s/v1.0/servicePrincipals/%s/appRoleAssignments", tss.BaseURL, principalSPID),
	)
	defer listResp.Body.Close()
	require.Equal(t, http.StatusOK, listResp.StatusCode)

	listResult := readJSON(t, listResp)
	values := listResult["value"].([]interface{})
	assert.GreaterOrEqual(t, len(values), 1, "Expected at least 1 app role assignment for principal SP")

	// 5. Also verify via appRoleAssignedTo on the resource SP
	assignedToResp := authedGet(t, tss,
		fmt.Sprintf("%s/v1.0/servicePrincipals/%s/appRoleAssignedTo", tss.BaseURL, resourceSPID),
	)
	defer assignedToResp.Body.Close()
	require.Equal(t, http.StatusOK, assignedToResp.StatusCode)
	assignedToResult := readJSON(t, assignedToResp)
	assignedToValues := assignedToResult["value"].([]interface{})
	assert.GreaterOrEqual(t, len(assignedToValues), 1, "Expected at least 1 appRoleAssignedTo for resource SP")

	// 6. Delete the assignment via SP endpoint
	deleteResp := authedDelete(t, tss,
		fmt.Sprintf("%s/v1.0/servicePrincipals/%s/appRoleAssignments/%s", tss.BaseURL, principalSPID, assignmentID),
	)
	defer deleteResp.Body.Close()
	assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)
}

// TestAppRoleAssignment_AppRoleAssignedTo tests appRoleAssignedTo endpoint
func TestAppRoleAssignment_AppRoleAssignedTo(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	// 1. Create a user
	user := createTestUserSDK(t, tss, "AssignedTo User", "assignedtouser@saldeti.local")
	userID := *user.GetId()

	// 2. Create an application with an app role (auto-creates SP)
	spID, roleID := createAppWithRole(t, tss, "AssignedTo App", "Assigned.Role")

	// 3. Create app role assignment via appRoleAssignedTo endpoint
	assignBody := map[string]interface{}{
		"principalId": userID,
		"resourceId":  spID,
		"appRoleId":   roleID,
	}
	assignJSON, err := json.Marshal(assignBody)
	require.NoError(t, err)

	resp := authedPost(t, tss,
		fmt.Sprintf("%s/v1.0/servicePrincipals/%s/appRoleAssignedTo", tss.BaseURL, spID),
		assignJSON,
	)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// 4. List appRoleAssignedTo
	listResp := authedGet(t, tss,
		fmt.Sprintf("%s/v1.0/servicePrincipals/%s/appRoleAssignedTo", tss.BaseURL, spID),
	)
	defer listResp.Body.Close()
	require.Equal(t, http.StatusOK, listResp.StatusCode)

	listResult := readJSON(t, listResp)
	values := listResult["value"].([]interface{})
	assert.GreaterOrEqual(t, len(values), 1, "Expected at least 1 appRoleAssignedTo")
}

// TestAppRoleAssignment_GroupAssignment tests creating an app role assignment for a group
func TestAppRoleAssignment_GroupAssignment(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	// 1. Create a group
	group := createTestGroupSDK(t, tss, "Role Group")
	groupID := *group.GetId()

	// 2. Create an application with an app role (auto-creates SP)
	spID, roleID := createAppWithRole(t, tss, "Group Role App", "Group.Role")

	// 3. Create app role assignment via group endpoint
	assignBody := map[string]interface{}{
		"principalId": groupID,
		"resourceId":  spID,
		"appRoleId":   roleID,
	}
	assignJSON, err := json.Marshal(assignBody)
	require.NoError(t, err)

	resp := authedPost(t, tss,
		fmt.Sprintf("%s/v1.0/groups/%s/appRoleAssignments", tss.BaseURL, groupID),
		assignJSON,
	)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	assignResp := readJSON(t, resp)
	assert.Equal(t, groupID, assignResp["principalId"])
	assert.Equal(t, spID, assignResp["resourceId"])
	assert.Equal(t, roleID, assignResp["appRoleId"])

	assignmentID := assignResp["id"].(string)

	// 4. List assignments for the group
	listResp := authedGet(t, tss,
		fmt.Sprintf("%s/v1.0/groups/%s/appRoleAssignments", tss.BaseURL, groupID),
	)
	defer listResp.Body.Close()
	require.Equal(t, http.StatusOK, listResp.StatusCode)

	listResult := readJSON(t, listResp)
	values := listResult["value"].([]interface{})
	assert.GreaterOrEqual(t, len(values), 1)

	// 5. Delete the assignment
	deleteResp := authedDelete(t, tss,
		fmt.Sprintf("%s/v1.0/groups/%s/appRoleAssignments/%s", tss.BaseURL, groupID, assignmentID),
	)
	defer deleteResp.Body.Close()
	assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)
}

func TestAppRoleAssignment_DeleteNonExistent(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	// Create a real SP to use in the URL
	app := createTestApplicationSDK(t, tss, "Delete NonExistent App")
	appId := *app.GetAppId()
	sp := getAutoCreatedSP(t, tss, appId)
	spID := *sp.GetId()

	// Generate a fake assignment ID
	fakeAssignmentID := uuid.New().String()

	// DELETE non-existent assignment → 404
	resp := authedDelete(t, tss,
		fmt.Sprintf("%s/v1.0/servicePrincipals/%s/appRoleAssignments/%s", tss.BaseURL, spID, fakeAssignmentID),
	)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
