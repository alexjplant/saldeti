package model

// SubscribedSku represents a license SKU available in the tenant (from GET /subscribedSkus)
type SubscribedSku struct {
	ODataType       string           `json:"@odata.type,omitempty"`
	SkuID           string           `json:"skuId,omitempty"`
	SkuPartNumber   string           `json:"skuPartNumber,omitempty"`
	CapabilityStatus string          `json:"capabilityStatus,omitempty"`
	ConsumedUnits   int              `json:"consumedUnits,omitempty"`
	PrepaidUnits    *LicenseUnits    `json:"prepaidUnits,omitempty"`
	ServicePlans    []ServicePlanInfo `json:"servicePlans,omitempty"`
	AppliesTo       string           `json:"appliesTo,omitempty"`
}

type LicenseUnits struct {
	Enabled   int `json:"enabled"`
	Suspended int `json:"suspended"`
	Warning   int `json:"warning"`
}

type ServicePlanInfo struct {
	ServicePlanID      string `json:"servicePlanId,omitempty"`
	ServicePlanName    string `json:"servicePlanName,omitempty"`
	ProvisioningStatus string `json:"provisioningStatus,omitempty"`
	AppliesTo          string `json:"appliesTo,omitempty"`
}

// LicenseAssignmentRequest is the request body for POST /users/{id}/assignLicense
type LicenseAssignmentRequest struct {
	AddLicenses    []LicenseAssignment `json:"addLicenses"`
	RemoveLicenses []LicenseRemoval    `json:"removeLicenses"`
}

type LicenseAssignment struct {
	SkuID         string   `json:"skuId,omitempty"`
	DisabledPlans []string `json:"disabledPlans,omitempty"`
}

type LicenseRemoval struct {
	SkuID string `json:"skuId,omitempty"`
}

// SeedLicense is used in seed JSON for user license assignment
type SeedLicense struct {
	SkuPartNumber string   `json:"skuPartNumber"`
	DisabledPlans []string `json:"disabledPlans,omitempty"`
}

// DefaultSubscribedSkus returns the static catalog of well-known Microsoft 365 SKUs
func DefaultSubscribedSkus() []SubscribedSku {
	return []SubscribedSku{
		{
			ODataType:       "#microsoft.graph.subscribedSku",
			SkuID:           "6fd2c87f-b296-42f0-b197-1e91e994b900",
			SkuPartNumber:   "ENTERPRISEPACK",
			CapabilityStatus: "Enabled",
			ConsumedUnits:   0,
			PrepaidUnits:    &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo:       "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "e212cbc7-0961-4c40-9825-0111774ccef5", ServicePlanName: "EXCHANGE_S_ENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "6634e71b-ee17-40e3-85e0-18d584d0d6c6", ServicePlanName: "SHAREPOINTENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "0feaeb32-d00e-4d66-bd5a-43b5b83db82c", ServicePlanName: "MCOSTANDARD", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "b737dad2-0f2c-4270-9b15-1e7e890e0c79", ServicePlanName: "OFFICESUBSCRIPTION", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "5136a095-5cf0-4aff-bec3-e76f5f5cd1bf", ServicePlanName: "INTUNE_O365", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType:       "#microsoft.graph.subscribedSku",
			SkuID:           "c7df2760-2c81-4ef7-b578-5b5392b571df",
			SkuPartNumber:   "ENTERPRISEPREMIUM",
			CapabilityStatus: "Enabled",
			ConsumedUnits:   0,
			PrepaidUnits:    &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo:       "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "e212cbc7-0961-4c40-9825-0111774ccef5", ServicePlanName: "EXCHANGE_S_ENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "6634e71b-ee17-40e3-85e0-18d584d0d6c6", ServicePlanName: "SHAREPOINTENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "0feaeb32-d00e-4d66-bd5a-43b5b83db82c", ServicePlanName: "MCOSTANDARD", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "b737dad2-0f2c-4270-9b15-1e7e890e0c79", ServicePlanName: "OFFICESUBSCRIPTION", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType:       "#microsoft.graph.subscribedSku",
			SkuID:           "efccb6f7-5641-4e0e-bd10-b4976e1bf68e",
			SkuPartNumber:   "EMS",
			CapabilityStatus: "Enabled",
			ConsumedUnits:   0,
			PrepaidUnits:    &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo:       "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "c22ec197-13eb-4e5a-89a2-9fe2e4a5b37d", ServicePlanName: "INTUNE_A", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "4b81e5c9-4b9e-450e-9004-28c5d6d7f384", ServicePlanName: "AAD_PREMIUM", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "8b83667d-1a44-4e9e-bd8e-8272ef49d6e6", ServicePlanName: "RMS_S_PREMIUM", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType:       "#microsoft.graph.subscribedSku",
			SkuID:           "b05e124f-c7cc-45a0-a6aa-89778aceada2",
			SkuPartNumber:   "EMS_E5",
			CapabilityStatus: "Enabled",
			ConsumedUnits:   0,
			PrepaidUnits:    &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo:       "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "c22ec197-13eb-4e5a-89a2-9fe2e4a5b37d", ServicePlanName: "INTUNE_A", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "eec0eb4f-6444-4e2d-bf5c-544ed1a147c5", ServicePlanName: "AAD_PREMIUM_P2", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "e7beca80-9b89-4e27-8c62-e1c5b894e9d0", ServicePlanName: "ADALLOM_S_O365", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType:       "#microsoft.graph.subscribedSku",
			SkuID:           "1fc78016-a0e7-4f5f-b8d1-4e898572eeb4",
			SkuPartNumber:   "M365_E3",
			CapabilityStatus: "Enabled",
			ConsumedUnits:   0,
			PrepaidUnits:    &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo:       "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "e212cbc7-0961-4c40-9825-0111774ccef5", ServicePlanName: "EXCHANGE_S_ENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "6634e71b-ee17-40e3-85e0-18d584d0d6c6", ServicePlanName: "SHAREPOINTENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "0feaeb32-d00e-4d66-bd5a-43b5b83db82c", ServicePlanName: "MCOSTANDARD", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "b737dad2-0f2c-4270-9b15-1e7e890e0c79", ServicePlanName: "OFFICESUBSCRIPTION", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "5136a095-5cf0-4aff-bec3-e76f5f5cd1bf", ServicePlanName: "INTUNE_O365", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "c22ec197-13eb-4e5a-89a2-9fe2e4a5b37d", ServicePlanName: "INTUNE_A", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "4b81e5c9-4b9e-450e-9004-28c5d6d7f384", ServicePlanName: "AAD_PREMIUM", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType:       "#microsoft.graph.subscribedSku",
			SkuID:           "e8f81a67-4b57-4de3-a716-6110c0012e3d",
			SkuPartNumber:   "M365_E5",
			CapabilityStatus: "Enabled",
			ConsumedUnits:   0,
			PrepaidUnits:    &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo:       "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "e212cbc7-0961-4c40-9825-0111774ccef5", ServicePlanName: "EXCHANGE_S_ENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "6634e71b-ee17-40e3-85e0-18d584d0d6c6", ServicePlanName: "SHAREPOINTENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "0feaeb32-d00e-4d66-bd5a-43b5b83db82c", ServicePlanName: "MCOSTANDARD", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "b737dad2-0f2c-4270-9b15-1e7e890e0c79", ServicePlanName: "OFFICESUBSCRIPTION", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "eec0eb4f-6444-4e2d-bf5c-544ed1a147c5", ServicePlanName: "AAD_PREMIUM_P2", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType:       "#microsoft.graph.subscribedSku",
			SkuID:           "3b555118-da6a-4418-894f-7df1e2096870",
			SkuPartNumber:   "O365_BUSINESS_ESSENTIALS",
			CapabilityStatus: "Enabled",
			ConsumedUnits:   0,
			PrepaidUnits:    &LicenseUnits{Enabled: 300, Suspended: 0, Warning: 0},
			AppliesTo:       "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "3385bc5d-1401-4962-bd29-7b6994fe12e6", ServicePlanName: "EXCHANGE_S_FOUNDATION", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "2b9c8e7c-319c-43e2-8198-2c6e5f5383d1", ServicePlanName: "SHAREPOINTDESKLESS", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "0feaeb32-d00e-4d66-bd5a-43b5b83db82c", ServicePlanName: "MCOSTANDARD", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType:       "#microsoft.graph.subscribedSku",
			SkuID:           "531ee2f8-bb36-4b62-9f87-9cbd920c0c13",
			SkuPartNumber:   "O365_BUSINESS_PREMIUM",
			CapabilityStatus: "Enabled",
			ConsumedUnits:   0,
			PrepaidUnits:    &LicenseUnits{Enabled: 300, Suspended: 0, Warning: 0},
			AppliesTo:       "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "3385bc5d-1401-4962-bd29-7b6994fe12e6", ServicePlanName: "EXCHANGE_S_FOUNDATION", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "6634e71b-ee17-40e3-85e0-18d584d0d6c6", ServicePlanName: "SHAREPOINTENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "0feaeb32-d00e-4d66-bd5a-43b5b83db82c", ServicePlanName: "MCOSTANDARD", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "b737dad2-0f2c-4270-9b15-1e7e890e0c79", ServicePlanName: "OFFICESUBSCRIPTION", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType:       "#microsoft.graph.subscribedSku",
			SkuID:           "4b81e5c9-4b9e-450e-9004-28c5d6d7f384",
			SkuPartNumber:   "AAD_PREMIUM",
			CapabilityStatus: "Enabled",
			ConsumedUnits:   0,
			PrepaidUnits:    &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo:       "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "4b81e5c9-4b9e-450e-9004-28c5d6d7f384", ServicePlanName: "AAD_PREMIUM", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType:       "#microsoft.graph.subscribedSku",
			SkuID:           "eec0eb4f-6444-4e2d-bf5c-544ed1a147c5",
			SkuPartNumber:   "AAD_PREMIUM_P2",
			CapabilityStatus: "Enabled",
			ConsumedUnits:   0,
			PrepaidUnits:    &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo:       "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "eec0eb4f-6444-4e2d-bf5c-544ed1a147c5", ServicePlanName: "AAD_PREMIUM_P2", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType:       "#microsoft.graph.subscribedSku",
			SkuID:           "ee02fd1b-340e-4a4b-bcca-11cb1d7c7d3c",
			SkuPartNumber:   "EXCHANGEENTERPRISE",
			CapabilityStatus: "Enabled",
			ConsumedUnits:   0,
			PrepaidUnits:    &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo:       "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "e212cbc7-0961-4c40-9825-0111774ccef5", ServicePlanName: "EXCHANGE_S_ENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType:       "#microsoft.graph.subscribedSku",
			SkuID:           "c22ec197-13eb-4e5a-89a2-9fe2e4a5b37d",
			SkuPartNumber:   "INTUNE_A",
			CapabilityStatus: "Enabled",
			ConsumedUnits:   0,
			PrepaidUnits:    &LicenseUnits{Enabled: 1000, Suspended: 0, Warning: 0},
			AppliesTo:       "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "c22ec197-13eb-4e5a-89a2-9fe2e4a5b37d", ServicePlanName: "INTUNE_A", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType:       "#microsoft.graph.subscribedSku",
			SkuID:           "a403ebcc-fae0-4ca2-8c8c-7a907fd6c235",
			SkuPartNumber:   "POWER_BI_STANDARD",
			CapabilityStatus: "Enabled",
			ConsumedUnits:   0,
			PrepaidUnits:    &LicenseUnits{Enabled: 10000, Suspended: 0, Warning: 0},
			AppliesTo:       "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "a403ebcc-fae0-4ca2-8c8c-7a907fd6c235", ServicePlanName: "POWER_BI_STANDARD", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType:       "#microsoft.graph.subscribedSku",
			SkuID:           "f8a1db68-be16-43ed-842f-16655c5a4107",
			SkuPartNumber:   "POWER_BI_PRO",
			CapabilityStatus: "Enabled",
			ConsumedUnits:   0,
			PrepaidUnits:    &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo:       "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "f8a1db68-be16-43ed-842f-16655c5a4107", ServicePlanName: "POWER_BI_PRO", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType:       "#microsoft.graph.subscribedSku",
			SkuID:           "66b55226-2e6f-4a24-9024-43aa496726ca",
			SkuPartNumber:   "M365_F1",
			CapabilityStatus: "Enabled",
			ConsumedUnits:   0,
			PrepaidUnits:    &LicenseUnits{Enabled: 1000, Suspended: 0, Warning: 0},
			AppliesTo:       "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "3385bc5d-1401-4962-bd29-7b6994fe12e6", ServicePlanName: "EXCHANGE_S_FOUNDATION", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "2b9c8e7c-319c-43e2-8198-2c6e5f5383d1", ServicePlanName: "SHAREPOINTDESKLESS", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "0feaeb32-d00e-4d66-bd5a-43b5b83db82c", ServicePlanName: "MCOSTANDARD", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType:       "#microsoft.graph.subscribedSku",
			SkuID:           "a9732ec9-b17f-4a6d-a6b8-76d0e8202c68",
			SkuPartNumber:   "SHAREPOINTENTERPRISE",
			CapabilityStatus: "Enabled",
			ConsumedUnits:   0,
			PrepaidUnits:    &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo:       "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "6634e71b-ee17-40e3-85e0-18d584d0d6c6", ServicePlanName: "SHAREPOINTENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType:       "#microsoft.graph.subscribedSku",
			SkuID:           "3d0b7e4b-5b46-4b5d-8cd5-3d10c2afe1e8",
			SkuPartNumber:   "WINDOWS_DEFENDER_ATP",
			CapabilityStatus: "Enabled",
			ConsumedUnits:   0,
			PrepaidUnits:    &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo:       "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "3d0b7e4b-5b46-4b5d-8cd5-3d10c2afe1e8", ServicePlanName: "WINDOWS_DEFENDER_ATP", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType:       "#microsoft.graph.subscribedSku",
			SkuID:           "1e8965e0-5a3e-4de3-a716-6110c0012e3d",
			SkuPartNumber:   "ADALLOM_S_O365",
			CapabilityStatus: "Enabled",
			ConsumedUnits:   0,
			PrepaidUnits:    &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo:       "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "1e8965e0-5a3e-4de3-a716-6110c0012e3d", ServicePlanName: "ADALLOM_S_O365", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
	}
}

// FindSkuByPartNumber looks up a SkuID by its skuPartNumber from the default catalog
func FindSkuByPartNumber(skuPartNumber string) (skuId string, found bool) {
	for _, sku := range DefaultSubscribedSkus() {
		if sku.SkuPartNumber == skuPartNumber {
			return sku.SkuID, true
		}
	}
	return "", false
}

// FindSkuBySkuID looks up a skuPartNumber by its skuId from the default catalog
func FindSkuBySkuID(skuId string) (skuPartNumber string, found bool) {
	for _, sku := range DefaultSubscribedSkus() {
		if sku.SkuID == skuId {
			return sku.SkuPartNumber, true
		}
	}
	return "", false
}
