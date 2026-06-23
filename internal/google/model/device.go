package model

type ChromeOSDevice struct {
	Kind                 string                `json:"kind"`
	Etag                 string                `json:"etag,omitempty"`
	DeviceID             string                `json:"deviceId"`
	SerialNumber         string                `json:"serialNumber,omitempty"`
	Status               string                `json:"status,omitempty"`
	LastSync             string                `json:"lastSync,omitempty"`
	Model                string                `json:"model,omitempty"`
	OSVersion            string                `json:"osVersion,omitempty"`
	PlatformVersion      string                `json:"platformVersion,omitempty"`
	FirmwareVersion      string                `json:"firmwareVersion,omitempty"`
	MACAddress           string                `json:"macAddress,omitempty"`
	BootMode             string                `json:"bootMode,omitempty"`
	Notes                string                `json:"notes,omitempty"`
	OrderNumber          string                `json:"orderNumber,omitempty"`
	OrgUnitPath          string                `json:"orgUnitPath,omitempty"`
	OrgUnitId            string                `json:"orgUnitId,omitempty"`
	AnnotatedUser        string                `json:"annotatedUser,omitempty"`
	AnnotatedLocation    string                `json:"annotatedLocation,omitempty"`
	AnnotatedAssetId     string                `json:"annotatedAssetId,omitempty"`
	RecentUsers          []UserDate            `json:"recentUsers,omitempty"`
	ActiveTimeRanges     []ActiveTimeRange     `json:"activeTimeRanges,omitempty"`
	EthernetMacAddress   string                `json:"ethernetMacAddress,omitempty"`
	AutoUpdateExpiration string                `json:"autoUpdateExpiration,omitempty"`
	SupportEndDate       string                `json:"supportEndDate,omitempty"`
	TigersHistory        string                `json:"tigersHistory,omitempty"`
	DiskSpaceUsage       string                `json:"diskSpaceUsage,omitempty"`
	CPUStatusReport      []CPUStatusReport     `json:"cpuStatusReport,omitempty"`
	SystemRamFreeReport  []SystemRamFreeReport `json:"systemRamFreeReport,omitempty"`
	LastKnownNetwork     []NetworkInterface    `json:"lastKnownNetwork,omitempty"`
	DeprovisionReason    string                `json:"deprovisionReason,omitempty"`
	WillAutoRenew        bool                  `json:"willAutoRenew"`
	Meid                 string                `json:"meid,omitempty"`
	Manufacturer         string                `json:"manufacturer,omitempty"`
	Sku                  string                `json:"sku,omitempty"`
	CPUModel             string                `json:"cpuModel,omitempty"`
	CPUModelInfo         []CPUModelInfo        `json:"cpuModelInfo,omitempty"`
	DeviceFiles          []DeviceFile          `json:"deviceFiles,omitempty"`
	FeatureLevel         string                `json:"featureLevel,omitempty"`
}

type UserDate struct {
	Email string `json:"email,omitempty"`
	Type  string `json:"type,omitempty"`
}

type ActiveTimeRange struct {
	ActiveTime int64  `json:"activeTime,omitempty"`
	Date       string `json:"date,omitempty"`
}

type CPUStatusReport struct {
	CPUTemperatureInfo []CPUTemperatureInfo `json:"cpuTemperatureInfo,omitempty"`
	CPUUtilizationPct  int64                `json:"cpuUtilizationPct,omitempty"`
	ReportTime         string               `json:"reportTime,omitempty"`
	SampleFrequency    string               `json:"sampleFrequency,omitempty"`
}

type CPUTemperatureInfo struct {
	Label       string `json:"label,omitempty"`
	Temperature int64  `json:"temperature,omitempty"`
}

type CPUModelInfo struct {
	ModelName     string `json:"modelName,omitempty"`
	Architecture  string `json:"architecture,omitempty"`
	MaxClockSpeed int64  `json:"maxClockSpeed,omitempty"`
}

type SystemRamFreeReport struct {
	ReportTime      string `json:"reportTime,omitempty"`
	SystemRamFree   int64  `json:"systemRamFree,omitempty"`
	SampleFrequency string `json:"sampleFrequency,omitempty"`
}

type NetworkInterface struct {
	IpAddress  string `json:"ipAddress,omitempty"`
	MACAddress string `json:"macAddress,omitempty"`
	Type       string `json:"type,omitempty"`
}

type DeviceFile struct {
	Name       string `json:"name,omitempty"`
	Type       string `json:"type,omitempty"`
	Content    string `json:"content,omitempty"`
	CreateTime string `json:"createTime,omitempty"`
}

type ChromeOSCommand struct {
	CommandType string `json:"commandType"`
	Payload     string `json:"payload,omitempty"`
}

type ChromeOSCommandResult struct {
	CommandId   string `json:"commandId,omitempty"`
	CommandType string `json:"commandType,omitempty"`
	DeviceId    string `json:"deviceId,omitempty"`
	State       string `json:"state,omitempty"`
	ExecuteTime string `json:"executeTime,omitempty"`
	Output      string `json:"output,omitempty"`
}

type MobileDevice struct {
	Kind                   string              `json:"kind"`
	Etag                   string              `json:"etag,omitempty"`
	ResourceId             string              `json:"resourceId"`
	DeviceId               string              `json:"deviceId,omitempty"`
	Name                   []string            `json:"name,omitempty"`
	Email                  []string            `json:"email,omitempty"`
	Model                  string              `json:"model,omitempty"`
	Os                     string              `json:"os,omitempty"`
	Type                   string              `json:"type,omitempty"`
	Status                 string              `json:"status,omitempty"`
	HardwareId             string              `json:"hardwareId,omitempty"`
	FirstSync              string              `json:"firstSync,omitempty"`
	LastSync               string              `json:"lastSync,omitempty"`
	UserAgent              string              `json:"userAgent,omitempty"`
	SerialNumber           string              `json:"serialNumber,omitempty"`
	Manufacturer           string              `json:"manufacturer,omitempty"`
	NetworkOperator        string              `json:"networkOperator,omitempty"`
	Meid                   string              `json:"meid,omitempty"`
	Imei                   string              `json:"imei,omitempty"`
	WifiMacAddress         string              `json:"wifiMacAddress,omitempty"`
	BuildNumber            string              `json:"buildNumber,omitempty"`
	DeviceCompromised      string              `json:"deviceCompromised,omitempty"`
	DevicePasswordStatus   string              `json:"devicePasswordStatus,omitempty"`
	ApplicationPermissions []string            `json:"applicationPermissions,omitempty"`
	Applications           []MobileApplication `json:"applications,omitempty"`
	OtherAccountsInfo      []string            `json:"otherAccountsInfo,omitempty"`
	AccountPermissions     []string            `json:"accountPermissions,omitempty"`
	SupportsWorkProfile    bool                `json:"supportsWorkProfile"`
	AdvancedSecurityState  string              `json:"advancedSecurityState,omitempty"`
	EnrollmentTime         string              `json:"enrollmentTime,omitempty"`
	EncryptionStatus       string              `json:"encryptionStatus,omitempty"`
	OsSources              []string            `json:"osSources,omitempty"`
	Privilege              string              `json:"privilege,omitempty"`
	ReleaseVersion         string              `json:"releaseVersion,omitempty"`
	SecurityPatchLevel     string              `json:"securityPatchLevel,omitempty"`
	Brand                  string              `json:"brand,omitempty"`
	KernelVersion          string              `json:"kernelVersion,omitempty"`
	BasebandVersion        string              `json:"basebandVersion,omitempty"`
}

type MobileApplication struct {
	DisplayName string `json:"displayName,omitempty"`
	PackageName string `json:"packageName,omitempty"`
	VersionName string `json:"versionName,omitempty"`
	Permission  string `json:"permission,omitempty"`
}

type MobileDeviceAction struct {
	Action string `json:"action"`
}

type CloudIdentityDevice struct {
	Name                         string                        `json:"name"`
	DeviceType                   string                        `json:"deviceType,omitempty"`
	SerialNumber                 string                        `json:"serialNumber,omitempty"`
	Manufacturer                 string                        `json:"manufacturer,omitempty"`
	Model                        string                        `json:"model,omitempty"`
	OsVersion                    string                        `json:"osVersion,omitempty"`
	AssetTag                     string                        `json:"assetTag,omitempty"`
	DisplayName                  string                        `json:"displayName,omitempty"`
	State                        string                        `json:"state,omitempty"`
	LastSyncTime                 string                        `json:"lastSyncTime,omitempty"`
	OperatingSystem              string                        `json:"operatingSystem,omitempty"`
	Ownership                    string                        `json:"ownership,omitempty"`
	CompromisedState             string                        `json:"compromisedState,omitempty"`
	DeviceId                     string                        `json:"deviceId,omitempty"`
	EncryptionState              string                        `json:"encryptionState,omitempty"`
	Brand                        string                        `json:"brand,omitempty"`
	BuildNumber                  string                        `json:"buildNumber,omitempty"`
	KernelVersion                string                        `json:"kernelVersion,omitempty"`
	BasebandVersion              string                        `json:"basebandVersion,omitempty"`
	SecurityPatchLevel           string                        `json:"securityPatchLevel,omitempty"`
	ManagementState              string                        `json:"managementState,omitempty"`
	IMEI                         string                        `json:"imei,omitempty"`
	MEID                         string                        `json:"meid,omitempty"`
	WifiMacAddresses             []string                      `json:"wifiMacAddresses,omitempty"`
	NetworkInterfaces            []NetworkInterface            `json:"networkInterfaces,omitempty"`
	EndpointVerificationMetadata *EndpointVerificationMetadata `json:"endpointVerificationMetadata,omitempty"`
	CustomAttributes             map[string]interface{}        `json:"customAttributes,omitempty"`
	CreateTime                   string                        `json:"createTime,omitempty"`
	UpdateTime                   string                        `json:"updateTime,omitempty"`
}

type EndpointVerificationMetadata struct {
	BrowserVersion                   string `json:"browserVersion,omitempty"`
	ChromeVersion                    string `json:"chromeVersion,omitempty"`
	IsBrowserManaged                 bool   `json:"isBrowserManaged"`
	IsBuiltInDnsClientEnabled        bool   `json:"isBuiltInDnsClientEnabled"`
	IsChromeRemoteDesktopAppBlocked  bool   `json:"isChromeRemoteDesktopAppBlocked"`
	IsFileDownloadAnalysisEnabled    bool   `json:"isFileDownloadAnalysisEnabled"`
	IsFileUploadAnalysisEnabled      bool   `json:"isFileUploadAnalysisEnabled"`
	IsRealtimeUrlCheckEnabled        bool   `json:"isRealtimeUrlCheckEnabled"`
	IsSecurityEventAnalysisEnabled   bool   `json:"isSecurityEventAnalysisEnabled"`
	IsSiteIsolationEnabled           bool   `json:"isSiteIsolationEnabled"`
	LastDeviceInfoReportTime         string `json:"lastDeviceInfoReportTime,omitempty"`
	LastPolicyFetchTime              string `json:"lastPolicyFetchTime,omitempty"`
	PasswordProtectionWarningTrigger string `json:"passwordProtectionWarningTrigger,omitempty"`
	ProfileEnrollmentDomain          string `json:"profileEnrollmentDomain,omitempty"`
}

type DeviceUser struct {
	Name             string `json:"name"`
	UserEmail        string `json:"userEmail,omitempty"`
	UserType         string `json:"userType,omitempty"`
	ManagementState  string `json:"managementState,omitempty"`
	FirstSyncTime    string `json:"firstSyncTime,omitempty"`
	LastSyncTime     string `json:"lastSyncTime,omitempty"`
	CompromisedState string `json:"compromisedState,omitempty"`
}
