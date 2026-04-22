//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// E2E Tests: License Management
// ============================================================================


// TestListSubscribedSkus tests GET /v1.0/subscribedSkus endpoint
func TestListSubscribedSkus(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	token := getToken(t, tss)

	req, _ := http.NewRequest("GET", tss.BaseURL+"/v1.0/subscribedSkus", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 200 status
	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK for subscribedSkus")

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))

	// Assert @odata.context contains "subscribedSkus"
	context, ok := result["@odata.context"].(string)
	require.True(t, ok, "Expected @odata.context in response")
	assert.Contains(t, context, "subscribedSkus", "Expected @odata.context to contain 'subscribedSkus'")

	// Assert value is an array
	value, ok := result["value"].([]interface{})
	require.True(t, ok, "Expected value to be an array")

	// Assert at least 10 items
	assert.GreaterOrEqual(t, len(value), 10, "Expected at least 10 subscribed SKUs")

	// Assert each item has required fields
	for i, sku := range value {
		skuMap, ok := sku.(map[string]interface{})
		require.True(t, ok, "Expected SKU item %d to be an object", i)

		// Check skuId
		_, hasSkuId := skuMap["skuId"]
		assert.True(t, hasSkuId, "Expected SKU item %d to have skuId", i)

		// Check skuPartNumber
		_, hasSkuPartNumber := skuMap["skuPartNumber"]
		assert.True(t, hasSkuPartNumber, "Expected SKU item %d to have skuPartNumber", i)

		// Check servicePlans
		_, hasServicePlans := skuMap["servicePlans"]
		assert.True(t, hasServicePlans, "Expected SKU item %d to have servicePlans", i)
	}

	// Verify ENTERPRISEPACK and SPE_E3 are present
	foundEnterprisePack := false
	foundSpeE3 := false
	for _, sku := range value {
		skuMap := sku.(map[string]interface{})
		if skuPartNumber, ok := skuMap["skuPartNumber"].(string); ok {
			if skuPartNumber == "ENTERPRISEPACK" {
				foundEnterprisePack = true
				t.Logf("Found ENTERPRISEPACK SKU: %v", skuMap["skuId"])
			}
			if skuPartNumber == "SPE_E3" {
				foundSpeE3 = true
				t.Logf("Found SPE_E3 SKU: %v", skuMap["skuId"])
			}
		}
	}
	assert.True(t, foundEnterprisePack, "Expected ENTERPRISEPACK SKU to be present")
	assert.True(t, foundSpeE3, "Expected SPE_E3 SKU to be present")
}

// TestAssignLicense tests POST /v1.0/users/{id}/assignLicense
func TestAssignLicense(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	token := getToken(t, tss)

	// Create a user
	user := createTestUserSDK(t, tss, "License Test User", "licensetest@saldeti.local")
	userID := *user.GetId()
	t.Logf("Created user with ID: %s", userID)

	// Assign ENTERPRISEPACK license
	enterpriseSkuId := "6fd2c87f-b296-42f0-b197-1e91e994b900"
	assignReq := map[string]interface{}{
		"addLicenses": []map[string]interface{}{
			{"skuId": enterpriseSkuId},
		},
		"removeLicenses": []interface{}{},
	}

	assignBody, _ := json.Marshal(assignReq)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1.0/users/%s/assignLicense", tss.BaseURL, userID), strings.NewReader(string(assignBody)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 200 OK
	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK from assignLicense")

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))

	// Assert assignedLicenses contains ENTERPRISEPACK entry
	assignedLicenses, ok := result["assignedLicenses"].([]interface{})
	require.True(t, ok, "Expected assignedLicenses in response")
	assert.NotEmpty(t, assignedLicenses, "Expected at least one assigned license")

	// Verify the assigned license has the correct SKU
	foundLicense := false
	for _, lic := range assignedLicenses {
		licMap := lic.(map[string]interface{})
		if skuId, ok := licMap["skuId"].(string); ok && skuId == enterpriseSkuId {
			foundLicense = true
			if skuPartNumber, ok := licMap["skuPartNumber"].(string); ok {
				assert.Equal(t, "ENTERPRISEPACK", skuPartNumber, "Expected skuPartNumber to be ENTERPRISEPACK")
			}
		}
	}
	assert.True(t, foundLicense, "Expected to find ENTERPRISEPACK license in assignedLicenses")

	// GET user to confirm license is persisted
	getReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, userID), nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getResp, err := tss.Server.Client().Do(getReq)
	require.NoError(t, err)
	defer getResp.Body.Close()

	getBody, _ := io.ReadAll(getResp.Body)
	var getUser map[string]interface{}
	require.NoError(t, json.Unmarshal(getBody, &getUser))

	getAssignedLicenses, ok := getUser["assignedLicenses"].([]interface{})
	require.True(t, ok, "Expected assignedLicenses in GET user response")
	assert.NotEmpty(t, getAssignedLicenses, "Expected license to be persisted")

	// Verify the persisted license
	foundPersistedLicense := false
	for _, lic := range getAssignedLicenses {
		licMap := lic.(map[string]interface{})
		if skuId, ok := licMap["skuId"].(string); ok && skuId == enterpriseSkuId {
			foundPersistedLicense = true
			if skuPartNumber, ok := licMap["skuPartNumber"].(string); ok {
				assert.Equal(t, "ENTERPRISEPACK", skuPartNumber, "Expected persisted skuPartNumber to be ENTERPRISEPACK")
			}
		}
	}
	assert.True(t, foundPersistedLicense, "Expected to find ENTERPRISEPACK license in persisted user")
}

// TestAssignLicenseWithDisabledPlans tests assigning a license with disabled plans
func TestAssignLicenseWithDisabledPlans(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	token := getToken(t, tss)

	// Create a user
	user := createTestUserSDK(t, tss, "License Disabled Plans User", "licensedisabled@saldeti.local")
	userID := *user.GetId()
	t.Logf("Created user with ID: %s", userID)

	// Assign ENTERPRISEPACK license with disabled plans
	enterpriseSkuId := "6fd2c87f-b296-42f0-b197-1e91e994b900"
	disabledPlans := []string{"Exchange Online", "SharePoint Online"}
	assignReq := map[string]interface{}{
		"addLicenses": []map[string]interface{}{
			{
				"skuId":         enterpriseSkuId,
				"disabledPlans": disabledPlans,
			},
		},
		"removeLicenses": []interface{}{},
	}

	assignBody, _ := json.Marshal(assignReq)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1.0/users/%s/assignLicense", tss.BaseURL, userID), strings.NewReader(string(assignBody)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 200 OK
	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK from assignLicense with disabled plans")

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))

	// Assert assignedLicenses contains the license with disabled plans
	assignedLicenses, ok := result["assignedLicenses"].([]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, assignedLicenses)

	// Verify the disabled plans are preserved
	foundLicense := false
	for _, lic := range assignedLicenses {
		licMap := lic.(map[string]interface{})
		if skuId, ok := licMap["skuId"].(string); ok && skuId == enterpriseSkuId {
			foundLicense = true
			if licDisabledPlans, ok := licMap["disabledPlans"].([]interface{}); ok {
				// Verify disabled plans match
				assert.Equal(t, len(disabledPlans), len(licDisabledPlans), "Expected disabled plans count to match")
				for _, dp := range licDisabledPlans {
					dpStr := dp.(string)
					assert.Contains(t, disabledPlans, dpStr, "Expected disabled plan to be in original list")
				}
			} else {
				t.Error("Expected disabledPlans to be present in license")
			}
		}
	}
	assert.True(t, foundLicense, "Expected to find license with disabled plans")
}

// TestRemoveLicense tests removing a license via assignLicense
func TestRemoveLicense(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	token := getToken(t, tss)

	// Create a user with ENTERPRISEPACK license
	enterpriseSkuId := "6fd2c87f-b296-42f0-b197-1e91e994b900"
	userWithLicense := map[string]interface{}{
		"displayName":       "Remove License User",
		"userPrincipalName": "removelicense@saldeti.local",
		"mail":              "removelicense@saldeti.local",
		"accountEnabled":    true,
		"userType":          "Member",
		"assignedLicenses": []map[string]interface{}{
			{"skuId": enterpriseSkuId, "skuPartNumber": "ENTERPRISEPACK"},
		},
	}
	createdUser := createUserViaHTTP(t, tss, token, userWithLicense)
	userID := createdUser["id"].(string)
	t.Logf("Created user with ID: %s", userID)

	// Verify license is initially present
	getReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, userID), nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getResp, err := tss.Server.Client().Do(getReq)
	require.NoError(t, err)
	defer getResp.Body.Close()

	getBody, _ := io.ReadAll(getResp.Body)
	var getUser map[string]interface{}
	require.NoError(t, json.Unmarshal(getBody, &getUser))
	initialLicenses := getUser["assignedLicenses"].([]interface{})
	assert.NotEmpty(t, initialLicenses, "Expected initial license to be present")

	// Remove the license
	removeReq := map[string]interface{}{
		"addLicenses":    []interface{}{},
		"removeLicenses": []map[string]interface{}{{"skuId": enterpriseSkuId}},
	}

	removeBody, _ := json.Marshal(removeReq)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1.0/users/%s/assignLicense", tss.BaseURL, userID), strings.NewReader(string(removeBody)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 200 OK
	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK from assignLicense removing license")

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(respBody, &result))

	// Verify license is removed from response (assignedLicenses should be nil or empty)
	finalLicenses, ok := result["assignedLicenses"]
	if ok && finalLicenses != nil {
		finalLicensesSlice, ok := finalLicenses.([]interface{})
		if !ok {
			t.Logf("assignedLicenses is not a slice: %T", finalLicenses)
		} else {
			assert.Empty(t, finalLicensesSlice, "Expected no assigned licenses after removal")
		}
	}
	// If assignedLicenses is nil or not present, that's expected after removal

	// GET user to confirm license is persistently removed
	getReq2, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1.0/users/%s", tss.BaseURL, userID), nil)
	getReq2.Header.Set("Authorization", "Bearer "+token)
	getResp2, err := tss.Server.Client().Do(getReq2)
	require.NoError(t, err)
	defer getResp2.Body.Close()

	getBody2, _ := io.ReadAll(getResp2.Body)
	var getUser2 map[string]interface{}
	require.NoError(t, json.Unmarshal(getBody2, &getUser2))

	persistedLicenses := getUser2["assignedLicenses"]
	if persistedLicenses != nil {
		persistedLicensesSlice, ok := persistedLicenses.([]interface{})
		if ok {
			assert.Empty(t, persistedLicensesSlice, "Expected no licenses in persisted user after removal")
		}
	}
	// If assignedLicenses is nil, that's expected after removal
}

// TestAssignLicenseMultiple tests assigning multiple licenses and removing one
func TestAssignLicenseMultiple(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	token := getToken(t, tss)

	// Create a user
	user := createTestUserSDK(t, tss, "Multiple License User", "multilicense@saldeti.local")
	userID := *user.GetId()
	t.Logf("Created user with ID: %s", userID)

	// Assign EMS + ENTERPRISEPACK licenses
	emsSkuId := "efccb6f7-5641-4e0e-bd10-b4976e1bf68e"
	enterpriseSkuId := "6fd2c87f-b296-42f0-b197-1e91e994b900"
	assignReq := map[string]interface{}{
		"addLicenses": []map[string]interface{}{
			{"skuId": emsSkuId},
			{"skuId": enterpriseSkuId},
		},
		"removeLicenses": []interface{}{},
	}

	assignBody, _ := json.Marshal(assignReq)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1.0/users/%s/assignLicense", tss.BaseURL, userID), strings.NewReader(string(assignBody)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 200 OK
	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK from assignLicense with multiple licenses")

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))

	// Verify both licenses are present
	assignedLicenses, ok := result["assignedLicenses"].([]interface{})
	require.True(t, ok)
	assert.Len(t, assignedLicenses, 2, "Expected 2 assigned licenses")

	foundEms := false
	foundEnterprise := false
	for _, lic := range assignedLicenses {
		licMap := lic.(map[string]interface{})
		if skuId, ok := licMap["skuId"].(string); ok {
			if skuId == emsSkuId {
				foundEms = true
			}
			if skuId == enterpriseSkuId {
				foundEnterprise = true
			}
		}
	}
	assert.True(t, foundEms, "Expected EMS license to be present")
	assert.True(t, foundEnterprise, "Expected ENTERPRISEPACK license to be present")

	// Remove EMS license, keep ENTERPRISEPACK
	removeReq := map[string]interface{}{
		"addLicenses":    []interface{}{},
		"removeLicenses": []map[string]interface{}{{"skuId": emsSkuId}},
	}

	removeBody, _ := json.Marshal(removeReq)
	req2, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1.0/users/%s/assignLicense", tss.BaseURL, userID), strings.NewReader(string(removeBody)))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := tss.Server.Client().Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	// Assert 200 OK
	require.Equal(t, http.StatusOK, resp2.StatusCode, "Expected 200 OK from assignLicense removing EMS")

	body2, _ := io.ReadAll(resp2.Body)
	var result2 map[string]interface{}
	require.NoError(t, json.Unmarshal(body2, &result2))

	// Verify only ENTERPRISEPACK remains
	finalLicenses, ok := result2["assignedLicenses"].([]interface{})
	require.True(t, ok)
	assert.Len(t, finalLicenses, 1, "Expected 1 assigned license after removing EMS")

	foundOnlyEnterprise := false
	for _, lic := range finalLicenses {
		licMap := lic.(map[string]interface{})
		if skuId, ok := licMap["skuId"].(string); ok && skuId == enterpriseSkuId {
			foundOnlyEnterprise = true
		}
	}
	assert.True(t, foundOnlyEnterprise, "Expected only ENTERPRISEPACK license to remain")
}

// TestAssignLicenseDuplicate tests that assigning the same SKU twice is idempotent
func TestAssignLicenseDuplicate(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	token := getToken(t, tss)

	// Create a user
	user := createTestUserSDK(t, tss, "Duplicate License User", "duplicatelic@saldeti.local")
	userID := *user.GetId()
	t.Logf("Created user with ID: %s", userID)

	// Assign ENTERPRISEPACK license first time
	enterpriseSkuId := "6fd2c87f-b296-42f0-b197-1e91e994b900"
	assignReq1 := map[string]interface{}{
		"addLicenses": []map[string]interface{}{
			{"skuId": enterpriseSkuId},
		},
		"removeLicenses": []interface{}{},
	}

	assignBody1, _ := json.Marshal(assignReq1)
	req1, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1.0/users/%s/assignLicense", tss.BaseURL, userID), strings.NewReader(string(assignBody1)))
	req1.Header.Set("Authorization", "Bearer "+token)
	req1.Header.Set("Content-Type", "application/json")
	resp1, err := tss.Server.Client().Do(req1)
	require.NoError(t, err)
	defer resp1.Body.Close()

	require.Equal(t, http.StatusOK, resp1.StatusCode, "Expected 200 OK from first assignLicense")

	body1, _ := io.ReadAll(resp1.Body)
	var result1 map[string]interface{}
	require.NoError(t, json.Unmarshal(body1, &result1))

	licenses1, ok := result1["assignedLicenses"].([]interface{})
	require.True(t, ok)
	assert.Len(t, licenses1, 1, "Expected 1 license after first assignment")

	// Assign the same license again (should be idempotent)
	assignReq2 := map[string]interface{}{
		"addLicenses": []map[string]interface{}{
			{"skuId": enterpriseSkuId},
		},
		"removeLicenses": []interface{}{},
	}

	assignBody2, _ := json.Marshal(assignReq2)
	req2, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1.0/users/%s/assignLicense", tss.BaseURL, userID), strings.NewReader(string(assignBody2)))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := tss.Server.Client().Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	require.Equal(t, http.StatusOK, resp2.StatusCode, "Expected 200 OK from second assignLicense")

	body2, _ := io.ReadAll(resp2.Body)
	var result2 map[string]interface{}
	require.NoError(t, json.Unmarshal(body2, &result2))

	// Verify only one license entry (not duplicated)
	licenses2, ok := result2["assignedLicenses"].([]interface{})
	require.True(t, ok)
	assert.Len(t, licenses2, 1, "Expected 1 license after duplicate assignment (idempotent)")
}

// TestAssignLicenseInvalidUser tests assigning a license to a non-existent user
func TestAssignLicenseInvalidUser(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	token := getToken(t, tss)

	// Try to assign a license to a non-existent user ID
	nonExistentUserID := "00000000-0000-0000-0000-000000000000"
	enterpriseSkuId := "6fd2c87f-b296-42f0-b197-1e91e994b900"
	assignReq := map[string]interface{}{
		"addLicenses": []map[string]interface{}{
			{"skuId": enterpriseSkuId},
		},
		"removeLicenses": []interface{}{},
	}

	assignBody, _ := json.Marshal(assignReq)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1.0/users/%s/assignLicense", tss.BaseURL, nonExistentUserID), strings.NewReader(string(assignBody)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 404 Not Found
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Expected 404 Not Found for non-existent user")
}

// TestFilterByLicenseAfterAssign tests filtering users by assignedLicenses
func TestFilterByLicenseAfterAssign(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	token := getToken(t, tss)

	// Create two users
	licensedUser := createTestUserSDK(t, tss, "Licensed Filter User", "licensedfilter@saldeti.local")
	licensedUserID := *licensedUser.GetId()

	unlicensedUser := createTestUserSDK(t, tss, "Unlicensed Filter User", "unlicensedfilter@saldeti.local")
	unlicensedUserID := *unlicensedUser.GetId()

	t.Logf("Created licensed user ID: %s", licensedUserID)
	t.Logf("Created unlicensed user ID: %s", unlicensedUserID)

	// Assign ENTERPRISEPACK license to one user
	enterpriseSkuId := "6fd2c87f-b296-42f0-b197-1e91e994b900"
	assignReq := map[string]interface{}{
		"addLicenses": []map[string]interface{}{
			{"skuId": enterpriseSkuId},
		},
		"removeLicenses": []interface{}{},
	}

	assignBody, _ := json.Marshal(assignReq)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1.0/users/%s/assignLicense", tss.BaseURL, licensedUserID), strings.NewReader(string(assignBody)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK from assignLicense")

	// Filter users by assignedLicenses using $filter
	filterQuery := url.QueryEscape(`assignedLicenses/any(a:a/skuPartNumber eq 'ENTERPRISEPACK')`)
	filterReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1.0/users?$filter=%s", tss.BaseURL, filterQuery), nil)
	filterReq.Header.Set("Authorization", "Bearer "+token)
	filterResp, err := tss.Server.Client().Do(filterReq)
	require.NoError(t, err)
	defer filterResp.Body.Close()

	require.Equal(t, http.StatusOK, filterResp.StatusCode, "Expected 200 OK from filter query")

	filterBody, _ := io.ReadAll(filterResp.Body)
	var filterResult map[string]interface{}
	require.NoError(t, json.Unmarshal(filterBody, &filterResult))

	values, ok := filterResult["value"].([]interface{})
	require.True(t, ok, "Expected value to be an array")

	// Verify only the licensed user is returned
	assert.NotEmpty(t, values, "Expected at least one user in filter results")

	foundLicensed := false
	foundUnlicensed := false
	for _, v := range values {
		user := v.(map[string]interface{})
		userID, ok := user["id"].(string)
		if !ok {
			continue
		}

		if userID == licensedUserID {
			foundLicensed = true
			t.Logf("Found licensed user in filter results: %s", user["displayName"])
		}
		if userID == unlicensedUserID {
			foundUnlicensed = true
			t.Logf("Found unlicensed user in filter results: %s", user["displayName"])
		}
	}

	assert.True(t, foundLicensed, "Expected to find licensed user in filter results")
	assert.False(t, foundUnlicensed, "Expected unlicensed user NOT to be in filter results")
}
