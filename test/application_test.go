//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/microsoftgraph/msgraph-sdk-go/applications"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestApplication_CRUDLifecycle tests creating, reading, updating, and deleting an application
func TestApplication_CRUDLifecycle(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// 1. Create application via SDK
	newApp := models.NewApplication()
	displayName := "Test Application CRUD"
	newApp.SetDisplayName(&displayName)

	created, err := tss.SDKClient.Applications().Post(ctx, newApp, nil)
	require.NoError(t, err, "Failed to create application")
	require.NotNil(t, created)

	appID := *created.GetId()
	assert.NotEmpty(t, appID, "Application should have an id")
	assert.Equal(t, "Test Application CRUD", *created.GetDisplayName())

	// Verify appId was auto-generated
	assert.NotEmpty(t, created.GetAppId(), "Application should have an auto-generated appId")

	// 2. GET the application by ID via SDK
	app, err := tss.SDKClient.Applications().ByApplicationId(appID).Get(ctx, nil)
	require.NoError(t, err, "Failed to get application")
	assert.Equal(t, "Test Application CRUD", *app.GetDisplayName())
	assert.Equal(t, appID, *app.GetId())

	// 3. PATCH the application (change displayName)
	patch := models.NewApplication()
	updatedName := "Updated Application"
	patch.SetDisplayName(&updatedName)

	updated, err := tss.SDKClient.Applications().ByApplicationId(appID).Patch(ctx, patch, nil)
	require.NoError(t, err, "Failed to patch application")
	assert.Equal(t, "Updated Application", *updated.GetDisplayName())

	// 4. GET again to verify update
	verifyApp, err := tss.SDKClient.Applications().ByApplicationId(appID).Get(ctx, nil)
	require.NoError(t, err, "Failed to get updated application")
	assert.Equal(t, "Updated Application", *verifyApp.GetDisplayName())

	// 5. DELETE the application
	err = tss.SDKClient.Applications().ByApplicationId(appID).Delete(ctx, nil)
	require.NoError(t, err, "Failed to delete application")

	// 6. GET again, expect error (404)
	_, err = tss.SDKClient.Applications().ByApplicationId(appID).Get(ctx, nil)
	require.Error(t, err, "Expected error getting deleted application")
}

// TestApplication_ListApplications tests listing applications
func TestApplication_ListApplications(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// Create 3 applications via SDK
	for i := 1; i <= 3; i++ {
		displayName := fmt.Sprintf("List Test App %d", i)
		createTestApplicationSDK(t, tss, displayName)
	}

	// List applications via SDK
	result, err := tss.SDKClient.Applications().Get(ctx, nil)
	require.NoError(t, err, "Failed to list applications")

	appList := result.GetValue()
	assert.GreaterOrEqual(t, len(appList), 3, "Expected at least 3 applications")

	// Verify structure of first application
	firstApp := appList[0]
	assert.NotNil(t, firstApp.GetId(), "Application missing id field")
	assert.NotNil(t, firstApp.GetDisplayName(), "Application missing displayName field")
}

// TestApplication_AddRemovePassword tests adding and removing a password credential
func TestApplication_AddRemovePassword(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// 1. Create application
	app := createTestApplicationSDK(t, tss, "Password Test App")
	appID := *app.GetId()

	// 2. Add password
	pwBody := applications.NewItemAddPasswordPostRequestBody()
	pwCred := models.NewPasswordCredential()
	pwDisplayName := "Test Password"
	pwCred.SetDisplayName(&pwDisplayName)
	pwBody.SetPasswordCredential(pwCred)

	addedCred, err := tss.SDKClient.Applications().ByApplicationId(appID).AddPassword().Post(ctx, pwBody, nil)
	require.NoError(t, err, "Failed to add password")
	require.NotNil(t, addedCred)

	// Verify returned credential has a secretText and keyId
	assert.NotNil(t, addedCred.GetSecretText(), "Password credential should have secretText")
	assert.NotNil(t, addedCred.GetKeyId(), "Password credential should have keyId")

	keyIDPtr := addedCred.GetKeyId()
	keyID := *keyIDPtr

	// 3. Verify password appears on application
	updatedApp, err := tss.SDKClient.Applications().ByApplicationId(appID).Get(ctx, nil)
	require.NoError(t, err, "Failed to get application after adding password")
	pwCreds := updatedApp.GetPasswordCredentials()
	require.NotNil(t, pwCreds, "Password credentials should not be nil")
	found := false
	for _, pc := range pwCreds {
		if pc.GetKeyId() != nil && *pc.GetKeyId() == keyID {
			found = true
			break
		}
	}
	assert.True(t, found, "Added password credential should appear on application")

	// 4. Remove password
	removeBody := applications.NewItemRemovePasswordPostRequestBody()
	removeBody.SetKeyId(keyIDPtr)

	err = tss.SDKClient.Applications().ByApplicationId(appID).RemovePassword().Post(ctx, removeBody, nil)
	require.NoError(t, err, "Failed to remove password")

	// 5. Verify password removed from application
	verifyApp, err := tss.SDKClient.Applications().ByApplicationId(appID).Get(ctx, nil)
	require.NoError(t, err, "Failed to get application after removing password")
	remainingCreds := verifyApp.GetPasswordCredentials()
	if remainingCreds != nil {
		for _, pc := range remainingCreds {
			if pc.GetKeyId() != nil {
				assert.NotEqual(t, keyID, *pc.GetKeyId(), "Removed password should not appear on application")
			}
		}
	}
}

// TestApplication_Validation tests application creation validation
func TestApplication_Validation(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// 1. Try to create application without displayName
	app := models.NewApplication()
	_, err := tss.SDKClient.Applications().Post(ctx, app, nil)
	require.Error(t, err, "Expected error creating application without displayName")
}

// TestApplication_FilterByName tests filtering applications by displayName
func TestApplication_FilterByName(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	uniqueName := "UniqueFilterApp_" + uuid.New().String()[:8]
	createTestApplicationSDK(t, tss, uniqueName)

	filter := fmt.Sprintf("displayName eq '%s'", uniqueName)
	result, err := tss.SDKClient.Applications().Get(ctx, &applications.ApplicationsRequestBuilderGetRequestConfiguration{
		QueryParameters: &applications.ApplicationsRequestBuilderGetQueryParameters{
			Filter: &filter,
		},
	})
	require.NoError(t, err, "Failed to filter applications")

	appList := result.GetValue()
	require.Len(t, appList, 1, "Expected exactly 1 application matching filter")
	assert.Equal(t, uniqueName, *appList[0].GetDisplayName())
}

// TestApplication_GetByAppID tests getting an application by appId using OData key lookup
func TestApplication_GetByAppID(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	app := createTestApplicationSDK(t, tss, "GetByAppID Test App")
	appAppID := *app.GetAppId()

	url := tss.BaseURL + "/v1.0/applications/(appId=" + appAppID + ")"
	resp := authedGet(t, tss, url)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	result := readJSON(t, resp)
	assert.Equal(t, appAppID, result["appId"])
}

// TestApplication_AddRemoveKey tests adding and removing a key credential
func TestApplication_AddRemoveKey(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	app := createTestApplicationSDK(t, tss, "Key Test App")
	appID := *app.GetId()

	// Add key
	addKeyURL := tss.BaseURL + "/v1.0/applications/" + appID + "/addKey"
	addKeyBody := []byte(`{"displayName":"Test Key","type":"AsymmetricX509Cert","usage":"Verify","key":"dGVzdGtleQ=="}`)
	resp := authedPost(t, tss, addKeyURL, addKeyBody)
	defer resp.Body.Close()

	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated,
		"Expected 200 or 201, got %d", resp.StatusCode)

	result := readJSON(t, resp)
	keyID, ok := result["keyId"].(string)
	require.True(t, ok, "Response should contain keyId")
	assert.NotEmpty(t, keyID)

	// Remove key
	removeKeyURL := tss.BaseURL + "/v1.0/applications/" + appID + "/removeKey"
	removeKeyBody := []byte(fmt.Sprintf(`{"keyId":"%s"}`, keyID))
	resp2 := authedPost(t, tss, removeKeyURL, removeKeyBody)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp2.StatusCode)
}

// TestApplication_Owners tests adding, listing, and removing application owners
func TestApplication_Owners(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	app := createTestApplicationSDK(t, tss, "Owner Test App")
	appID := *app.GetId()

	user := createTestUserSDK(t, tss, "Owner Test User", "owneruser@saldeti.local")
	userID := *user.GetId()

	// Add owner
	addOwnerURL := tss.BaseURL + "/v1.0/applications/" + appID + "/owners/$ref"
	addOwnerBody := []byte(fmt.Sprintf(`{"@odata.id":"https://graph.microsoft.com/v1.0/directoryObjects/%s"}`, userID))
	resp := authedPost(t, tss, addOwnerURL, addOwnerBody)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// List owners
	listOwnersURL := tss.BaseURL + "/v1.0/applications/" + appID + "/owners"
	resp2 := authedGet(t, tss, listOwnersURL)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	result := readJSON(t, resp2)
	value, ok := result["value"].([]interface{})
	require.True(t, ok, "Response should have value array")
	assert.NotEmpty(t, value, "Owners list should not be empty")

	// Verify the owner is in the list
	found := false
	for _, item := range value {
		if owner, ok := item.(map[string]interface{}); ok {
			if owner["id"] == userID {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "Added owner should appear in owners list")

	// Remove owner
	removeOwnerURL := tss.BaseURL + "/v1.0/applications/" + appID + "/owners/" + userID + "/$ref"
	resp3 := authedDelete(t, tss, removeOwnerURL)
	defer resp3.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp3.StatusCode)
}

// TestApplication_ExtensionProperties tests creating, listing, and deleting extension properties
func TestApplication_ExtensionProperties(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	app := createTestApplicationSDK(t, tss, "ExtProp Test App")
	appID := *app.GetId()

	// Create extension property
	createExtURL := tss.BaseURL + "/v1.0/applications/" + appID + "/extensionProperties"
	createExtBody := []byte(`{"name":"testExt","dataType":"String","targetObjects":["User"]}`)
	resp := authedPost(t, tss, createExtURL, createExtBody)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	result := readJSON(t, resp)
	extID, ok := result["id"].(string)
	require.True(t, ok, "Response should contain id")
	assert.NotEmpty(t, extID)
	extName, ok := result["name"].(string)
	require.True(t, ok, "Response should contain name")
	assert.Equal(t, "testExt", extName)

	// List extension properties
	listExtURL := tss.BaseURL + "/v1.0/applications/" + appID + "/extensionProperties"
	resp2 := authedGet(t, tss, listExtURL)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	result2 := readJSON(t, resp2)
	value, ok := result2["value"].([]interface{})
	require.True(t, ok, "Response should have value array")
	found := false
	for _, item := range value {
		if ext, ok := item.(map[string]interface{}); ok {
			if ext["id"] == extID {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "Created extension property should appear in list")

	// Delete extension property
	deleteExtURL := tss.BaseURL + "/v1.0/applications/" + appID + "/extensionProperties/" + extID
	resp3 := authedDelete(t, tss, deleteExtURL)
	defer resp3.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp3.StatusCode)
}

// TestApplication_Delta tests the applications delta endpoint
func TestApplication_Delta(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	app := createTestApplicationSDK(t, tss, "Delta Test App")
	appID := *app.GetId()

	// Get delta
	deltaURL := tss.BaseURL + "/v1.0/applications/delta"
	resp := authedGet(t, tss, deltaURL)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	result := readJSON(t, resp)

	// Verify response structure
	_, hasContext := result["@odata.context"]
	assert.True(t, hasContext, "Response should have @odata.context")

	value, ok := result["value"].([]interface{})
	require.True(t, ok, "Response should have value array")

	_, hasDeltaLink := result["@odata.deltaLink"]
	assert.True(t, hasDeltaLink, "Response should have @odata.deltaLink")

	// Verify the created app appears in delta results
	found := false
	for _, item := range value {
		if appItem, ok := item.(map[string]interface{}); ok {
			if appItem["id"] == appID {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "Created application should appear in delta results")
}

// TestApplication_DuplicateAppID tests that creating an app with a duplicate appId returns 409
func TestApplication_DuplicateAppID(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	specificAppID := uuid.New().String()

	// Create first application with specific appId via raw HTTP
	body := []byte(fmt.Sprintf(`{"displayName":"First App","appId":"%s"}`, specificAppID))
	resp := authedPost(t, tss, tss.BaseURL+"/v1.0/applications", body)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// Try to create second application with same appId
	body2 := []byte(fmt.Sprintf(`{"displayName":"Second App","appId":"%s"}`, specificAppID))
	resp2 := authedPost(t, tss, tss.BaseURL+"/v1.0/applications", body2)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusConflict, resp2.StatusCode)
}

// TestApplication_ODataParams tests OData query parameters ($top, $count)
func TestApplication_ODataParams(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	// Create 3 applications with unique names
	for i := 1; i <= 3; i++ {
		createTestApplicationSDK(t, tss, fmt.Sprintf("OData Test App %d %s", i, uuid.New().String()[:8]))
	}

	// Test $top
	topURL := tss.BaseURL + "/v1.0/applications?$top=1"
	resp := authedGet(t, tss, topURL)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	result := readJSON(t, resp)
	value, ok := result["value"].([]interface{})
	require.True(t, ok, "Response should have value array")
	assert.Len(t, value, 1, "$top=1 should return exactly 1 result")

	// Test $count
	countURL := tss.BaseURL + "/v1.0/applications?$count=true"
	resp2 := authedGet(t, tss, countURL)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	result2 := readJSON(t, resp2)
	count, hasCount := result2["@odata.count"]
	assert.True(t, hasCount, "Response should have @odata.count when $count=true")
	if countFloat, ok := count.(float64); ok {
		assert.GreaterOrEqual(t, int(countFloat), 3, "@odata.count should be >= 3")
	}
}

// TestApplication_OrderBy tests $orderby ascending and descending
func TestApplication_OrderBy(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// Create 3 apps with sortable names using uuid suffix to avoid collisions
	nameA := "aaa-orderby-" + uuid.New().String()[:8]
	nameM := "mmm-orderby-" + uuid.New().String()[:8]
	nameZ := "zzz-orderby-" + uuid.New().String()[:8]

	appA := createTestApplicationSDK(t, tss, nameA)
	appM := createTestApplicationSDK(t, tss, nameM)
	appZ := createTestApplicationSDK(t, tss, nameZ)

	defer func() {
		tss.SDKClient.Applications().ByApplicationId(*appA.GetId()).Delete(ctx, nil)
		tss.SDKClient.Applications().ByApplicationId(*appM.GetId()).Delete(ctx, nil)
		tss.SDKClient.Applications().ByApplicationId(*appZ.GetId()).Delete(ctx, nil)
	}()

	// Test ascending order
	resp := authedGet(t, tss, tss.BaseURL+"/v1.0/applications?$orderby=displayName")
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	result := readJSON(t, resp)
	value, ok := result["value"].([]interface{})
	require.True(t, ok, "Response should have value array")

	// Filter to only our test apps
	var testApps []map[string]interface{}
	for _, item := range value {
		if app, ok := item.(map[string]interface{}); ok {
			dn, _ := app["displayName"].(string)
			if (strings.HasPrefix(dn, "aaa-") || strings.HasPrefix(dn, "mmm-") || strings.HasPrefix(dn, "zzz-")) && strings.Contains(dn, "-orderby-") {
				testApps = append(testApps, app)
			}
		}
	}
	require.Len(t, testApps, 3, "Should find exactly 3 test apps")

	dn0, _ := testApps[0]["displayName"].(string)
	dn1, _ := testApps[1]["displayName"].(string)
	dn2, _ := testApps[2]["displayName"].(string)
	assert.True(t, dn0 < dn1 && dn1 < dn2, "Apps should be in ascending displayName order: %s, %s, %s", dn0, dn1, dn2)

	// Test descending order
	resp2 := authedGet(t, tss, tss.BaseURL+"/v1.0/applications?$orderby=displayName%20desc")
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	result2 := readJSON(t, resp2)
	value2, ok := result2["value"].([]interface{})
	require.True(t, ok, "Response should have value array")

	var testApps2 []map[string]interface{}
	for _, item := range value2 {
		if app, ok := item.(map[string]interface{}); ok {
			dn, _ := app["displayName"].(string)
			if (strings.HasPrefix(dn, "aaa-") || strings.HasPrefix(dn, "mmm-") || strings.HasPrefix(dn, "zzz-")) && strings.Contains(dn, "-orderby-") {
				testApps2 = append(testApps2, app)
			}
		}
	}
	require.Len(t, testApps2, 3, "Should find exactly 3 test apps")

	dn0d, _ := testApps2[0]["displayName"].(string)
	dn1d, _ := testApps2[1]["displayName"].(string)
	dn2d, _ := testApps2[2]["displayName"].(string)
	assert.True(t, dn0d > dn1d && dn1d > dn2d, "Apps should be in descending displayName order: %s, %s, %s", dn0d, dn1d, dn2d)
}

// TestApplication_Select tests $select query parameter
func TestApplication_Select(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// Create app with description via raw HTTP POST
	body := []byte(fmt.Sprintf(`{"displayName":"Select Test App %s","description":"This is a test description"}`, uuid.New().String()[:8]))
	resp := authedPost(t, tss, tss.BaseURL+"/v1.0/applications", body)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	result := readJSON(t, resp)
	appID, ok := result["id"].(string)
	require.True(t, ok, "Response should contain id")
	require.NotEmpty(t, appID)

	defer func() {
		tss.SDKClient.Applications().ByApplicationId(appID).Delete(ctx, nil)
	}()

	// GET with $select=displayName,id
	resp2 := authedGet(t, tss, tss.BaseURL+"/v1.0/applications/"+appID+"?$select=displayName,id")
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	result2 := readJSON(t, resp2)

	// Verify selected fields are present
	_, hasDisplayName := result2["displayName"]
	assert.True(t, hasDisplayName, "displayName should be present")
	_, hasID := result2["id"]
	assert.True(t, hasID, "id should be present")
	_, hasContext := result2["@odata.context"]
	assert.True(t, hasContext, "@odata.context should be present")

	// Verify non-selected fields are absent
	_, hasDescription := result2["description"]
	assert.False(t, hasDescription, "description should not be present when not selected")
	_, hasSignInAudience := result2["signInAudience"]
	assert.False(t, hasSignInAudience, "signInAudience should not be present when not selected")
}
