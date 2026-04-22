package store

import (
	"testing"

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
