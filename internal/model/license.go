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
			ODataType: "#microsoft.graph.subscribedSku",
			SkuID: "6fd2c87f-b296-42f0-b197-1e91e994b900",
			SkuPartNumber: "ENTERPRISEPACK",
			CapabilityStatus: "Enabled",
			ConsumedUnits: 0,
			PrepaidUnits: &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo: "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "e212cbc7-0961-4c40-9825-0111774ccef5", ServicePlanName: "EXCHANGE_S_ENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "6634e71b-ee17-40e3-85e0-18d584d0d6c6", ServicePlanName: "SHAREPOINTENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "d42c793f-6c78-4f43-92ca-e8f6a02b035f", ServicePlanName: "MCOSTANDARD", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "b737dad2-0f2c-4270-9b15-1e7e890e0c79", ServicePlanName: "OFFICESUBSCRIPTION", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "5136a095-5cf0-4aff-bec3-e76f5f5cd1bf", ServicePlanName: "INTUNE_O365", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType: "#microsoft.graph.subscribedSku",
			SkuID: "c7df2760-2c81-4ef7-b578-5b5392b571df",
			SkuPartNumber: "ENTERPRISEPREMIUM",
			CapabilityStatus: "Enabled",
			ConsumedUnits: 0,
			PrepaidUnits: &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo: "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "e212cbc7-0961-4c40-9825-0111774ccef5", ServicePlanName: "EXCHANGE_S_ENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "6634e71b-ee17-40e3-85e0-18d584d0d6c6", ServicePlanName: "SHAREPOINTENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "d42c793f-6c78-4f43-92ca-e8f6a02b035f", ServicePlanName: "MCOSTANDARD", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "b737dad2-0f2c-4270-9b15-1e7e890e0c79", ServicePlanName: "OFFICESUBSCRIPTION", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType: "#microsoft.graph.subscribedSku",
			SkuID: "efccb6f7-5641-4e0e-bd10-b4976e1bf68e",
			SkuPartNumber: "EMS",
			CapabilityStatus: "Enabled",
			ConsumedUnits: 0,
			PrepaidUnits: &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo: "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "061f9ace-7d42-4136-88ac-31dc755f143f", ServicePlanName: "INTUNE_A", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "078d2b04-f1bd-4111-bbd4-b4b1b354cef4", ServicePlanName: "AAD_PREMIUM", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "c52ea49f-fe5d-4e95-93ba-1de91d380f89", ServicePlanName: "RIGHTSMANAGEMENT", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType: "#microsoft.graph.subscribedSku",
			SkuID: "b05e124f-c7cc-45a0-a6aa-8cf78c946968",
			SkuPartNumber: "EMSPREMIUM",
			CapabilityStatus: "Enabled",
			ConsumedUnits: 0,
			PrepaidUnits: &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo: "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "061f9ace-7d42-4136-88ac-31dc755f143f", ServicePlanName: "INTUNE_A", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "84a661c4-e949-4bd2-a560-ed7766fcaf2b", ServicePlanName: "AAD_PREMIUM_P2", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "3dd6cf57-d688-4eed-ba52-9e40b5468c3e", ServicePlanName: "THREAT_INTELLIGENCE", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType: "#microsoft.graph.subscribedSku",
			SkuID: "05e9a617-0261-4cee-bb44-138d3ef5d965",
			SkuPartNumber: "SPE_E3",
			CapabilityStatus: "Enabled",
			ConsumedUnits: 0,
			PrepaidUnits: &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo: "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "e212cbc7-0961-4c40-9825-0111774ccef5", ServicePlanName: "EXCHANGE_S_ENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "6634e71b-ee17-40e3-85e0-18d584d0d6c6", ServicePlanName: "SHAREPOINTENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "d42c793f-6c78-4f43-92ca-e8f6a02b035f", ServicePlanName: "MCOSTANDARD", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "b737dad2-0f2c-4270-9b15-1e7e890e0c79", ServicePlanName: "OFFICESUBSCRIPTION", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "5136a095-5cf0-4aff-bec3-e76f5f5cd1bf", ServicePlanName: "INTUNE_O365", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "061f9ace-7d42-4136-88ac-31dc755f143f", ServicePlanName: "INTUNE_A", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "078d2b04-f1bd-4111-bbd4-b4b1b354cef4", ServicePlanName: "AAD_PREMIUM", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType: "#microsoft.graph.subscribedSku",
			SkuID: "06ebc4ee-1bb5-47dd-8120-11324bc54e06",
			SkuPartNumber: "SPE_E5",
			CapabilityStatus: "Enabled",
			ConsumedUnits: 0,
			PrepaidUnits: &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo: "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "e212cbc7-0961-4c40-9825-0111774ccef5", ServicePlanName: "EXCHANGE_S_ENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "6634e71b-ee17-40e3-85e0-18d584d0d6c6", ServicePlanName: "SHAREPOINTENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "d42c793f-6c78-4f43-92ca-e8f6a02b035f", ServicePlanName: "MCOSTANDARD", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "b737dad2-0f2c-4270-9b15-1e7e890e0c79", ServicePlanName: "OFFICESUBSCRIPTION", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "84a661c4-e949-4bd2-a560-ed7766fcaf2b", ServicePlanName: "AAD_PREMIUM_P2", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType: "#microsoft.graph.subscribedSku",
			SkuID: "3b555118-da6a-4418-894f-7df1e2096870",
			SkuPartNumber: "O365_BUSINESS_ESSENTIALS",
			CapabilityStatus: "Enabled",
			ConsumedUnits: 0,
			PrepaidUnits: &LicenseUnits{Enabled: 300, Suspended: 0, Warning: 0},
			AppliesTo: "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "4b9405b0-7788-4568-add1-99614e613b69", ServicePlanName: "EXCHANGESTANDARD", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "1fc08a02-8b3d-43b9-831e-f76859e04e1a", ServicePlanName: "SHAREPOINTSTANDARD", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "d42c793f-6c78-4f43-92ca-e8f6a02b035f", ServicePlanName: "MCOSTANDARD", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType: "#microsoft.graph.subscribedSku",
			SkuID: "f245ecc8-75af-4f8e-b61f-27d8114de5f3",
			SkuPartNumber: "O365_BUSINESS_PREMIUM",
			CapabilityStatus: "Enabled",
			ConsumedUnits: 0,
			PrepaidUnits: &LicenseUnits{Enabled: 300, Suspended: 0, Warning: 0},
			AppliesTo: "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "4b9405b0-7788-4568-add1-99614e613b69", ServicePlanName: "EXCHANGESTANDARD", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "6634e71b-ee17-40e3-85e0-18d584d0d6c6", ServicePlanName: "SHAREPOINTENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "d42c793f-6c78-4f43-92ca-e8f6a02b035f", ServicePlanName: "MCOSTANDARD", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "b737dad2-0f2c-4270-9b15-1e7e890e0c79", ServicePlanName: "OFFICESUBSCRIPTION", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType: "#microsoft.graph.subscribedSku",
			SkuID: "078d2b04-f1bd-4111-bbd4-b4b1b354cef4",
			SkuPartNumber: "AAD_PREMIUM",
			CapabilityStatus: "Enabled",
			ConsumedUnits: 0,
			PrepaidUnits: &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo: "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "078d2b04-f1bd-4111-bbd4-b4b1b354cef4", ServicePlanName: "AAD_PREMIUM", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType: "#microsoft.graph.subscribedSku",
			SkuID: "84a661c4-e949-4bd2-a560-ed7766fcaf2b",
			SkuPartNumber: "AAD_PREMIUM_P2",
			CapabilityStatus: "Enabled",
			ConsumedUnits: 0,
			PrepaidUnits: &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo: "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "84a661c4-e949-4bd2-a560-ed7766fcaf2b", ServicePlanName: "AAD_PREMIUM_P2", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType: "#microsoft.graph.subscribedSku",
			SkuID: "19ec0d23-8335-4cbd-94ac-6050e30712fa",
			SkuPartNumber: "EXCHANGEENTERPRISE",
			CapabilityStatus: "Enabled",
			ConsumedUnits: 0,
			PrepaidUnits: &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo: "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "e212cbc7-0961-4c40-9825-0111774ccef5", ServicePlanName: "EXCHANGE_S_ENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType: "#microsoft.graph.subscribedSku",
			SkuID: "061f9ace-7d42-4136-88ac-31dc755f143f",
			SkuPartNumber: "INTUNE_A",
			CapabilityStatus: "Enabled",
			ConsumedUnits: 0,
			PrepaidUnits: &LicenseUnits{Enabled: 1000, Suspended: 0, Warning: 0},
			AppliesTo: "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "061f9ace-7d42-4136-88ac-31dc755f143f", ServicePlanName: "INTUNE_A", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType: "#microsoft.graph.subscribedSku",
			SkuID: "a403ebcc-fae0-4ca2-8c8c-7a907fd6c235",
			SkuPartNumber: "POWER_BI_STANDARD",
			CapabilityStatus: "Enabled",
			ConsumedUnits: 0,
			PrepaidUnits: &LicenseUnits{Enabled: 10000, Suspended: 0, Warning: 0},
			AppliesTo: "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "a403ebcc-fae0-4ca2-8c8c-7a907fd6c235", ServicePlanName: "POWER_BI_STANDARD", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType: "#microsoft.graph.subscribedSku",
			SkuID: "f8a1db68-be16-40ed-86d5-cb42ce701560",
			SkuPartNumber: "POWER_BI_PRO",
			CapabilityStatus: "Enabled",
			ConsumedUnits: 0,
			PrepaidUnits: &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo: "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "f8a1db68-be16-40ed-86d5-cb42ce701560", ServicePlanName: "POWER_BI_PRO", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType: "#microsoft.graph.subscribedSku",
			SkuID: "44575883-256e-4a79-9da4-ebe9acabe2b2",
			SkuPartNumber: "M365_F1",
			CapabilityStatus: "Enabled",
			ConsumedUnits: 0,
			PrepaidUnits: &LicenseUnits{Enabled: 1000, Suspended: 0, Warning: 0},
			AppliesTo: "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "4b9405b0-7788-4568-add1-99614e613b69", ServicePlanName: "EXCHANGESTANDARD", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "1fc08a02-8b3d-43b9-831e-f76859e04e1a", ServicePlanName: "SHAREPOINTSTANDARD", ProvisioningStatus: "Success", AppliesTo: "User"},
				{ServicePlanID: "d42c793f-6c78-4f43-92ca-e8f6a02b035f", ServicePlanName: "MCOSTANDARD", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType: "#microsoft.graph.subscribedSku",
			SkuID: "a9732ec9-17d9-494c-a51c-d6b45b384dcb",
			SkuPartNumber: "SHAREPOINTENTERPRISE",
			CapabilityStatus: "Enabled",
			ConsumedUnits: 0,
			PrepaidUnits: &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo: "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "a9732ec9-17d9-494c-a51c-d6b45b384dcb", ServicePlanName: "SHAREPOINTENTERPRISE", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType: "#microsoft.graph.subscribedSku",
			SkuID: "111046dd-295b-4d6d-9724-d52ac90bd1f2",
			SkuPartNumber: "WIN_DEF_ATP",
			CapabilityStatus: "Enabled",
			ConsumedUnits: 0,
			PrepaidUnits: &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo: "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "111046dd-295b-4d6d-9724-d52ac90bd1f2", ServicePlanName: "WIN_DEF_ATP", ProvisioningStatus: "Success", AppliesTo: "User"},
			},
		},
		{
			ODataType: "#microsoft.graph.subscribedSku",
			SkuID: "3dd6cf57-d688-4eed-ba52-9e40b5468c3e",
			SkuPartNumber: "THREAT_INTELLIGENCE",
			CapabilityStatus: "Enabled",
			ConsumedUnits: 0,
			PrepaidUnits: &LicenseUnits{Enabled: 100, Suspended: 0, Warning: 0},
			AppliesTo: "User",
			ServicePlans: []ServicePlanInfo{
				{ServicePlanID: "3dd6cf57-d688-4eed-ba52-9e40b5468c3e", ServicePlanName: "THREAT_INTELLIGENCE", ProvisioningStatus: "Success", AppliesTo: "User"},
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

// FindServicePlanID looks up a service plan ID by its skuPartNumber and planName from the default catalog
func FindServicePlanID(skuPartNumber, planName string) (planID string, found bool) {
	for _, sku := range DefaultSubscribedSkus() {
		if sku.SkuPartNumber == skuPartNumber {
			for _, plan := range sku.ServicePlans {
				if plan.ServicePlanName == planName {
					return plan.ServicePlanID, true
				}
			}
		}
	}
	return "", false
}
