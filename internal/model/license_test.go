package model

import (
	"testing"
)

func TestDefaultSubscribedSkus(t *testing.T) {
	skus := DefaultSubscribedSkus()

	if len(skus) == 0 {
		t.Errorf("Expected at least one SKU, got %d", len(skus))
	}

	// Verify the first SKU has required fields
	sku := skus[0]
	if sku.SkuID == "" {
		t.Error("Expected SkuID to be set")
	}
	if sku.SkuPartNumber == "" {
		t.Error("Expected SkuPartNumber to be set")
	}
	if sku.PrepaidUnits == nil {
		t.Error("Expected PrepaidUnits to be set")
	}
	if len(sku.ServicePlans) == 0 {
		t.Error("Expected at least one ServicePlan")
	}
}

func TestFindSkuByPartNumber(t *testing.T) {
	tests := []struct {
		name         string
		skuPartNum   string
		expectFound  bool
		expectSkuID  string
	}{
		{
			name:        "ENTERPRISEPACK",
			skuPartNum:  "ENTERPRISEPACK",
			expectFound: true,
			expectSkuID:  "6fd2c87f-b296-42f0-b197-1e91e994b900",
		},
		{
			name:        "M365_E5",
			skuPartNum:  "M365_E5",
			expectFound: true,
			expectSkuID:  "e8f81a67-4b57-4de3-a716-6110c0012e3d",
		},
		{
			name:        "NONEXISTENT",
			skuPartNum:  "NONEXISTENT",
			expectFound: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			skuId, found := FindSkuByPartNumber(tc.skuPartNum)
			if found != tc.expectFound {
				t.Errorf("Expected found=%v, got %v", tc.expectFound, found)
			}
			if tc.expectFound && skuId != tc.expectSkuID {
				t.Errorf("Expected SKU ID %s, got %s", tc.expectSkuID, skuId)
			}
		})
	}
}

func TestFindSkuBySkuID(t *testing.T) {
	tests := []struct {
		name             string
		skuId            string
		expectFound      bool
		expectPartNumber string
	}{
		{
			name:             "ENTERPRISEPACK SKU ID",
			skuId:            "6fd2c87f-b296-42f0-b197-1e91e994b900",
			expectFound:      true,
			expectPartNumber: "ENTERPRISEPACK",
		},
		{
			name:             "M365_E5 SKU ID",
			skuId:            "e8f81a67-4b57-4de3-a716-6110c0012e3d",
			expectFound:      true,
			expectPartNumber: "M365_E5",
		},
		{
			name:        "Non-existent SKU ID",
			skuId:       "00000000-0000-0000-0000-000000000000",
			expectFound: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			skuPartNumber, found := FindSkuBySkuID(tc.skuId)
			if found != tc.expectFound {
				t.Errorf("Expected found=%v, got %v", tc.expectFound, found)
			}
			if tc.expectFound && skuPartNumber != tc.expectPartNumber {
				t.Errorf("Expected PartNumber %s, got %s", tc.expectPartNumber, skuPartNumber)
			}
		})
	}
}

func TestBidirectionalLookup(t *testing.T) {
	// Test that looking up a SKU by part number and then by SKU ID returns consistent results
	skus := DefaultSubscribedSkus()

	for _, sku := range skus {
		// Look up by part number
		skuId, found := FindSkuByPartNumber(sku.SkuPartNumber)
		if !found {
			t.Errorf("Failed to find SKU by part number: %s", sku.SkuPartNumber)
			continue
		}
		if skuId != sku.SkuID {
			t.Errorf("SKU ID mismatch for %s: expected %s, got %s", sku.SkuPartNumber, sku.SkuID, skuId)
			continue
		}

		// Look up by SKU ID
		skuPartNumber, found := FindSkuBySkuID(sku.SkuID)
		if !found {
			t.Errorf("Failed to find SKU by SKU ID: %s", sku.SkuID)
			continue
		}
		if skuPartNumber != sku.SkuPartNumber {
			t.Errorf("Part number mismatch for %s: expected %s, got %s", sku.SkuID, sku.SkuPartNumber, skuPartNumber)
		}
	}
}
