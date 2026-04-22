//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// E2E Tests: $search field-qualified syntax
// ============================================================================

func TestE2E_SearchFieldQualified_DisplayName(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	// Create users with distinct names
	createTestUserSDK(t, tss, "Alpha Beta", "alpha.beta@saldeti.local")
	createTestUserSDK(t, tss, "Gamma Delta", "gamma.delta@saldeti.local")

	// Use raw HTTP to test $search with field-qualified syntax
	token := getToken(t, tss)

	// Search displayName:Alpha — should match only Alpha Beta
	req, _ := http.NewRequest("GET", tss.BaseURL+"/v1.0/users?$search=%22displayName%3AAlpha%22", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))

	values, ok := result["value"].([]interface{})
	require.True(t, ok, "Expected value to be an array")
	assert.Len(t, values, 1, "Expected exactly 1 result for displayName:Alpha")

	if len(values) > 0 {
		user := values[0].(map[string]interface{})
		assert.Equal(t, "Alpha Beta", user["displayName"])
	}
}

func TestE2E_SearchFieldQualified_Mail(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	token := getToken(t, tss)

	// Create users
	createTestUserSDK(t, tss, "Test User Search1", "search1@saldeti.local")
	createTestUserSDK(t, tss, "Test User Search2", "search2@saldeti.local")

	// Search mail:search1@ — should match only user with search1@saldeti.local
	req, _ := http.NewRequest("GET", tss.BaseURL+"/v1.0/users?$search=%22mail%3Asearch1%40%22", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))

	values, ok := result["value"].([]interface{})
	require.True(t, ok)
	assert.Len(t, values, 1, "Expected exactly 1 result for mail:search1@")
}

func TestE2E_SearchUnqualified(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	token := getToken(t, tss)

	createTestUserSDK(t, tss, "UniqueSearchName User", "uniquesearch@saldeti.local")
	createTestUserSDK(t, tss, "OtherPerson User", "otherperson@saldeti.local")

	// Unqualified search — should match across displayName, mail, UPN
	req, _ := http.NewRequest("GET", tss.BaseURL+"/v1.0/users?$search=%22UniqueSearchName%22", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))

	values, ok := result["value"].([]interface{})
	require.True(t, ok)
	assert.Len(t, values, 1, "Expected exactly 1 result for UniqueSearchName")
}

// ============================================================================
// E2E Tests: $filter userType eq 'Member' / 'Guest'
// ============================================================================

func TestE2E_FilterUserType(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// Create a member user
	memberUser := models.NewUser()
	memberUser.SetDisplayName(ptrString("Member User"))
	memberUser.SetUserPrincipalName(ptrString("member.filter@saldeti.local"))
	memberUser.SetMail(ptrString("member.filter@saldeti.local"))
	enabled := true
	memberUser.SetAccountEnabled(&enabled)
	pp := models.NewPasswordProfile()
	pp.SetPassword(ptrString("Test1234!"))
	memberUser.SetPasswordProfile(pp)
	memberUser.SetUserType(ptrString("Member"))
	_, err := tss.SDKClient.Users().Post(ctx, memberUser, nil)
	require.NoError(t, err)

	// Create a guest user
	guestUser := models.NewUser()
	guestUser.SetDisplayName(ptrString("Guest User"))
	guestUser.SetUserPrincipalName(ptrString("guest.filter@saldeti.local"))
	guestUser.SetMail(ptrString("guest.filter@saldeti.local"))
	guestUser.SetAccountEnabled(&enabled)
	pp2 := models.NewPasswordProfile()
	pp2.SetPassword(ptrString("Test1234!"))
	guestUser.SetPasswordProfile(pp2)
	guestUser.SetUserType(ptrString("Guest"))
	_, err = tss.SDKClient.Users().Post(ctx, guestUser, nil)
	require.NoError(t, err)

	// Filter by userType eq 'Member'
	filter := "userType eq 'Member'"
	result, err := tss.SDKClient.Users().Get(ctx, &users.UsersRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
			Filter: &filter,
		},
	})
	require.NoError(t, err)

	userList := result.GetValue()
	assert.NotEmpty(t, userList, "Expected at least one Member user")

	// Verify all results are Members
	for _, u := range userList {
		ut := u.GetUserType()
		if ut != nil {
			assert.Equal(t, "Member", *ut, "Expected userType to be 'Member'")
		}
	}

	// Filter by userType eq 'Guest'
	filter = "userType eq 'Guest'"
	result, err = tss.SDKClient.Users().Get(ctx, &users.UsersRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
			Filter: &filter,
		},
	})
	require.NoError(t, err)

	guestList := result.GetValue()
	assert.NotEmpty(t, guestList, "Expected at least one Guest user")

	for _, u := range guestList {
		ut := u.GetUserType()
		if ut != nil {
			assert.Equal(t, "Guest", *ut, "Expected userType to be 'Guest'")
		}
	}
}

func TestE2E_FilterUserTypeCombinedWithAccountEnabled(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// Create enabled member
	enabled := true
	u1 := models.NewUser()
	u1.SetDisplayName(ptrString("Enabled Member"))
	u1.SetUserPrincipalName(ptrString("enabled.member@saldeti.local"))
	u1.SetMail(ptrString("enabled.member@saldeti.local"))
	u1.SetAccountEnabled(&enabled)
	pp := models.NewPasswordProfile()
	pp.SetPassword(ptrString("Test1234!"))
	u1.SetPasswordProfile(pp)
	u1.SetUserType(ptrString("Member"))
	_, err := tss.SDKClient.Users().Post(ctx, u1, nil)
	require.NoError(t, err)

	// Create disabled member
	disabled := false
	u2 := models.NewUser()
	u2.SetDisplayName(ptrString("Disabled Member"))
	u2.SetUserPrincipalName(ptrString("disabled.member@saldeti.local"))
	u2.SetMail(ptrString("disabled.member@saldeti.local"))
	u2.SetAccountEnabled(&disabled)
	pp2 := models.NewPasswordProfile()
	pp2.SetPassword(ptrString("Test1234!"))
	u2.SetPasswordProfile(pp2)
	u2.SetUserType(ptrString("Member"))
	_, err = tss.SDKClient.Users().Post(ctx, u2, nil)
	require.NoError(t, err)

	// Filter: userType eq 'Member' and accountEnabled eq true
	filter := "userType eq 'Member' and accountEnabled eq true"
	result, err := tss.SDKClient.Users().Get(ctx, &users.UsersRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
			Filter: &filter,
		},
	})
	require.NoError(t, err)

	userList := result.GetValue()
	found := false
	for _, u := range userList {
		if u.GetDisplayName() != nil && *u.GetDisplayName() == "Enabled Member" {
			found = true
		}
	}
	assert.True(t, found, "Expected to find 'Enabled Member' in results")

	// Verify "Disabled Member" is NOT in results
	for _, u := range userList {
		if u.GetDisplayName() != nil {
			assert.NotEqual(t, "Disabled Member", *u.GetDisplayName(), "Disabled member should not be in enabled results")
		}
	}
}

// ============================================================================
// E2E Tests: $filter on assignedLicenses with nested any()
// ============================================================================

func TestE2E_FilterAssignedLicenses(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	token := getToken(t, tss)

	// Create user with licenses via raw HTTP (SDK may not support assignedLicenses directly)
	licensedUser := map[string]interface{}{
		"displayName":       "Licensed User",
		"userPrincipalName": "licensed@saldeti.local",
		"mail":              "licensed@saldeti.local",
		"accountEnabled":    true,
		"userType":          "Member",
		"assignedLicenses": []map[string]interface{}{
			{"skuId": "sku-enterprise-001", "skuPartNumber": "ENTERPRISEPACK"},
			{"skuId": "sku-ems-002", "skuPartNumber": "EMS"},
		},
	}
	createUserViaHTTP(t, tss, token, licensedUser)

	// Create user without licenses
	createTestUserSDK(t, tss, "Unlicensed User", "unlicensed@saldeti.local")

	// Filter by assignedLicenses/any(a:a/skuId eq 'sku-enterprise-001')
	req, _ := http.NewRequest("GET", tss.BaseURL+`/v1.0/users?$filter=assignedLicenses%2Fany(a%3Aa%2FskuId+eq+'sku-enterprise-001')`, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))

	values, ok := result["value"].([]interface{})
	require.True(t, ok)

	found := false
	for _, v := range values {
		user := v.(map[string]interface{})
		if user["displayName"] == "Licensed User" {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected to find 'Licensed User' in assignedLicenses filter results")
}

func TestE2E_FilterAssignedLicensesBySkuPartNumber(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	token := getToken(t, tss)

	// Create user with EMS license
	emsUser := map[string]interface{}{
		"displayName":       "EMS User",
		"userPrincipalName": "ems@saldeti.local",
		"mail":              "ems@saldeti.local",
		"accountEnabled":    true,
		"userType":          "Member",
		"assignedLicenses": []map[string]interface{}{
			{"skuId": "sku-ems-001", "skuPartNumber": "EMS"},
		},
	}
	createUserViaHTTP(t, tss, token, emsUser)

	// Create user with FLOW license
	flowUser := map[string]interface{}{
		"displayName":       "Flow User",
		"userPrincipalName": "flow@saldeti.local",
		"mail":              "flow@saldeti.local",
		"accountEnabled":    true,
		"userType":          "Member",
		"assignedLicenses": []map[string]interface{}{
			{"skuId": "sku-flow-001", "skuPartNumber": "FLOW_FREE"},
		},
	}
	createUserViaHTTP(t, tss, token, flowUser)

	// Filter by skuPartNumber eq 'EMS'
	req, _ := http.NewRequest("GET", tss.BaseURL+`/v1.0/users?$filter=assignedLicenses%2Fany(a%3Aa%2FskuPartNumber+eq+'EMS')`, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))

	values, ok := result["value"].([]interface{})
	require.True(t, ok)

	found := false
	for _, v := range values {
		user := v.(map[string]interface{})
		if user["displayName"] == "EMS User" {
			found = true
		}
		if user["displayName"] == "Flow User" {
			t.Error("Flow User should not match EMS filter")
		}
	}
	assert.True(t, found, "Expected to find 'EMS User' in results")
}

func TestE2E_FilterAssignedLicensesNoMatch(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	token := getToken(t, tss)

	createTestUserSDK(t, tss, "No License User", "nolicense@saldeti.local")

	// Filter for a license that doesn't exist
	req, _ := http.NewRequest("GET", tss.BaseURL+`/v1.0/users?$filter=assignedLicenses%2Fany(a%3Aa%2FskuId+eq+'nonexistent-sku')`, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))

	values, ok := result["value"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, values, "Expected no results for nonexistent skuId")
}

// ============================================================================
// E2E Tests: $filter groupTypes/any(a:a eq 'Unified') (regression)
// ============================================================================

func TestE2E_FilterGroupTypesAny(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()
	ctx := context.Background()

	// Create a Unified (M365) group
	unifiedGroup := models.NewGroup()
	unifiedGroup.SetDisplayName(ptrString("Unified Test Group"))
	unifiedGroup.SetMailNickname(ptrString("unifiedtest"))
	unifiedGroup.SetGroupTypes([]string{"Unified"})
	secEnabled := false
	unifiedGroup.SetSecurityEnabled(&secEnabled)
	mailEnabled := true
	unifiedGroup.SetMailEnabled(&mailEnabled)
	_, err := tss.SDKClient.Groups().Post(ctx, unifiedGroup, nil)
	require.NoError(t, err)

	// Create a security group
	secGroup := models.NewGroup()
	secGroup.SetDisplayName(ptrString("Security Test Group"))
	secGroup.SetMailNickname(ptrString("sectest"))
	secEnabled2 := true
	secGroup.SetSecurityEnabled(&secEnabled2)
	mailEnabled2 := false
	secGroup.SetMailEnabled(&mailEnabled2)
	_, err = tss.SDKClient.Groups().Post(ctx, secGroup, nil)
	require.NoError(t, err)

	// Filter groups by groupTypes/any(a:a eq 'Unified') using raw HTTP
	// (SDK may encode the filter differently)
	req, _ := http.NewRequest("GET", tss.BaseURL+`/v1.0/groups?$filter=groupTypes%2Fany(a%3Aa+eq+'Unified')`, nil)
	req.Header.Set("Authorization", "Bearer "+getToken(t, tss))
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))

	values, ok := result["value"].([]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, values, "Expected at least one Unified group")

	// Verify all results have Unified groupType
	for _, v := range values {
		group := v.(map[string]interface{})
		assert.Equal(t, "Unified Test Group", group["displayName"])
	}
}

// ============================================================================
// Helpers
// ============================================================================

func getToken(t *testing.T, tss *TestServer) string {
	t.Helper()
	req, _ := http.NewRequest("POST", tss.BaseURL+"/sim-tenant-id/oauth2/v2.0/token", strings.NewReader("grant_type=client_credentials&client_id=sim-client-id&client_secret=sim-client-secret&scope=https://graph.microsoft.com/.default"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var tokenResp map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &tokenResp))
	token, ok := tokenResp["access_token"].(string)
	require.True(t, ok, "Expected access_token in response")
	return token
}

func createUserViaHTTP(t *testing.T, tss *TestServer, token string, user map[string]interface{}) map[string]interface{} {
	t.Helper()
	userJSON, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", tss.BaseURL+"/v1.0/users", strings.NewReader(string(userJSON)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := tss.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode, fmt.Sprintf("Expected 201 creating user, got %d", resp.StatusCode))
	body, _ := io.ReadAll(resp.Body)
	var created map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &created))
	return created
}
