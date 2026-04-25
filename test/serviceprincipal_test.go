//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/serviceprincipals"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getAutoCreatedSP finds the service principal that was auto-created for the given appId
func getAutoCreatedSP(t *testing.T, ts *TestServer, appId string) models.ServicePrincipalable {
	t.Helper()
	ctx := context.Background()

	filter := fmt.Sprintf("appId eq '%s'", appId)
	result, err := ts.SDKClient.ServicePrincipals().Get(ctx, &serviceprincipals.ServicePrincipalsRequestBuilderGetRequestConfiguration{
		QueryParameters: &serviceprincipals.ServicePrincipalsRequestBuilderGetQueryParameters{
			Filter: &filter,
		},
	})
	require.NoError(t, err, "Failed to find auto-created SP")
	spList := result.GetValue()
	require.Len(t, spList, 1, "Expected exactly 1 SP for appId %s", appId)
	return spList[0]
}

// TestServicePrincipal_CRUDLifecycle tests reading, updating, and deleting a service principal
func TestServicePrincipal_CRUDLifecycle(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// 1. Create application (which auto-creates a SP)
	app := createTestApplicationSDK(t, tss, "SP Test App")
	appId := *app.GetAppId()

	// 2. Get the auto-created service principal
	createdSP := getAutoCreatedSP(t, tss, appId)
	spID := *createdSP.GetId()
	assert.NotEmpty(t, spID, "Service principal should have an id")
	assert.Equal(t, appId, *createdSP.GetAppId())

	// 3. GET the service principal by ID
	sp, err := tss.SDKClient.ServicePrincipals().ByServicePrincipalId(spID).Get(ctx, nil)
	require.NoError(t, err, "Failed to get service principal")
	assert.Equal(t, appId, *sp.GetAppId())

	// 4. PATCH the service principal (change displayName)
	patch := models.NewServicePrincipal()
	updatedName := "Updated SP Display Name"
	patch.SetDisplayName(&updatedName)

	updated, err := tss.SDKClient.ServicePrincipals().ByServicePrincipalId(spID).Patch(ctx, patch, nil)
	require.NoError(t, err, "Failed to patch service principal")
	assert.Equal(t, "Updated SP Display Name", *updated.GetDisplayName())

	// 5. GET again to verify update
	verifySP, err := tss.SDKClient.ServicePrincipals().ByServicePrincipalId(spID).Get(ctx, nil)
	require.NoError(t, err, "Failed to get updated service principal")
	assert.Equal(t, "Updated SP Display Name", *verifySP.GetDisplayName())

	// 6. DELETE the service principal
	err = tss.SDKClient.ServicePrincipals().ByServicePrincipalId(spID).Delete(ctx, nil)
	require.NoError(t, err, "Failed to delete service principal")

	// 7. GET again, expect error (404)
	_, err = tss.SDKClient.ServicePrincipals().ByServicePrincipalId(spID).Get(ctx, nil)
	require.Error(t, err, "Expected error getting deleted service principal")
}

// TestServicePrincipal_ListServicePrincipals tests listing service principals
func TestServicePrincipal_ListServicePrincipals(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// Create 2 applications (which auto-create SPs)
	for i := 1; i <= 2; i++ {
		displayName := fmt.Sprintf("SP List App %d", i)
		createTestApplicationSDK(t, tss, displayName)
	}

	// List service principals
	result, err := tss.SDKClient.ServicePrincipals().Get(ctx, nil)
	require.NoError(t, err, "Failed to list service principals")

	spList := result.GetValue()
	assert.GreaterOrEqual(t, len(spList), 2, "Expected at least 2 service principals")

	// Verify structure
	firstSP := spList[0]
	assert.NotNil(t, firstSP.GetId(), "Service principal missing id field")
	assert.NotNil(t, firstSP.GetAppId(), "Service principal missing appId field")
}

// TestServicePrincipal_FilterByAppId tests filtering service principals by appId
func TestServicePrincipal_FilterByAppId(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	app := createTestApplicationSDK(t, tss, "SP Filter App")
	appId := *app.GetAppId()

	filter := fmt.Sprintf("appId eq '%s'", appId)
	result, err := tss.SDKClient.ServicePrincipals().Get(ctx, &serviceprincipals.ServicePrincipalsRequestBuilderGetRequestConfiguration{
		QueryParameters: &serviceprincipals.ServicePrincipalsRequestBuilderGetQueryParameters{
			Filter: &filter,
		},
	})
	require.NoError(t, err, "Failed to filter service principals")

	spList := result.GetValue()
	require.Len(t, spList, 1, "Expected exactly 1 service principal matching filter")
	assert.Equal(t, appId, *spList[0].GetAppId())
}

// TestServicePrincipal_Validation tests service principal creation validation
func TestServicePrincipal_Validation(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// 1. Try to create SP without appId
	sp := models.NewServicePrincipal()
	_, err := tss.SDKClient.ServicePrincipals().Post(ctx, sp, nil)
	require.Error(t, err, "Expected error creating SP without appId")

	// 2. Try to create SP with non-existent appId
	nonExistentAppId := "00000000-0000-0000-0000-000000000000"
	sp2 := models.NewServicePrincipal()
	sp2.SetAppId(&nonExistentAppId)
	_, err = tss.SDKClient.ServicePrincipals().Post(ctx, sp2, nil)
	require.Error(t, err, "Expected error creating SP with non-existent appId")

	// 3. Create an application (which auto-creates a SP), then try creating a duplicate SP
	app := createTestApplicationSDK(t, tss, "Duplicate SP Test App")
	appId := *app.GetAppId()

	// Try to create another SP for the same appId - should fail with 409
	sp3 := models.NewServicePrincipal()
	sp3.SetAppId(&appId)
	_, err = tss.SDKClient.ServicePrincipals().Post(ctx, sp3, nil)
	require.Error(t, err, "Expected error creating duplicate SP for same appId")
}

// TestServicePrincipal_GetByAppID tests getting an SP by appId using alternate key lookup
func TestServicePrincipal_GetByAppID(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	app := createTestApplicationSDK(t, tss, "SP GetByAppID Test App")
	appAppID := *app.GetAppId()

	url := tss.BaseURL + "/v1.0/servicePrincipals/(appId=" + appAppID + ")"
	resp := authedGet(t, tss, url)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	result := readJSON(t, resp)
	assert.Equal(t, appAppID, result["appId"])
}

// TestServicePrincipal_Owners tests adding, listing, and removing SP owners
func TestServicePrincipal_Owners(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	// Create app and get the auto-created SP
	app := createTestApplicationSDK(t, tss, "SP Owner Test App")
	appId := *app.GetAppId()
	sp := getAutoCreatedSP(t, tss, appId)
	spID := *sp.GetId()

	// Create a user
	user := createTestUserSDK(t, tss, "SP Owner User", "spowneruser@saldeti.local")
	userID := *user.GetId()

	// Add owner
	addOwnerURL := tss.BaseURL + "/v1.0/servicePrincipals/" + spID + "/owners/$ref"
	addOwnerBody := []byte(fmt.Sprintf(`{"@odata.id":"https://graph.microsoft.com/v1.0/directoryObjects/%s"}`, userID))
	resp := authedPost(t, tss, addOwnerURL, addOwnerBody)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// List owners
	listOwnersURL := tss.BaseURL + "/v1.0/servicePrincipals/" + spID + "/owners"
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
	removeOwnerURL := tss.BaseURL + "/v1.0/servicePrincipals/" + spID + "/owners/" + userID + "/$ref"
	resp3 := authedDelete(t, tss, removeOwnerURL)
	defer resp3.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp3.StatusCode)
}

// TestServicePrincipal_MemberOf tests listing group membership for a service principal
func TestServicePrincipal_MemberOf(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	// Create app and get the auto-created SP
	app := createTestApplicationSDK(t, tss, "SP MemberOf Test App")
	appId := *app.GetAppId()
	sp := getAutoCreatedSP(t, tss, appId)
	spID := *sp.GetId()

	// Create a group
	group := createTestGroupSDK(t, tss, "SP MemberOf Test Group")
	groupID := *group.GetId()

	// Add SP to group as member
	addMemberURL := tss.BaseURL + "/v1.0/groups/" + groupID + "/members/$ref"
	addMemberBody := []byte(fmt.Sprintf(`{"@odata.id":"https://graph.microsoft.com/v1.0/directoryObjects/%s"}`, spID))
	resp := authedPost(t, tss, addMemberURL, addMemberBody)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// GET memberOf for the SP
	memberOfURL := tss.BaseURL + "/v1.0/servicePrincipals/" + spID + "/memberOf"
	resp2 := authedGet(t, tss, memberOfURL)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	result := readJSON(t, resp2)
	value, ok := result["value"].([]interface{})
	require.True(t, ok, "Response should have value array")

	// Verify the group appears in the memberOf list
	found := false
	for _, item := range value {
		if g, ok := item.(map[string]interface{}); ok {
			if g["id"] == groupID {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "Group should appear in SP memberOf list")
}

// TestServicePrincipal_OAuth2Grants tests creating and listing OAuth2 permission grants for an SP
func TestServicePrincipal_OAuth2Grants(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	// Create 2 apps (client + resource), get both SPs
	clientApp := createTestApplicationSDK(t, tss, "SP OAuth2 Client App")
	clientAppId := *clientApp.GetAppId()
	clientSP := getAutoCreatedSP(t, tss, clientAppId)
	clientSpID := *clientSP.GetId()

	resourceApp := createTestApplicationSDK(t, tss, "SP OAuth2 Resource App")
	resourceAppId := *resourceApp.GetAppId()
	resourceSP := getAutoCreatedSP(t, tss, resourceAppId)
	resourceSpID := *resourceSP.GetId()

	// Create grant
	grantURL := tss.BaseURL + "/v1.0/oauth2PermissionGrants"
	grantBody := []byte(fmt.Sprintf(`{"clientId":"%s","resourceId":"%s","scope":"User.Read","consentType":"AllPrincipals"}`, clientSpID, resourceSpID))
	resp := authedPost(t, tss, grantURL, grantBody)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// GET oauth2PermissionGrants for the client SP
	spGrantsURL := tss.BaseURL + "/v1.0/servicePrincipals/" + clientSpID + "/oauth2PermissionGrants"
	resp2 := authedGet(t, tss, spGrantsURL)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	result := readJSON(t, resp2)
	value, ok := result["value"].([]interface{})
	require.True(t, ok, "Response should have value array")
	assert.NotEmpty(t, value, "OAuth2 grants list should not be empty")
}

// TestServicePrincipal_PasswordCredentials tests adding and removing password credentials on an SP
func TestServicePrincipal_PasswordCredentials(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	// Create app and get the auto-created SP
	app := createTestApplicationSDK(t, tss, "SP Password Test App")
	appId := *app.GetAppId()
	sp := getAutoCreatedSP(t, tss, appId)
	spID := *sp.GetId()

	// Add password
	addPassURL := tss.BaseURL + "/v1.0/servicePrincipals/" + spID + "/addPassword"
	addPassBody := []byte(`{"displayName":"SP Test Pass"}`)
	resp := authedPost(t, tss, addPassURL, addPassBody)
	defer resp.Body.Close()

	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated,
		"Expected 200 or 201, got %d", resp.StatusCode)

	result := readJSON(t, resp)
	secretText, hasSecret := result["secretText"].(string)
	assert.True(t, hasSecret, "Response should contain secretText")
	assert.NotEmpty(t, secretText)

	keyID, hasKeyID := result["keyId"].(string)
	assert.True(t, hasKeyID, "Response should contain keyId")
	assert.NotEmpty(t, keyID)

	// Remove password
	removePassURL := tss.BaseURL + "/v1.0/servicePrincipals/" + spID + "/removePassword"
	removePassBody := []byte(fmt.Sprintf(`{"keyId":"%s"}`, keyID))
	resp2 := authedPost(t, tss, removePassURL, removePassBody)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp2.StatusCode)
}

// TestServicePrincipal_PolicyNavigation tests policy navigation endpoints return empty arrays
func TestServicePrincipal_PolicyNavigation(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	// Create app and get the auto-created SP
	app := createTestApplicationSDK(t, tss, "SP Policy Test App")
	appId := *app.GetAppId()
	sp := getAutoCreatedSP(t, tss, appId)
	spID := *sp.GetId()

	policyEndpoints := []string{
		"/homeRealmDiscoveryPolicies",
		"/claimsMappingPolicies",
		"/tokenIssuancePolicies",
		"/tokenLifetimePolicies",
	}

	for _, ep := range policyEndpoints {
		t.Run(ep, func(t *testing.T) {
			policyURL := tss.BaseURL + "/v1.0/servicePrincipals/" + spID + ep
			resp := authedGet(t, tss, policyURL)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			result := readJSON(t, resp)
			value, ok := result["value"].([]interface{})
			require.True(t, ok, "Response should have value array")
			assert.Empty(t, value, "Policy list should be empty")
		})
	}
}
