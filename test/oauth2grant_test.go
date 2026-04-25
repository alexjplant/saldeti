//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOAuth2Grant_CRUDViaHTTP tests OAuth2 permission grant CRUD using raw HTTP requests
func TestOAuth2Grant_CRUDViaHTTP(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	// 1. Create application (auto-creates SP)
	app := createTestApplicationSDK(t, tss, "Grant Test App")
	appId := *app.GetAppId()
	sp := getAutoCreatedSP(t, tss, appId)
	spID := *sp.GetId()

	// Also create a user to be the principal
	user := createTestUserSDK(t, tss, "Grant User", "grantuser@saldeti.local")
	userID := *user.GetId()

	// 2. Create OAuth2 permission grant via raw HTTP POST
	grantBody := map[string]interface{}{
		"clientId":    appId,
		"consentType": "Principal",
		"principalId": userID,
		"resourceId":  spID,
		"scope":       "User.Read Mail.Read",
	}
	grantJSON, err := json.Marshal(grantBody)
	require.NoError(t, err)

	createResp := authedPost(t, tss, tss.BaseURL+"/v1.0/oauth2PermissionGrants", grantJSON)
	defer createResp.Body.Close()
	require.Equal(t, http.StatusCreated, createResp.StatusCode, "Expected 201 for grant creation")

	grantResp := readJSON(t, createResp)
	grantID := grantResp["id"].(string)
	assert.NotEmpty(t, grantID, "Grant should have an id")
	assert.Equal(t, appId, grantResp["clientId"])
	assert.Equal(t, "Principal", grantResp["consentType"])
	assert.Equal(t, "User.Read Mail.Read", grantResp["scope"])

	// 3. GET the grant by ID
	getResp := authedGet(t, tss, tss.BaseURL+"/v1.0/oauth2PermissionGrants/"+grantID)
	defer getResp.Body.Close()
	assert.Equal(t, http.StatusOK, getResp.StatusCode)

	// 4. LIST grants
	listResp := authedGet(t, tss, tss.BaseURL+"/v1.0/oauth2PermissionGrants")
	defer listResp.Body.Close()
	assert.Equal(t, http.StatusOK, listResp.StatusCode)

	listResult := readJSON(t, listResp)
	values := listResult["value"].([]interface{})
	assert.GreaterOrEqual(t, len(values), 1, "Expected at least 1 grant")

	// 5. PATCH the grant (update scope)
	patchBody := map[string]interface{}{
		"scope": "User.Read",
	}
	patchJSON, err := json.Marshal(patchBody)
	require.NoError(t, err)

	patchResp := authedPatch(t, tss, tss.BaseURL+"/v1.0/oauth2PermissionGrants/"+grantID, patchJSON)
	defer patchResp.Body.Close()
	assert.Equal(t, http.StatusOK, patchResp.StatusCode)

	patchedGrant := readJSON(t, patchResp)
	assert.Equal(t, "User.Read", patchedGrant["scope"])

	// 6. DELETE the grant
	deleteResp := authedDelete(t, tss, tss.BaseURL+"/v1.0/oauth2PermissionGrants/"+grantID)
	defer deleteResp.Body.Close()
	assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

	// 7. GET again, expect 404
	getResp2 := authedGet(t, tss, tss.BaseURL+"/v1.0/oauth2PermissionGrants/"+grantID)
	defer getResp2.Body.Close()
	assert.Equal(t, http.StatusNotFound, getResp2.StatusCode)
}

// TestOAuth2Grant_ListGrantsForSP tests listing OAuth2 grants for a specific service principal
func TestOAuth2Grant_ListGrantsForSP(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	// Create application (auto-creates SP)
	app := createTestApplicationSDK(t, tss, "SP Grant List App")
	appId := *app.GetAppId()
	sp := getAutoCreatedSP(t, tss, appId)
	spID := *sp.GetId()

 	// Create a grant
 	grantBody := map[string]interface{}{
 		"clientId":    spID,
 		"consentType": "AllPrincipals",
		"resourceId":  spID,
		"scope":       "Directory.Read.All",
	}
	grantJSON, _ := json.Marshal(grantBody)
	createResp := authedPost(t, tss, tss.BaseURL+"/v1.0/oauth2PermissionGrants", grantJSON)
	defer createResp.Body.Close()
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	// List grants via SP navigation: GET /v1.0/servicePrincipals/{spID}/oauth2PermissionGrants
	listResp := authedGet(t, tss, fmt.Sprintf("%s/v1.0/servicePrincipals/%s/oauth2PermissionGrants", tss.BaseURL, spID))
	defer listResp.Body.Close()
	assert.Equal(t, http.StatusOK, listResp.StatusCode)

	listResult := readJSON(t, listResp)
	values := listResult["value"].([]interface{})
	assert.GreaterOrEqual(t, len(values), 1, "Expected at least 1 grant for SP")
}

// TestOAuth2Grant_FilterByClientId tests filtering OAuth2 grants by clientId
func TestOAuth2Grant_FilterByClientId(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	// Create 2 client applications (auto-create SPs)
	app1 := createTestApplicationSDK(t, tss, "Filter Client App 1")
	app1Id := *app1.GetAppId()
	sp1 := getAutoCreatedSP(t, tss, app1Id)
	clientSpId1 := *sp1.GetId()

	app2 := createTestApplicationSDK(t, tss, "Filter Client App 2")
	app2Id := *app2.GetAppId()
	sp2 := getAutoCreatedSP(t, tss, app2Id)
	clientSpId2 := *sp2.GetId()

	// Create a resource application/SP
	resourceApp := createTestApplicationSDK(t, tss, "Filter Resource App")
	resourceAppId := *resourceApp.GetAppId()
	resourceSP := getAutoCreatedSP(t, tss, resourceAppId)
	resourceSpId := *resourceSP.GetId()

	// Create OAuth2 grant with clientSpId1 as clientId and resourceSpId as resourceId
	grantBody1 := map[string]interface{}{
		"clientId":    clientSpId1,
		"consentType": "AllPrincipals",
		"resourceId":  resourceSpId,
		"scope":       "User.Read",
	}
	grantJSON1, err := json.Marshal(grantBody1)
	require.NoError(t, err)
	createResp1 := authedPost(t, tss, tss.BaseURL+"/v1.0/oauth2PermissionGrants", grantJSON1)
	defer createResp1.Body.Close()
	require.Equal(t, http.StatusCreated, createResp1.StatusCode, "Expected 201 for grant 1 creation")

	// Create OAuth2 grant with clientSpId2 as clientId and resourceSpId as resourceId
	grantBody2 := map[string]interface{}{
		"clientId":    clientSpId2,
		"consentType": "AllPrincipals",
		"resourceId":  resourceSpId,
		"scope":       "Mail.Read",
	}
	grantJSON2, err := json.Marshal(grantBody2)
	require.NoError(t, err)
	createResp2 := authedPost(t, tss, tss.BaseURL+"/v1.0/oauth2PermissionGrants", grantJSON2)
	defer createResp2.Body.Close()
	require.Equal(t, http.StatusCreated, createResp2.StatusCode, "Expected 201 for grant 2 creation")

	// GET /oauth2PermissionGrants?$filter=clientId eq '{clientSpId1}'
	filterExpr := url.QueryEscape(fmt.Sprintf("clientId eq '%s'", clientSpId1))
	filterResp := authedGet(t, tss, fmt.Sprintf("%s/v1.0/oauth2PermissionGrants?$filter=%s", tss.BaseURL, filterExpr))
	defer filterResp.Body.Close()
	assert.Equal(t, http.StatusOK, filterResp.StatusCode)

	filterResult := readJSON(t, filterResp)
	values := filterResult["value"].([]interface{})
	assert.Equal(t, 1, len(values), "Expected exactly 1 grant with matching clientId")

	grant := values[0].(map[string]interface{})
	assert.Equal(t, clientSpId1, grant["clientId"])
}

// TestOAuth2Grant_FilterByResourceId tests filtering OAuth2 grants by resourceId
func TestOAuth2Grant_FilterByResourceId(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	// Create 2 client applications (auto-create SPs)
	app1 := createTestApplicationSDK(t, tss, "Resource Filter Client App 1")
	app1Id := *app1.GetAppId()
	sp1 := getAutoCreatedSP(t, tss, app1Id)
	clientSpId1 := *sp1.GetId()

	app2 := createTestApplicationSDK(t, tss, "Resource Filter Client App 2")
	app2Id := *app2.GetAppId()
	sp2 := getAutoCreatedSP(t, tss, app2Id)
	clientSpId2 := *sp2.GetId()

	// Create a resource application/SP
	resourceApp := createTestApplicationSDK(t, tss, "Resource Filter Resource App")
	resourceAppId := *resourceApp.GetAppId()
	resourceSP := getAutoCreatedSP(t, tss, resourceAppId)
	resourceSpId := *resourceSP.GetId()

	// Create OAuth2 grant with clientSpId1 as clientId and resourceSpId as resourceId
	grantBody1 := map[string]interface{}{
		"clientId":    clientSpId1,
		"consentType": "AllPrincipals",
		"resourceId":  resourceSpId,
		"scope":       "User.Read",
	}
	grantJSON1, err := json.Marshal(grantBody1)
	require.NoError(t, err)
	createResp1 := authedPost(t, tss, tss.BaseURL+"/v1.0/oauth2PermissionGrants", grantJSON1)
	defer createResp1.Body.Close()
	require.Equal(t, http.StatusCreated, createResp1.StatusCode, "Expected 201 for grant 1 creation")

	// Create OAuth2 grant with clientSpId2 as clientId and resourceSpId as resourceId
	grantBody2 := map[string]interface{}{
		"clientId":    clientSpId2,
		"consentType": "AllPrincipals",
		"resourceId":  resourceSpId,
		"scope":       "Mail.Read",
	}
	grantJSON2, err := json.Marshal(grantBody2)
	require.NoError(t, err)
	createResp2 := authedPost(t, tss, tss.BaseURL+"/v1.0/oauth2PermissionGrants", grantJSON2)
	defer createResp2.Body.Close()
	require.Equal(t, http.StatusCreated, createResp2.StatusCode, "Expected 201 for grant 2 creation")

	// GET /oauth2PermissionGrants?$filter=resourceId eq '{resourceSpId}'
	filterExpr := url.QueryEscape(fmt.Sprintf("resourceId eq '%s'", resourceSpId))
	filterResp := authedGet(t, tss, fmt.Sprintf("%s/v1.0/oauth2PermissionGrants?$filter=%s", tss.BaseURL, filterExpr))
	defer filterResp.Body.Close()
	assert.Equal(t, http.StatusOK, filterResp.StatusCode)

	filterResult := readJSON(t, filterResp)
	values := filterResult["value"].([]interface{})
	assert.Equal(t, 2, len(values), "Expected exactly 2 grants with matching resourceId")

	for _, v := range values {
		grant := v.(map[string]interface{})
		assert.Equal(t, resourceSpId, grant["resourceId"])
	}
}

func TestOAuth2Grant_CreateValidation(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	// Create a resource SP so we have a valid resourceId
	resourceApp := createTestApplicationSDK(t, tss, "Validation Resource App")
	resourceAppId := *resourceApp.GetAppId()
	resourceSP := getAutoCreatedSP(t, tss, resourceAppId)
	resourceSpId := *resourceSP.GetId()

	// Also create a client SP for valid clientId
	clientApp := createTestApplicationSDK(t, tss, "Validation Client App")
	clientAppId := *clientApp.GetAppId()
	clientSP := getAutoCreatedSP(t, tss, clientAppId)
	clientSpId := *clientSP.GetId()

	// Subtest 1: Empty body → 400
	resp := authedPost(t, tss, tss.BaseURL+"/v1.0/oauth2PermissionGrants", []byte(`{}`))
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Subtest 2: Missing clientId → 400
	body, _ := json.Marshal(map[string]interface{}{
		"consentType": "AllPrincipals",
		"resourceId":  resourceSpId,
		"scope":       "User.Read",
	})
	resp2 := authedPost(t, tss, tss.BaseURL+"/v1.0/oauth2PermissionGrants", body)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp2.StatusCode)

	// Subtest 3: Invalid consentType → 400
	body3, _ := json.Marshal(map[string]interface{}{
		"clientId":    clientSpId,
		"consentType": "InvalidType",
		"resourceId":  resourceSpId,
		"scope":       "User.Read",
	})
	resp3 := authedPost(t, tss, tss.BaseURL+"/v1.0/oauth2PermissionGrants", body3)
	defer resp3.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp3.StatusCode)

	// Subtest 4: Principal consentType without principalId → 400
	body4, _ := json.Marshal(map[string]interface{}{
		"clientId":    clientSpId,
		"consentType": "Principal",
		"resourceId":  resourceSpId,
		"scope":       "User.Read",
	})
	resp4 := authedPost(t, tss, tss.BaseURL+"/v1.0/oauth2PermissionGrants", body4)
	defer resp4.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp4.StatusCode)
}
