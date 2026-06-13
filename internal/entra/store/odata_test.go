package store

import (
	"testing"

	"github.com/saldeti/saldeti/internal/entra/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// $search Tests — Field-Qualified Syntax
// ============================================================================

func TestApplySearch_Unqualified(t *testing.T) {
	items := []map[string]interface{}{
		{"displayName": "Alice Smith", "userPrincipalName": "alice@example.com", "mail": "alice@example.com", "mailNickname": "alice"},
		{"displayName": "Bob Jones", "userPrincipalName": "bob@example.com", "mail": "bob@example.com", "mailNickname": "bob"},
		{"displayName": "Charlie Brown", "userPrincipalName": "charlie@example.com", "mail": "charlie@example.com", "mailNickname": "charlie"},
	}

	// Unqualified search matches displayName
	result := applySearch(items, "Alice")
	assert.Len(t, result, 1)
	assert.Equal(t, "Alice Smith", result[0]["displayName"])

	// Unqualified search matches mail
	result = applySearch(items, "bob@")
	assert.Len(t, result, 1)
	assert.Equal(t, "Bob Jones", result[0]["displayName"])

	// Unqualified search matches userPrincipalName
	result = applySearch(items, "charlie@")
	assert.Len(t, result, 1)
	assert.Equal(t, "Charlie Brown", result[0]["displayName"])

	// Unqualified search matches mailNickname
	result = applySearch(items, "alice")
	assert.Len(t, result, 1)
	assert.Equal(t, "Alice Smith", result[0]["displayName"])
}

func TestApplySearch_FieldQualified(t *testing.T) {
	items := []map[string]interface{}{
		{"displayName": "Alice Smith", "userPrincipalName": "alice@example.com", "mail": "alice@corp.com", "mailNickname": "alice"},
		{"displayName": "Alice Jones", "userPrincipalName": "alice2@example.com", "mail": "alice2@corp.com", "mailNickname": "alice2"},
		{"displayName": "Bob Smith", "userPrincipalName": "bob@example.com", "mail": "bob@corp.com", "mailNickname": "bob"},
	}

	// displayName: — only match displayName field
	result := applySearch(items, "displayName:Alice Smith")
	assert.Len(t, result, 1)
	assert.Equal(t, "Alice Smith", result[0]["displayName"])

	// mail: — only match mail field
	result = applySearch(items, "mail:alice@corp.com")
	assert.Len(t, result, 1)
	assert.Equal(t, "Alice Smith", result[0]["displayName"])

	// userPrincipalName: — only match UPN
	result = applySearch(items, "userPrincipalName:bob@")
	assert.Len(t, result, 1)
	assert.Equal(t, "Bob Smith", result[0]["displayName"])

	// displayName:Alice matches both Alices
	result = applySearch(items, "displayName:Alice")
	assert.Len(t, result, 2)
}

func TestApplySearch_MultipleTermsImplicitAnd(t *testing.T) {
	items := []map[string]interface{}{
		{"displayName": "Alice Smith", "mail": "alice@example.com", "userPrincipalName": "alice@example.com", "mailNickname": "alice"},
		{"displayName": "Alice Jones", "mail": "alice2@corp.com", "userPrincipalName": "alice2@corp.com", "mailNickname": "alice2"},
		{"displayName": "Bob Smith", "mail": "bob@example.com", "userPrincipalName": "bob@example.com", "mailNickname": "bob"},
	}

	// "displayName:Alice mail:example" — must match BOTH terms
	result := applySearch(items, "displayName:Alice mail:example")
	assert.Len(t, result, 1)
	assert.Equal(t, "Alice Smith", result[0]["displayName"])
}

func TestApplySearch_MultipleTermsExplicitAND(t *testing.T) {
	items := []map[string]interface{}{
		{"displayName": "Alice Smith", "mail": "alice@example.com", "userPrincipalName": "alice@example.com", "mailNickname": "alice"},
		{"displayName": "Alice Jones", "mail": "alice2@corp.com", "userPrincipalName": "alice2@corp.com", "mailNickname": "alice2"},
	}

	// "displayName:Alice AND mail:corp" — explicit AND
	result := applySearch(items, "displayName:Alice AND mail:corp")
	assert.Len(t, result, 1)
	assert.Equal(t, "Alice Jones", result[0]["displayName"])
}

func TestApplySearch_QuotedInput(t *testing.T) {
	items := []map[string]interface{}{
		{"displayName": "Alice Smith", "userPrincipalName": "alice@example.com", "mail": "alice@example.com", "mailNickname": "alice"},
		{"displayName": "Bob Jones", "userPrincipalName": "bob@example.com", "mail": "bob@example.com", "mailNickname": "bob"},
	}

	// Double-quoted search (as sent by SDK)
	result := applySearch(items, `"Alice"`)
	assert.Len(t, result, 1)
	assert.Equal(t, "Alice Smith", result[0]["displayName"])

	// Field-qualified with quotes
	result = applySearch(items, `"displayName:Alice"`)
	assert.Len(t, result, 1)
	assert.Equal(t, "Alice Smith", result[0]["displayName"])
}

func TestApplySearch_CaseInsensitive(t *testing.T) {
	items := []map[string]interface{}{
		{"displayName": "Alice Smith", "userPrincipalName": "alice@example.com", "mail": "alice@example.com", "mailNickname": "alice"},
	}

	// Search is case-insensitive
	result := applySearch(items, "ALICE")
	assert.Len(t, result, 1)

	result = applySearch(items, "displayName:ALICE")
	assert.Len(t, result, 1)
}

func TestApplySearch_EmptySearch(t *testing.T) {
	items := []map[string]interface{}{
		{"displayName": "Alice"},
	}

	// Empty search returns all items
	result := applySearch(items, "")
	assert.Len(t, result, 1)
}

func TestApplySearch_NoMatch(t *testing.T) {
	items := []map[string]interface{}{
		{"displayName": "Alice", "userPrincipalName": "alice@example.com", "mail": "alice@example.com", "mailNickname": "alice"},
	}

	result := applySearch(items, "Zebra")
	assert.Len(t, result, 0)

	result = applySearch(items, "displayName:Zebra")
	assert.Len(t, result, 0)
}

func TestApplySearch_SubstringMatch(t *testing.T) {
	items := []map[string]interface{}{
		{"displayName": "Alice Smith", "userPrincipalName": "alice.smith@example.com", "mail": "alice.smith@example.com", "mailNickname": "alice"},
	}

	// Partial matches work
	result := applySearch(items, "lice")
	assert.Len(t, result, 1)

	result = applySearch(items, "smith@")
	assert.Len(t, result, 1)
}

// ============================================================================
// $filter — userType eq 'Member' / 'Guest'
// ============================================================================

func TestApplyFilter_UserType(t *testing.T) {
	items := []map[string]interface{}{
		{"userType": "Member", "displayName": "Alice"},
		{"userType": "Guest", "displayName": "Bob"},
		{"userType": "Member", "displayName": "Charlie"},
	}

	result, err := applyFilter(items, "userType eq 'Member'")
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Alice", result[0]["displayName"])
	assert.Equal(t, "Charlie", result[1]["displayName"])

	result, err = applyFilter(items, "userType eq 'Guest'")
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Bob", result[0]["displayName"])

	result, err = applyFilter(items, "userType ne 'Member'")
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Bob", result[0]["displayName"])
}

func TestApplyFilter_UserTypeCombinedWithOtherFilters(t *testing.T) {
	items := []map[string]interface{}{
		{"userType": "Member", "accountEnabled": true, "displayName": "Alice"},
		{"userType": "Member", "accountEnabled": false, "displayName": "Bob"},
		{"userType": "Guest", "accountEnabled": true, "displayName": "Charlie"},
	}

	// userType AND accountEnabled
	result, err := applyFilter(items, "userType eq 'Member' and accountEnabled eq true")
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Alice", result[0]["displayName"])

	// userType OR accountEnabled
	result, err = applyFilter(items, "userType eq 'Guest' or accountEnabled eq false")
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

// ============================================================================
// $filter — any() with nested property paths
// ============================================================================

func TestApplyFilter_AnyNestedPath_SkuId(t *testing.T) {
	type License struct {
		SkuID         string `json:"skuId,omitempty"`
		SkuPartNumber string `json:"skuPartNumber,omitempty"`
	}

	items := []map[string]interface{}{
		{
			"displayName": "Alice",
			"assignedLicenses": []License{
				{SkuID: "sku-1234", SkuPartNumber: "ENTERPRISEPACK"},
				{SkuID: "sku-5678", SkuPartNumber: "EMS"},
			},
		},
		{
			"displayName": "Bob",
			"assignedLicenses": []License{
				{SkuID: "sku-9999", SkuPartNumber: "FLOW_FREE"},
			},
		},
		{
			"displayName": "Charlie",
			"assignedLicenses": []License{},
		},
	}

	// Filter by skuId
	result, err := applyFilter(items, "assignedLicenses/any(a:a/skuId eq 'sku-1234')")
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Alice", result[0]["displayName"])

	// Filter by skuPartNumber
	result, err = applyFilter(items, "assignedLicenses/any(a:a/skuPartNumber eq 'FLOW_FREE')")
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Bob", result[0]["displayName"])

	// Filter that matches none
	result, err = applyFilter(items, "assignedLicenses/any(a:a/skuId eq 'nonexistent')")
	require.NoError(t, err)
	assert.Len(t, result, 0)

	// Empty array — no match
	result, err = applyFilter(items, "assignedLicenses/any(a:a/skuPartNumber eq 'ENTERPRISEPACK')")
	require.NoError(t, err)
	// Alice also has ENTERPRISEPACK, so this should match Alice
	assert.Len(t, result, 1)
}

func TestApplyFilter_AnyFlatStringArray(t *testing.T) {
	items := []map[string]interface{}{
		{"displayName": "Group1", "groupTypes": []string{"Unified"}},
		{"displayName": "Group2", "groupTypes": []string{"Security"}},
		{"displayName": "Group3", "groupTypes": []string{}},
	}

	// Existing flat any() still works
	result, err := applyFilter(items, "groupTypes/any(a:a eq 'Unified')")
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Group1", result[0]["displayName"])
}

func TestApplyFilter_AnyNestedWithAnd(t *testing.T) {
	type License struct {
		SkuID         string `json:"skuId,omitempty"`
		SkuPartNumber string `json:"skuPartNumber,omitempty"`
	}

	items := []map[string]interface{}{
		{
			"displayName":     "Alice",
			"accountEnabled":  true,
			"assignedLicenses": []License{{SkuID: "sku-1234", SkuPartNumber: "ENTERPRISEPACK"}},
		},
		{
			"displayName":     "Bob",
			"accountEnabled":  false,
			"assignedLicenses": []License{{SkuID: "sku-1234", SkuPartNumber: "ENTERPRISEPACK"}},
		},
		{
			"displayName":     "Charlie",
			"accountEnabled":  true,
			"assignedLicenses": []License{{SkuID: "sku-9999", SkuPartNumber: "FLOW_FREE"}},
		},
	}

	// Combine nested any() with a simple filter
	result, err := applyFilter(items, "accountEnabled eq true and assignedLicenses/any(a:a/skuId eq 'sku-1234')")
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Alice", result[0]["displayName"])
}

// ============================================================================
// $filter — assignedLicenses null check
// ============================================================================

func TestApplyFilter_NullAssignedLicenses(t *testing.T) {
	items := []map[string]interface{}{
		{"displayName": "Alice", "assignedLicenses": nil},
		{"displayName": "Bob", "assignedLicenses": []interface{}{}},
	}

	// Filter where assignedLicenses is null
	result, err := applyFilter(items, "assignedLicenses eq null")
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Alice", result[0]["displayName"])
}

// ============================================================================
// mapToStruct tests — type coercion (float64→int)
// ============================================================================

// testItemWithInts is a test-only struct with integer fields
// used to verify mapToStruct handles float64→int conversions safely.
type testItemWithInts struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Count   int    `json:"count,omitempty"`
	Balance int64  `json:"balance,omitempty"`
	Flag    *bool  `json:"flag,omitempty"`
}

func TestMapToStruct_Float64ToIntConversion(t *testing.T) {
	// Simulate a map produced by JSON unmarshaling where numbers are float64
	m := map[string]interface{}{
		"id":      "abc-123",
		"name":    "Test Item",
		"count":   float64(42),
		"balance": float64(999),
	}

	result, err := mapToStruct[testItemWithInts](m)
	require.NoError(t, err)
	assert.Equal(t, "abc-123", result.ID)
	assert.Equal(t, "Test Item", result.Name)
	assert.Equal(t, 42, result.Count)
	assert.Equal(t, int64(999), result.Balance)
}

func TestMapToStruct_NilAndIncompatibleValues(t *testing.T) {
	m := map[string]interface{}{
		"id":    "xyz",
		"name":  "Another",
		"count": "not-a-number", // string targeting int — should be skipped
		"flag":  nil,             // nil for *bool — should be zero
	}

	result, err := mapToStruct[testItemWithInts](m)
	require.NoError(t, err)
	assert.Equal(t, "xyz", result.ID)
	assert.Equal(t, "Another", result.Name)
	assert.Equal(t, 0, result.Count) // zero value — skipped
	assert.Nil(t, result.Flag)        // nil — zeroed
}

func TestApplyOData_SelectThenFilter_NoPanic(t *testing.T) {
	items := []testItemWithInts{
		{ID: "1", Name: "Alpha", Count: 10, Balance: 100},
		{ID: "2", Name: "Beta", Count: 20, Balance: 200},
		{ID: "3", Name: "Gamma", Count: 30, Balance: 300},
	}

	opts := model.ListOptions{
		Filter: "name eq 'Beta'",
		Select: []string{"id", "name"},
	}

	// This must not panic
	result, totalCount, err := ApplyOData(items, opts)
	require.NoError(t, err)
	assert.Equal(t, 1, totalCount)
	require.Len(t, result, 1)
	assert.Equal(t, "2", result[0].ID)
	assert.Equal(t, "Beta", result[0].Name)
}

func TestApplyFilter_NestedParentheses(t *testing.T) {
	items := []map[string]interface{}{
		{"displayName": "Test", "accountEnabled": true},
		{"displayName": "Other", "accountEnabled": true},
		{"displayName": "Third", "accountEnabled": false},
	}

	// Double-nested parentheses with OR inside AND
	// Without the fix, the second loop in parseOr finds the FIRST ')' instead
	// of the matching ')', causing the inner "or" to be detected at the top level.
	result, err := applyFilter(items, "((displayName eq 'Test') or (displayName eq 'Other')) and accountEnabled eq true")
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Test", result[0]["displayName"])
	assert.Equal(t, "Other", result[1]["displayName"])

	// Simple double-nested parens around a single comparison
	result, err = applyFilter(items, "((displayName eq 'Test'))")
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Test", result[0]["displayName"])
}

func TestApplySelect_PreservesId(t *testing.T) {
	item := map[string]interface{}{
		"id":          "abc-123",
		"displayName": "Test User",
		"mail":        "test@example.com",
	}

	result := selectFields(item, []string{"displayName"})

	// id must always be preserved per MS Graph API behavior
	assert.Equal(t, "abc-123", result["id"], "id should always be preserved")
	assert.Equal(t, "Test User", result["displayName"])
	_, hasMail := result["mail"]
	assert.False(t, hasMail, "mail should not be included when not in $select")

	// When id is explicitly in select, it should still be there
	result = selectFields(item, []string{"id", "displayName"})
	assert.Equal(t, "abc-123", result["id"])
	assert.Equal(t, "Test User", result["displayName"])

	// When id is not present in source map, nothing should be added
	itemNoId := map[string]interface{}{
		"displayName": "NoId",
	}
	result = selectFields(itemNoId, []string{"displayName"})
	_, hasId := result["id"]
	assert.False(t, hasId, "id should not be fabricated if not in source")
}

func TestSelectFields_PreservesODataType(t *testing.T) {
	item := map[string]interface{}{
		"@odata.type": "#microsoft.graph.user",
		"id":          "user-123",
		"displayName": "Alice Smith",
		"mail":        "alice@example.com",
	}

	// Select only displayName — @odata.type and id must survive
	result := selectFields(item, []string{"displayName"})

	assert.Equal(t, "#microsoft.graph.user", result["@odata.type"], "@odata.type should always be preserved")
	assert.Equal(t, "user-123", result["id"], "id should always be preserved")
	assert.Equal(t, "Alice Smith", result["displayName"])
	_, hasMail := result["mail"]
	assert.False(t, hasMail, "mail should not be included when not in $select")

	// When @odata.type is not in source map, nothing should be added
	itemNoType := map[string]interface{}{
		"id":          "user-456",
		"displayName": "Bob",
	}
	result = selectFields(itemNoType, []string{"displayName"})
	_, hasType := result["@odata.type"]
	assert.False(t, hasType, "@odata.type should not be fabricated if not in source")
}

func TestApplyOData_Select_PreservesODataType(t *testing.T) {
	users := []model.User{
		{
			ODataType:         "#microsoft.graph.user",
			ID:                "u-1",
			DisplayName:       "Alice",
			UserPrincipalName: "alice@example.com",
			Mail:              "alice@example.com",
		},
		{
			ODataType:         "#microsoft.graph.user",
			ID:                "u-2",
			DisplayName:       "Bob",
			UserPrincipalName: "bob@example.com",
			Mail:              "bob@example.com",
		},
	}

	opts := model.ListOptions{
		Select: []string{"displayName"},
	}

	result, totalCount, err := ApplyOData(users, opts)
	require.NoError(t, err)
	assert.Equal(t, 2, totalCount)
	require.Len(t, result, 2)

	// Verify @odata.type is preserved even though not in $select
	for _, u := range result {
		assert.Equal(t, "#microsoft.graph.user", u.ODataType, "@odata.type must be preserved by $select")
		assert.Equal(t, "user", u.ODataType[len("#microsoft.graph."):]) // sanity check value
	}
}
