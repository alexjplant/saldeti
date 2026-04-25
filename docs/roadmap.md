# Entra Graph API Simulator — Tiered Implementation Plan

> **Base URL:** `https://graph.microsoft.com/v1.0`  
> **Auth:** OAuth 2.0 (client_credentials, authorization_code, on_behalf_of)  
> **Common query params:** `$filter`, `$select`, `$expand`, `$top`, `$orderby`, `$count`, `$skip`

---

## Tier 1 — Identity Core (Users & Groups)

*The minimum viable simulator. Every Entra integration starts here.*

### 1A. Authentication Layer
| # | Method | Path | Description |
|---|--------|------|-------------|
| 1 | `POST` | `https://login.microsoftonline.com/{tenant}/oauth2/v2.0/token` | Client-credentials & auth-code token exchange |
| 2 | — | — | Token validation middleware (decode JWT, check `scp`/`roles`, expiry) |

### 1B. Users
| # | Method | Path | Description |
|---|--------|------|-------------|
| 3 | `GET` | `/users` | List users (supports `$filter`, `$select`, `$top`, `$orderby`, `$count`) |
| 4 | `GET` | `/users/{id}` | Get a user by ID or UPN |
| 5 | `POST` | `/users` | Create a user |
| 6 | `PATCH` | `/users/{id}` | Update a user |
| 7 | `DELETE` | `/users/{id}` | Delete a user |
| 8 | `GET` | `/users/{id}/memberOf` | Get groups and directory roles the user is a direct member of |
| 9 | `GET` | `/users/{id}/transitiveMemberOf` | Get groups and directory roles (transitive/nested) |
| 10 | `GET` | `/users/{id}/directReports` | Get direct reports |
| 11 | `GET` | `/users/{id}/manager` | Get the user's manager |
| 12 | `PUT` | `/users/{id}/manager/$ref` | Assign a manager |
| 13 | `GET` | `/users/{id}/photo` | Get user photo (metadata) |
| 14 | `GET` | `/users/{id}/photo/$value` | Get user photo (binary) |
| 15 | `PATCH` | `/users/{id}/photo/$value` | Update user photo |
| 16 | `POST` | `/users/{id}/changePassword` | Change password |
| 17 | `POST` | `/users/{id}/reprocessLicenseAssignment` | Reprocess license assignments |
| 18 | `GET` | `/users/{id}/licenseDetails` | Get license details |
| 19 | `GET` | `/users/{id}/appRoleAssignments` | List app role assignments for user |
| 20 | `POST` | `/users/{id}/appRoleAssignments` | Assign an app role to user |

### 1C. Groups
| # | Method | Path | Description |
|---|--------|------|-------------|
| 21 | `GET` | `/groups` | List groups (supports `$filter`, `$select`, `$top`, `$count`) |
| 22 | `GET` | `/groups/{id}` | Get a group by ID |
| 23 | `POST` | `/groups` | Create a group (security, Microsoft 365, mail-enabled) |
| 24 | `PATCH` | `/groups/{id}` | Update a group |
| 25 | `DELETE` | `/groups/{id}` | Delete a group |
| 26 | `GET` | `/groups/{id}/members` | List direct members |
| 27 | `POST` | `/groups/{id}/members/$ref` | Add a member |
| 28 | `DELETE` | `/groups/{id}/members/{memberId}/$ref` | Remove a member |
| 29 | `GET` | `/groups/{id}/transitiveMembers` | List transitive (nested) members |
| 30 | `GET` | `/groups/{id}/owners` | List owners |
| 31 | `POST` | `/groups/{id}/owners/$ref` | Add an owner |
| 32 | `DELETE` | `/groups/{id}/owners/{ownerId}/$ref` | Remove an owner |
| 33 | `GET` | `/groups/{id}/memberOf` | Groups this group is a member of |
| 34 | `GET` | `/groups/{id}/transitiveMemberOf` | Transitive memberOf |
| 35 | `POST` | `/groups/{id}/checkMemberGroups` | Check membership (returns group IDs) |
| 36 | `POST` | `/groups/{id}/getMemberGroups` | Return all groups the group is a member of (transitive) |
| 37 | `POST` | `/groups/{id}/getMemberObjects` | Return all groups, roles, admin units (transitive) |
| 38 | `GET` | `/groups/{id}/appRoleAssignments` | List app role assignments for group |

---

## Tier 2 — Application & Service Principal Registry

*Required for any scenario involving app registrations, API permissions, or service principals.*

### 2A. Applications (App Registrations)
| # | Method | Path | Description |
|---|--------|------|-------------|
| 39 | `GET` | `/applications` | List applications |
| 40 | `GET` | `/applications/{id}` | Get application by ID |
| 41 | `GET` | `/applications(appId='{appId}')` | Get application by appId (alternate key) |
| 42 | `POST` | `/applications` | Create an application |
| 43 | `PATCH` | `/applications/{id}` | Update an application |
| 44 | `DELETE` | `/applications/{id}` | Delete an application |
| 45 | `POST` | `/applications/{id}/addPassword` | Add a password credential |
| 46 | `POST` | `/applications/{id}/removePassword` | Remove a password credential |
| 47 | `POST` | `/applications/{id}/addKey` | Add a key credential |
| 48 | `POST` | `/applications/{id}/removeKey` | Remove a key credential |
| 49 | `GET` | `/applications/{id}/owners` | List owners |
| 50 | `POST` | `/applications/{id}/owners/$ref` | Add an owner |
| 51 | `DELETE` | `/applications/{id}/owners/{ownerId}/$ref` | Remove an owner |
| 52 | `GET` | `/applications/{id}/extensionProperties` | List extension properties |
| 53 | `POST` | `/applications/{id}/extensionProperties` | Create extension property |
| 54 | `DELETE` | `/applications/{id}/extensionProperties/{extId}` | Delete extension property |
| 55 | `POST` | `/applications/{id}/setVerifiedPublisher` | Set verified publisher |
| 56 | `GET` | `/applications/delta` | Delta query for applications |

### 2B. Service Principals (Enterprise Apps)
| # | Method | Path | Description |
|---|--------|------|-------------|
| 57 | `GET` | `/servicePrincipals` | List service principals |
| 58 | `GET` | `/servicePrincipals/{id}` | Get service principal by ID |
| 59 | `GET` | `/servicePrincipals(appId='{appId}')` | Get by appId (alternate key) |
| 60 | `POST` | `/servicePrincipals` | Create a service principal |
| 61 | `PATCH` | `/servicePrincipals/{id}` | Update a service principal |
| 62 | `DELETE` | `/servicePrincipals/{id}` | Delete a service principal |
| 63 | `GET` | `/servicePrincipals/{id}/owners` | List owners |
| 64 | `POST` | `/servicePrincipals/{id}/owners/$ref` | Add an owner |
| 65 | `DELETE` | `/servicePrincipals/{id}/owners/{ownerId}/$ref` | Remove an owner |
| 66 | `GET` | `/servicePrincipals/{id}/memberOf` | Groups the SP is a member of |
| 67 | `GET` | `/servicePrincipals/{id}/transitiveMemberOf` | Transitive memberOf |
| 68 | `GET` | `/servicePrincipals/{id}/appRoleAssignments` | List app roles assigned to this SP |
| 69 | `POST` | `/servicePrincipals/{id}/appRoleAssignments` | Assign an app role to this SP |
| 70 | `DELETE` | `/servicePrincipals/{id}/appRoleAssignments/{assignmentId}` | Remove an app role assignment |
| 71 | `GET` | `/servicePrincipals/{id}/appRoleAssignedTo` | List principals assigned app roles of this SP |
| 72 | `POST` | `/servicePrincipals/{id}/appRoleAssignedTo` | Assign an app role to a principal for this SP |
| 73 | `DELETE` | `/servicePrincipals/{id}/appRoleAssignedTo/{assignmentId}` | Remove |
| 74 | `GET` | `/servicePrincipals/{id}/oauth2PermissionGrants` | List delegated permission grants |
| 75 | `POST` | `/servicePrincipals/{id}/addPassword` | Add a password credential |
| 76 | `POST` | `/servicePrincipals/{id}/removePassword` | Remove a password credential |
| 77 | `POST` | `/servicePrincipals/{id}/addKey` | Add a key credential |
| 78 | `POST` | `/servicePrincipals/{id}/removeKey` | Remove a key credential |
| 79 | `GET` | `/servicePrincipals/{id}/homeRealmDiscoveryPolicies` | List HRD policies |
| 80 | `GET` | `/servicePrincipals/{id}/claimsMappingPolicies` | List claims mapping policies |
| 81 | `GET` | `/servicePrincipals/{id}/tokenIssuancePolicies` | List token issuance policies |
| 82 | `GET` | `/servicePrincipals/{id}/tokenLifetimePolicies` | List token lifetime policies |

### 2C. OAuth2 Permission Grants
| # | Method | Path | Description |
|---|--------|------|-------------|
| 83 | `GET` | `/oauth2PermissionGrants` | List delegated permission grants (OAuth2 consent) |
| 84 | `GET` | `/oauth2PermissionGrants/{id}` | Get a specific grant |
| 85 | `POST` | `/oauth2PermissionGrants` | Create a delegated permission grant |
| 86 | `PATCH` | `/oauth2PermissionGrants/{id}` | Update a grant |
| 87 | `DELETE` | `/oauth2PermissionGrants/{id}` | Delete a grant |

### 2D. App Role Assignments (global)
| # | Method | Path | Description |
|---|--------|------|-------------|
| 88 | `GET` | `/users/{id}/appRoleAssignments` | (repeated from 1B for completeness) |
| 89 | `POST` | `/users/{id}/appRoleAssignments` | Assign app role to user |
| 90 | `DELETE` | `/users/{id}/appRoleAssignments/{assignmentId}` | Remove |
| 91 | `GET` | `/groups/{id}/appRoleAssignments` | List group app role assignments |
| 92 | `POST` | `/groups/{id}/appRoleAssignments` | Assign app role to group |
| 93 | `DELETE` | `/groups/{id}/appRoleAssignments/{assignmentId}` | Remove |

---

## Tier 3 — Directory Roles & Administrative Units

*Essential for RBAC simulation and organizational structure.*

### 3A. Directory Roles
| # | Method | Path | Description |
|---|--------|------|-------------|
| 94 | `GET` | `/directoryRoles` | List activated directory roles |
| 95 | `GET` | `/directoryRoles/{id}` | Get a directory role |
| 96 | `GET` | `/directoryRoles/roleTemplateId={templateId}` | Get by role template ID |
| 97 | `POST` | `/directoryRoles` | Activate a directory role (from role template) |
| 98 | `GET` | `/directoryRoles/{id}/members` | List members of a role |
| 99 | `POST` | `/directoryRoles/{id}/members/$ref` | Add a member to a role |
| 100 | `DELETE` | `/directoryRoles/{id}/members/{memberId}/$ref` | Remove a member from a role |
| 101 | `GET` | `/directoryRoles/{id}/memberOf` | Groups the role is a member of |
| 102 | `GET` | `/directoryRoles/{id}/scopedMembers` | List scoped members (admin unit scoped) |

### 3B. Role Management (Unified RBAC)
| # | Method | Path | Description |
|---|--------|------|-------------|
| 103 | `GET` | `/roleManagement/directory/roleDefinitions` | List role definitions |
| 104 | `GET` | `/roleManagement/directory/roleDefinitions/{id}` | Get a role definition |
| 105 | `POST` | `/roleManagement/directory/roleDefinitions` | Create a custom role |
| 106 | `PATCH` | `/roleManagement/directory/roleDefinitions/{id}` | Update a role definition |
| 107 | `DELETE` | `/roleManagement/directory/roleDefinitions/{id}` | Delete a role definition |
| 108 | `GET` | `/roleManagement/directory/roleAssignments` | List role assignments |
| 109 | `GET` | `/roleManagement/directory/roleAssignments/{id}` | Get a role assignment |
| 110 | `POST` | `/roleManagement/directory/roleAssignments` | Create a role assignment |
| 111 | `DELETE` | `/roleManagement/directory/roleAssignments/{id}` | Delete a role assignment |
| 112 | `GET` | `/roleManagement/directory/roleAssignmentScheduleInstances` | List active assignments (PIM) |
| 113 | `GET` | `/roleManagement/directory/roleEligibilityScheduleInstances` | List eligible assignments (PIM) |
| 114 | `POST` | `/roleManagement/directory/roleAssignmentScheduleRequests` | Request role activation (PIM) |

### 3C. Administrative Units
| # | Method | Path | Description |
|---|--------|------|-------------|
| 115 | `GET` | `/administrativeUnits` | List administrative units |
| 116 | `GET` | `/administrativeUnits/{id}` | Get an admin unit |
| 117 | `POST` | `/administrativeUnits` | Create an admin unit |
| 118 | `PATCH` | `/administrativeUnits/{id}` | Update an admin unit |
| 119 | `DELETE` | `/administrativeUnits/{id}` | Delete an admin unit |
| 120 | `GET` | `/administrativeUnits/{id}/members` | List members |
| 121 | `POST` | `/administrativeUnits/{id}/members/$ref` | Add a member |
| 122 | `DELETE` | `/administrativeUnits/{id}/members/{memberId}/$ref` | Remove a member |
| 123 | `GET` | `/administrativeUnits/{id}/scopedRoleMemberships` | List scoped role memberships |
| 124 | `POST` | `/administrativeUnits/{id}/scopedRoleMemberships` | Add scoped role membership |
| 125 | `DELETE` | `/administrativeUnits/{id}/scopedRoleMemberships/{id}` | Remove scoped role membership |

### 3D. Directory Objects (cross-cutting)
| # | Method | Path | Description |
|---|--------|------|-------------|
| 126 | `POST` | `/directoryObjects/getByIds` | Get objects by IDs (batch lookup) |
| 127 | `POST` | `/directoryObjects/getUserOwnedObjects` | Get objects owned by a user |
| 128 | `POST` | `/directoryObjects/getUserOwnedDevices` | Get devices owned by a user |
| 129 | `POST` | `/users/{id}/checkMemberGroups` | Check if user is member of groups |
| 130 | `POST` | `/users/{id}/getMemberGroups` | Get all group memberships (transitive) |
| 131 | `POST` | `/users/{id}/getMemberObjects` | Get all directory objects (transitive) |
| 132 | `POST` | `/directoryObjects/validateProperties` | Validate a display name or UPN |
| 133 | `GET` | `/directoryObjects/delta` | Delta query |
| 134 | `GET` | `/users/delta` | Delta query for users |
| 135 | `GET` | `/groups/delta` | Delta query for groups |

---

## Tier 4 — Organization, Domains & Devices

*Tenant-level configuration and device management.*

### 4A. Organization
| # | Method | Path | Description |
|---|--------|------|-------------|
| 136 | `GET` | `/organization` | Get organization details (tenant info) |
| 137 | `GET` | `/organization/{id}` | Get organization by ID |
| 138 | `PATCH` | `/organization/{id}` | Update organization |
| 139 | `GET` | `/organization/{id}/branding` | Get organizational branding |
| 140 | `PATCH` | `/organization/{id}/branding` | Update branding |
| 141 | `GET` | `/organization/{id}/branding/localizations` | List branding localizations |
| 142 | `GET` | `/organization/{id}/certificateBasedAuthConfiguration` | Get CBA config |
| 143 | `POST` | `/organization/{id}/certificateBasedAuthConfiguration` | Create CBA config |

### 4B. Domains
| # | Method | Path | Description |
|---|--------|------|-------------|
| 144 | `GET` | `/domains` | List domains |
| 145 | `GET` | `/domains/{id}` | Get a domain |
| 146 | `POST` | `/domains` | Create a domain |
| 147 | `PATCH` | `/domains/{id}` | Update a domain |
| 148 | `DELETE` | `/domains/{id}` | Delete a domain |
| 149 | `GET` | `/domains/{id}/domainNameReferences` | List objects referencing the domain |
| 150 | `GET` | `/domains/{id}/serviceConfigurationRecords` | Get DNS records (service config) |
| 151 | `GET` | `/domains/{id}/verificationDnsRecords` | Get DNS records (verification) |
| 152 | `POST` | `/domains/{id}/verify` | Verify a domain |
| 153 | `POST` | `/domains/{id}/forceDelete` | Force delete a domain |
| 154 | `POST` | `/domains/{id}/promote` | Promote verified subdomain |

### 4C. Devices
| # | Method | Path | Description |
|---|--------|------|-------------|
| 155 | `GET` | `/devices` | List devices |
| 156 | `GET` | `/devices/{id}` | Get a device |
| 157 | `POST` | `/devices` | Register a device |
| 158 | `PATCH` | `/devices/{id}` | Update a device |
| 159 | `DELETE` | `/devices/{id}` | Delete a device |
| 160 | `GET` | `/devices/{id}/memberOf` | Groups the device is a member of |
| 161 | `GET` | `/devices/{id}/transitiveMemberOf` | Transitive memberOf |
| 162 | `GET` | `/devices/{id}/registeredUsers` | List registered users |
| 163 | `GET` | `/devices/{id}/registeredOwners` | List registered owners |

---

## Tier 5 — Conditional Access & Identity Protection

*Security policy simulation — important for testing access control logic.*

### 5A. Conditional Access
| # | Method | Path | Description |
|---|--------|------|-------------|
| 164 | `GET` | `/identity/conditionalAccess/policies` | List conditional access policies |
| 165 | `GET` | `/identity/conditionalAccess/policies/{id}` | Get a policy |
| 166 | `POST` | `/identity/conditionalAccess/policies` | Create a policy |
| 167 | `PATCH` | `/identity/conditionalAccess/policies/{id}` | Update a policy |
| 168 | `DELETE` | `/identity/conditionalAccess/policies/{id}` | Delete a policy |
| 169 | `GET` | `/identity/conditionalAccess/namedLocations` | List named locations |
| 170 | `GET` | `/identity/conditionalAccess/namedLocations/{id}` | Get a named location |
| 171 | `POST` | `/identity/conditionalAccess/namedLocations` | Create a named location |
| 172 | `PATCH` | `/identity/conditionalAccess/namedLocations/{id}` | Update a named location |
| 173 | `DELETE` | `/identity/conditionalAccess/namedLocations/{id}` | Delete a named location |
| 174 | `GET` | `/identity/conditionalAccess/templates` | List CA templates |
| 175 | `GET` | `/identity/conditionalAccess/templates/{id}` | Get a CA template |
| 176 | `GET` | `/identity/conditionalAccess/authenticationContext/classReferences` | List auth context class refs |
| 177 | `POST` | `/identity/conditionalAccess/authenticationContext/classReferences` | Create |
| 178 | `GET` | `/identity/continuousAccessEvaluationPolicy` | Get CAE policy |

### 5B. Identity Protection
| # | Method | Path | Description |
|---|--------|------|-------------|
| 179 | `GET` | `/identityProtection/riskDetections` | List risk detections |
| 180 | `GET` | `/identityProtection/riskDetections/{id}` | Get a risk detection |
| 181 | `GET` | `/identityProtection/riskyUsers` | List risky users |
| 182 | `GET` | `/identityProtection/riskyUsers/{id}` | Get a risky user |
| 183 | `POST` | `/identityProtection/riskyUsers/confirmCompromised` | Confirm users compromised |
| 184 | `POST` | `/identityProtection/riskyUsers/dismiss` | Dismiss user risk |
| 185 | `GET` | `/identityProtection/riskyServicePrincipals` | List risky service principals |
| 186 | `GET` | `/identityProtection/riskyServicePrincipals/{id}` | Get a risky SP |
| 187 | `POST` | `/identityProtection/riskyServicePrincipals/confirmCompromised` | Confirm SP compromised |
| 188 | `POST` | `/identityProtection/riskyServicePrincipals/dismiss` | Dismiss SP risk |
| 189 | `GET` | `/identityProtection/servicePrincipalRiskDetections` | List SP risk detections |

---

## Tier 6 — Audit Logs & Sign-in Reports

*Read-only telemetry — essential for compliance and monitoring test scenarios.*

### 6A. Directory Audit & Sign-in Logs
| # | Method | Path | Description |
|---|--------|------|-------------|
| 190 | `GET` | `/auditLogs/directoryAudits` | List directory audit logs |
| 191 | `GET` | `/auditLogs/directoryAudits/{id}` | Get a specific audit log entry |
| 192 | `GET` | `/auditLogs/signIns` | List sign-in logs |
| 193 | `GET` | `/auditLogs/signIns/{id}` | Get a specific sign-in log |
| 194 | `GET` | `/auditLogs/provisioning` | List provisioning logs |
| 195 | `GET` | `/auditLogs/provisioning/{id}` | Get a provisioning log entry |

---

## Tier 7 — B2B / Guest Invitations & Entitlement Management

*Cross-tenant collaboration and access package lifecycle.*

### 7A. Invitations
| # | Method | Path | Description |
|---|--------|------|-------------|
| 196 | `GET` | `/invitations` | List invitations |
| 197 | `POST` | `/invitations` | Create a guest user invitation |

### 7B. Entitlement Management (Identity Governance)
| # | Method | Path | Description |
|---|--------|------|-------------|
| 198 | `GET` | `/identityGovernance/entitlementManagement/accessPackages` | List access packages |
| 199 | `GET` | `/identityGovernance/entitlementManagement/accessPackages/{id}` | Get an access package |
| 200 | `POST` | `/identityGovernance/entitlementManagement/accessPackages` | Create an access package |
| 201 | `PATCH` | `/identityGovernance/entitlementManagement/accessPackages/{id}` | Update an access package |
| 202 | `DELETE` | `/identityGovernance/entitlementManagement/accessPackages/{id}` | Delete an access package |
| 203 | `GET` | `/identityGovernance/entitlementManagement/accessPackageCatalogs` | List catalogs |
| 204 | `POST` | `/identityGovernance/entitlementManagement/accessPackageCatalogs` | Create a catalog |
| 205 | `GET` | `/identityGovernance/entitlementManagement/accessPackageAssignments` | List assignments |
| 206 | `POST` | `/identityGovernance/entitlementManagement/accessPackageAssignmentRequests` | Create assignment request |
| 207 | `GET` | `/identityGovernance/entitlementManagement/accessPackageAssignmentPolicies` | List policies |
| 208 | `POST` | `/identityGovernance/entitlementManagement/accessPackageAssignmentPolicies` | Create a policy |

### 7C. Access Reviews
| # | Method | Path | Description |
|---|--------|------|-------------|
| 209 | `GET` | `/identityGovernance/accessReviews/definitions` | List access review definitions |
| 210 | `GET` | `/identityGovernance/accessReviews/definitions/{id}` | Get a definition |
| 211 | `POST` | `/identityGovernance/accessReviews/definitions` | Create a definition |
| 212 | `GET` | `/identityGovernance/accessReviews/definitions/{id}/instances` | List instances |
| 213 | `GET` | `/identityGovernance/accessReviews/definitions/{id}/instances/{instanceId}` | Get an instance |
| 214 | `POST` | `/identityGovernance/accessReviews/definitions/{id}/instances/{instanceId}/acceptRecommendations` | Accept recommendations |
| 215 | `POST` | `/identityGovernance/accessReviews/definitions/{id}/instances/{instanceId}/batchRecordDecisions` | Batch record decisions |
| 216 | `GET` | `/identityGovernance/accessReviews/definitions/{id}/instances/{instanceId}/decisions` | List decisions |

---

## Tier 8 — Policies, Lifecycle & Miscellaneous

*Rarely-needed for core scenarios but complete the surface.*

### 8A. Policies
| # | Method | Path | Description |
|---|--------|------|-------------|
| 217 | `GET` | `/policies` | List policy roots |
| 218 | `GET` | `/policies/authorizationPolicy` | Get authorization policy |
| 219 | `PATCH` | `/policies/authorizationPolicy` | Update authorization policy |
| 220 | `GET` | `/policies/identitySecurityDefaultsEnforcementPolicy` | Get security defaults |
| 221 | `PATCH` | `/policies/identitySecurityDefaultsEnforcementPolicy` | Update security defaults |
| 222 | `GET` | `/policies/permissionGrantPolicies` | List permission grant policies |
| 223 | `GET` | `/policies/permissionGrantPolicies/{id}` | Get a permission grant policy |
| 224 | `GET` | `/policies/roleManagementPolicies` | List role management policies |
| 225 | `GET` | `/policies/roleManagementPolicies/{id}` | Get a role management policy |
| 226 | `PATCH` | `/policies/roleManagementPolicies/{id}` | Update a role management policy |
| 227 | `GET` | `/policies/activityBasedTimeoutPolicies` | List ABT policies |
| 228 | `GET` | `/policies/tokenLifetimePolicies` | List token lifetime policies |
| 229 | `POST` | `/policies/tokenLifetimePolicies` | Create a token lifetime policy |
| 230 | `GET` | `/policies/tokenIssuancePolicies` | List token issuance policies |
| 231 | `GET` | `/policies/homeRealmDiscoveryPolicies` | List HRD policies |
| 232 | `GET` | `/policies/claimsMappingPolicies` | List claims mapping policies |

### 8B. Subscribed Skus / Licensing
| # | Method | Path | Description |
|---|--------|------|-------------|
| 233 | `GET` | `/subscribedSkus` | List subscribed SKUs (license info) |

### 8C. Directory Setting Templates & Settings
| # | Method | Path | Description |
|---|--------|------|-------------|
| 234 | `GET` | `/directorySettingTemplates` | List directory setting templates |
| 235 | `GET` | `/directorySettingTemplates/{id}` | Get a setting template |
| 236 | `GET` | `/settings` | List tenant-level settings |
| 237 | `POST` | `/settings` | Create a setting |
| 238 | `PATCH` | `/settings/{id}` | Update a setting |
| 239 | `DELETE` | `/settings/{id}` | Delete a setting |

### 8D. Contact Objects
| # | Method | Path | Description |
|---|--------|------|-------------|
| 240 | `GET` | `/contacts` | List org contacts |
| 241 | `GET` | `/contacts/{id}` | Get an org contact |
| 242 | `GET` | `/contacts/{id}/memberOf` | Groups the contact is in |
| 243 | `GET` | `/contacts/{id}/transitiveMemberOf` | Transitive memberOf |
| 244 | `GET` | `/contacts/{id}/directReports` | Direct reports |
| 245 | `GET` | `/contacts/{id}/manager` | Get manager |

### 8E. Batch & OData
| # | Method | Path | Description |
|---|--------|------|-------------|
| 246 | `POST` | `/$batch` | JSON batch (combine multiple requests) |
| 247 | — | — | OData query parameters: `$filter`, `$select`, `$expand`, `$top`, `$orderby`, `$count`, `$search` |

---

## Summary: Priority Rationale

| Tier | What | Why it's at this priority |
|------|------|--------------------------|
| **1** | Auth + Users + Groups | Every Entra integration needs these. Groups drive authorization decisions. Without auth, nothing else matters. |
| **2** | Applications + Service Principals + OAuth2 Grants + App Roles | App registrations are the #2 most-common operation. SPs are how apps authenticate. Permission grants are required for delegated flows. |
| **3** | Directory Roles + RBAC + Admin Units + Directory Objects | Role-based access control is foundational for enterprise scenarios. Admin units enable scoped administration. |
| **4** | Organization + Domains + Devices | Tenant config, domain verification, and device registration — needed for onboarding scenarios but not day-to-day API calls. |
| **5** | Conditional Access + Identity Protection | Security policies — critical for production but can be stubbed in a simulator that focuses on identity data. |
| **6** | Audit Logs + Sign-in Logs | Read-only telemetry. Important for compliance but the simulator can generate synthetic log data. |
| **7** | B2B Invitations + Entitlement Management + Access Reviews | Cross-tenant and governance workflows — important for specific enterprise scenarios but niche. |
| **8** | Policies + Licensing + Settings + Contacts + Batch | Long tail. Completes the surface but rarely the first thing a developer needs in a simulator. |

**Total unique endpoints catalogued: ~248** (including OData support as a cross-cutting concern)
