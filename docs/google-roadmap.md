# Google Workspace API Simulator — Tiered Implementation Plan

> **Base URL:** `https://admin.googleapis.com`  
> **Auth:** OAuth 2.0 (service account with domain-wide delegation, OAuth client, API key)  
> **Scopes:** `https://www.googleapis.com/auth/admin.directory.user`, `.group`, `.device.chromeos`, etc.  
> **Common query params:** `query`, `maxResults`, `pageToken`, `orderBy`, `sortOrder`, `projection`, `customer`, `domain`

---

## Tier 1 — Authentication & Identity Core (Users & Groups)

*The minimum viable simulator. Every Google Workspace integration starts here.*

### 1A. Authentication Layer
| # | Method | Path | Description |
|---|--------|------|-------------|
| 1 | `POST` | `https://oauth2.googleapis.com/token` | OAuth 2.0 token exchange (grant types: authorization_code, refresh_token, client_credentials via JWT assertion) |
| 2 | `POST` | `https://oauth2.googleapis.com/tokeninfo` | Token validation / introspection |
| 3 | — | — | Bearer token middleware (validate OAuth2 access token, check scopes, expiry) |

### 1B. Users
| # | Method | Path | Description |
|---|--------|------|-------------|
| 4 | `GET` | `/admin/directory/v1/users` | List users (supports `query`, `domain`, `customer`, `maxResults`, `orderBy`, `sortOrder`, `projection`, `viewType`, `showDeleted`) |
| 5 | `GET` | `/admin/directory/v1/users/{userKey}` | Get a user by primary email or unique ID |
| 6 | `POST` | `/admin/directory/v1/users` | Create a user |
| 7 | `PUT` | `/admin/directory/v1/users/{userKey}` | Update a user (full replacement) |
| 8 | `PATCH` | `/admin/directory/v1/users/{userKey}` | Update a user (patch semantics) |
| 9 | `DELETE` | `/admin/directory/v1/users/{userKey}` | Delete a user |
| 10 | `POST` | `/admin/directory/v1/users/{userKey}/makeAdmin` | Make a user a super administrator |
| 11 | `POST` | `/admin/directory/v1/users/{userKey}/undelete` | Undelete a deleted user |
| 12 | `POST` | `/admin/directory/v1/users/{userKey}/signOut` | Sign user out of all web and device sessions |
| 13 | `POST` | `/admin/directory/v1/users/{userKey}/aliases` | Add an alias for a user |
| 14 | `GET` | `/admin/directory/v1/users/{userKey}/aliases` | List all aliases for a user |
| 15 | `DELETE` | `/admin/directory/v1/users/{userKey}/aliases/{alias}` | Remove a user alias |
| 16 | `GET` | `/admin/directory/v1/users/{userKey}/photos/thumbnail` | Get user photo |
| 17 | `PUT` | `/admin/directory/v1/users/{userKey}/photos/thumbnail` | Update user photo |
| 18 | `PATCH` | `/admin/directory/v1/users/{userKey}/photos/thumbnail` | Patch user photo |
| 19 | `DELETE` | `/admin/directory/v1/users/{userKey}/photos/thumbnail` | Delete user photo |
| 20 | `POST` | `/admin/directory/v1/users:createGuest` | Create a guest user |
| 21 | `POST` | `/admin/directory/v1/users/watch` | Watch for changes in users list (Push notifications) |

### 1C. Groups
| # | Method | Path | Description |
|---|--------|------|-------------|
| 22 | `GET` | `/admin/directory/v1/groups` | List groups (supports `query`, `domain`, `customer`, `maxResults`, `pageToken`, `userKey`) |
| 23 | `GET` | `/admin/directory/v1/groups/{groupKey}` | Get a group by email or unique ID |
| 24 | `POST` | `/admin/directory/v1/groups` | Create a group |
| 25 | `PUT` | `/admin/directory/v1/groups/{groupKey}` | Update a group |
| 26 | `PATCH` | `/admin/directory/v1/groups/{groupKey}` | Patch a group |
| 27 | `DELETE` | `/admin/directory/v1/groups/{groupKey}` | Delete a group |
| 28 | `POST` | `/admin/directory/v1/groups/{groupKey}/aliases` | Add an alias for a group |
| 29 | `GET` | `/admin/directory/v1/groups/{groupKey}/aliases` | List all aliases for a group |
| 30 | `DELETE` | `/admin/directory/v1/groups/{groupKey}/aliases/{alias}` | Remove a group alias |

### 1D. Members
| # | Method | Path | Description |
|---|--------|------|-------------|
| 31 | `GET` | `/admin/directory/v1/groups/{groupKey}/members` | List members of a group |
| 32 | `GET` | `/admin/directory/v1/groups/{groupKey}/members/{memberKey}` | Get a member |
| 33 | `POST` | `/admin/directory/v1/groups/{groupKey}/members` | Add a member to a group |
| 34 | `PUT` | `/admin/directory/v1/groups/{groupKey}/members/{memberKey}` | Update a member |
| 35 | `PATCH` | `/admin/directory/v1/groups/{groupKey}/members/{memberKey}` | Patch a member |
| 36 | `DELETE` | `/admin/directory/v1/groups/{groupKey}/members/{memberKey}` | Remove a member from a group |
| 37 | `GET` | `/admin/directory/v1/groups/{groupKey}/hasMember/{memberKey}` | Check if user is a member of a group |

---

## Tier 2 — Organization Structure & RBAC

*Essential for modeling org hierarchy and role-based access control.*

### 2A. Organizational Units
| # | Method | Path | Description |
|---|--------|------|-------------|
| 38 | `GET` | `/admin/directory/v1/customer/{customerId}/orgunits` | List organizational units |
| 39 | `GET` | `/admin/directory/v1/customer/{customerId}/orgunits/{orgUnitPath}` | Get an organizational unit |
| 40 | `POST` | `/admin/directory/v1/customer/{customerId}/orgunits` | Create an organizational unit |
| 41 | `PUT` | `/admin/directory/v1/customer/{customerId}/orgunits/{orgUnitPath}` | Update an organizational unit |
| 42 | `PATCH` | `/admin/directory/v1/customer/{customerId}/orgunits/{orgUnitPath}` | Patch an organizational unit |
| 43 | `DELETE` | `/admin/directory/v1/customer/{customerId}/orgunits/{orgUnitPath}` | Delete an organizational unit |

### 2B. Roles
| # | Method | Path | Description |
|---|--------|------|-------------|
| 44 | `GET` | `/admin/directory/v1/customer/{customer}/roles` | List roles |
| 45 | `GET` | `/admin/directory/v1/customer/{customer}/roles/{roleId}` | Get a role |
| 46 | `POST` | `/admin/directory/v1/customer/{customer}/roles` | Create a role |
| 47 | `PUT` | `/admin/directory/v1/customer/{customer}/roles/{roleId}` | Update a role |
| 48 | `PATCH` | `/admin/directory/v1/customer/{customer}/roles/{roleId}` | Patch a role |
| 49 | `DELETE` | `/admin/directory/v1/customer/{customer}/roles/{roleId}` | Delete a role |

### 2C. Role Assignments
| # | Method | Path | Description |
|---|--------|------|-------------|
| 50 | `GET` | `/admin/directory/v1/customer/{customer}/roleassignments` | List role assignments |
| 51 | `GET` | `/admin/directory/v1/customer/{customer}/roleassignments/{roleAssignmentId}` | Get a role assignment |
| 52 | `POST` | `/admin/directory/v1/customer/{customer}/roleassignments` | Create a role assignment |
| 53 | `DELETE` | `/admin/directory/v1/customer/{customer}/roleassignments/{roleAssignmentId}` | Delete a role assignment |

### 2D. Privileges
| # | Method | Path | Description |
|---|--------|------|-------------|
| 54 | `GET` | `/admin/directory/v1/customer/{customer}/roles/ALL/privileges` | List all privileges for a customer |

---

## Tier 3 — Customer, Domains & Domain Aliases

*Tenant-level configuration and domain management.*

### 3A. Customers
| # | Method | Path | Description |
|---|--------|------|-------------|
| 55 | `GET` | `/admin/directory/v1/customers/{customerKey}` | Get customer (tenant) details |
| 56 | `PATCH` | `/admin/directory/v1/customers/{customerKey}` | Patch a customer |
| 57 | `PUT` | `/admin/directory/v1/customers/{customerKey}` | Update a customer |

### 3B. Domains
| # | Method | Path | Description |
|---|--------|------|-------------|
| 58 | `GET` | `/admin/directory/v1/customer/{customer}/domains` | List domains |
| 59 | `GET` | `/admin/directory/v1/customer/{customer}/domains/{domainName}` | Get a domain |
| 60 | `POST` | `/admin/directory/v1/customer/{customer}/domains` | Add a domain |
| 61 | `DELETE` | `/admin/directory/v1/customer/{customer}/domains/{domainName}` | Delete a domain |

### 3C. Domain Aliases
| # | Method | Path | Description |
|---|--------|------|-------------|
| 62 | `GET` | `/admin/directory/v1/customer/{customer}/domainaliases` | List domain aliases |
| 63 | `GET` | `/admin/directory/v1/customer/{customer}/domainaliases/{domainAliasName}` | Get a domain alias |
| 64 | `POST` | `/admin/directory/v1/customer/{customer}/domainaliases` | Create a domain alias |
| 65 | `DELETE` | `/admin/directory/v1/customer/{customer}/domainaliases/{domainAliasName}` | Delete a domain alias |

---

## Tier 4 — Devices (Chrome OS & Mobile)

*Device management — critical for enterprise device fleet management.*

### 4A. Chrome OS Devices
| # | Method | Path | Description |
|---|--------|------|-------------|
| 66 | `GET` | `/admin/directory/v1/customer/{customerId}/devices/chromeos` | List Chrome OS devices |
| 67 | `GET` | `/admin/directory/v1/customer/{customerId}/devices/chromeos/{deviceId}` | Get a Chrome OS device |
| 68 | `PATCH` | `/admin/directory/v1/customer/{customerId}/devices/chromeos/{deviceId}` | Patch a Chrome OS device |
| 69 | `PUT` | `/admin/directory/v1/customer/{customerId}/devices/chromeos/{deviceId}` | Update a Chrome OS device |
| 70 | `POST` | `/admin/directory/v1/customer/{customerId}/devices/chromeos/moveDevicesToOu` | Move Chrome OS devices to an OU |
| 71 | `POST` | `/admin/directory/v1/customer/{customerId}/devices/chromeos:batchChangeStatus` | Batch change Chrome OS device status |
| 72 | `GET` | `/admin/directory/v1/customer/{customerId}/devices/chromeos:countChromeOsDevices` | Count Chrome OS devices |
| 73 | `POST` | `/admin/directory/v1/customer/{customerId}/devices/chromeos/{deviceId}:issueCommand` | Issue a command to a Chrome OS device |
| 74 | `GET` | `/admin/directory/v1/customer/{customerId}/devices/chromeos/{deviceId}/commands/{commandId}` | Get command data for a Chrome OS device |

### 4B. Mobile Devices
| # | Method | Path | Description |
|---|--------|------|-------------|
| 75 | `GET` | `/admin/directory/v1/customer/{customerId}/devices/mobile` | List mobile devices |
| 76 | `GET` | `/admin/directory/v1/customer/{customerId}/devices/mobile/{resourceId}` | Get a mobile device |
| 77 | `DELETE` | `/admin/directory/v1/customer/{customerId}/devices/mobile/{resourceId}` | Delete a mobile device |
| 78 | `POST` | `/admin/directory/v1/customer/{customerId}/devices/mobile/{resourceId}/action` | Take action on a mobile device (wipe, approve, block, etc.) |

### 4C. Cloud Identity Devices
*(Service endpoint: `https://cloudidentity.googleapis.com`)*
| # | Method | Path | Description |
|---|--------|------|-------------|
| 79 | `GET` | `/v1/devices` | List/search Cloud Identity devices |
| 80 | `GET` | `/v1/{name=devices/*}` | Get a Cloud Identity device |
| 81 | `POST` | `/v1/devices` | Create a company-owned device |
| 82 | `DELETE` | `/v1/{name=devices/*}` | Delete a Cloud Identity device |
| 83 | `POST` | `/v1/{name=devices/*}:cancelWipe` | Cancel an unfinished device wipe |
| 84 | `POST` | `/v1/{name=devices/*}:wipe` | Wipe all data on a device |
| 85 | `GET` | `/v1/{parent=devices/*}/deviceUsers` | List device users |
| 86 | `GET` | `/v1/{name=devices/*/deviceUsers/*}` | Get a device user |
| 87 | `DELETE` | `/v1/{name=devices/*/deviceUsers/*}` | Delete a device user |
| 88 | `POST` | `/v1/{name=devices/*/deviceUsers/*}:approve` | Approve a device for user data access |
| 89 | `POST` | `/v1/{name=devices/*/deviceUsers/*}:block` | Block a device from accessing user data |
| 90 | `POST` | `/v1/{name=devices/*/deviceUsers/*}:wipe` | Wipe user's account on device |
| 91 | `POST` | `/v1/{name=devices/*/deviceUsers/*}:cancelWipe` | Cancel user account wipe |
| 92 | `GET` | `/v1/{parent=devices/*/deviceUsers}:lookup` | Lookup device users by caller's credentials |

---

## Tier 5 — Cloud Identity Groups & Memberships

*Advanced group management with transitive membership, search, and security settings.*

*(Service endpoint: `https://cloudidentity.googleapis.com`)*

### 5A. Groups
| # | Method | Path | Description |
|---|--------|------|-------------|
| 93 | `GET` | `/v1/groups` | List Cloud Identity groups |
| 94 | `GET` | `/v1/{name=groups/*}` | Get a Cloud Identity group |
| 95 | `POST` | `/v1/groups` | Create a Cloud Identity group |
| 96 | `PATCH` | `/v1/{resource.name=groups/*}` | Update a Cloud Identity group |
| 97 | `DELETE` | `/v1/{name=groups/*}` | Delete a Cloud Identity group |
| 98 | `GET` | `/v1/groups:lookup` | Lookup group by entity key |
| 99 | `GET` | `/v1/groups:search` | Search groups matching a query |
| 100 | `GET` | `/v1/{name=groups/*/securitySettings}` | Get group security settings |
| 101 | `PATCH` | `/v1/{securitySettings.name=groups/*/securitySettings}` | Update group security settings |

### 5B. Memberships
| # | Method | Path | Description |
|---|--------|------|-------------|
| 102 | `GET` | `/v1/{parent=groups/*}/memberships` | List memberships in a group |
| 103 | `GET` | `/v1/{name=groups/*/memberships/*}` | Get a membership |
| 104 | `POST` | `/v1/{parent=groups/*}/memberships` | Create a membership |
| 105 | `DELETE` | `/v1/{name=groups/*/memberships/*}` | Delete a membership |
| 106 | `GET` | `/v1/{parent=groups/*}/memberships:lookup` | Lookup membership by entity key |
| 107 | `POST` | `/v1/{name=groups/*/memberships/*}:modifyMembershipRoles` | Modify membership roles |
| 108 | `GET` | `/v1/{parent=groups/*}/memberships:checkTransitiveMembership` | Check transitive membership |
| 109 | `GET` | `/v1/{parent=groups/*}/memberships:getMembershipGraph` | Get membership graph |
| 110 | `GET` | `/v1/{parent=groups/*}/memberships:searchTransitiveGroups` | Search transitive groups of a member |
| 111 | `GET` | `/v1/{parent=groups/*}/memberships:searchTransitiveMemberships` | Search transitive memberships of a group |
| 112 | `GET` | `/v1/{parent=groups/*}/memberships:searchDirectGroups` | Search direct groups of a member |

---

## Tier 6 — Reports, Audit Logs & Usage

*Read-only telemetry — essential for compliance and monitoring test scenarios.*

### 6A. Activity Reports
| # | Method | Path | Description |
|---|--------|------|-------------|
| 113 | `GET` | `/admin/reports/v1/activity/users/{userKey}/applications/{applicationName}` | List activities for a specific application (admin, drive, login, token, etc.) |
| 114 | `POST` | `/admin/reports/v1/activity/users/{userKey}/applications/{applicationName}/watch` | Watch for activity changes (Push notifications) |

### 6B. Usage Reports
| # | Method | Path | Description |
|---|--------|------|-------------|
| 115 | `GET` | `/admin/reports/v1/usage/dates/{date}` | Get customer usage report |
| 116 | `GET` | `/admin/reports/v1/usage/users/{userKey}/dates/{date}` | Get user usage report |
| 117 | `GET` | `/admin/reports/v1/usage/{entityType}/{entityKey}/dates/{date}` | Get entity usage report |

---

## Tier 7 — Security, Tokens & User Invitations

*Security-critical operations: token management, 2SV, verification codes, and user invitations.*

### 7A. Tokens & ASPs
| # | Method | Path | Description |
|---|--------|------|-------------|
| 118 | `GET` | `/admin/directory/v1/users/{userKey}/tokens` | List tokens issued by a user to 3rd party apps |
| 119 | `GET` | `/admin/directory/v1/users/{userKey}/tokens/{clientId}` | Get a token |
| 120 | `DELETE` | `/admin/directory/v1/users/{userKey}/tokens/{clientId}` | Delete a token |
| 121 | `GET` | `/admin/directory/v1/users/{userKey}/asps` | List application-specific passwords (ASPs) |
| 122 | `GET` | `/admin/directory/v1/users/{userKey}/asps/{codeId}` | Get an ASP |
| 123 | `DELETE` | `/admin/directory/v1/users/{userKey}/asps/{codeId}` | Delete an ASP |

### 7B. Verification Codes & Two-Step Verification
| # | Method | Path | Description |
|---|--------|------|-------------|
| 124 | `GET` | `/admin/directory/v1/users/{userKey}/verificationCodes` | List backup verification codes |
| 125 | `POST` | `/admin/directory/v1/users/{userKey}/verificationCodes/generate` | Generate new backup verification codes |
| 126 | `POST` | `/admin/directory/v1/users/{userKey}/verificationCodes/invalidate` | Invalidate backup verification codes |
| 127 | `POST` | `/admin/directory/v1/users/{userKey}/twoStepVerification/turnOff` | Turn off 2-Step Verification for a user |

### 7C. User Invitations (Cloud Identity)
*(Service endpoint: `https://cloudidentity.googleapis.com`)*
| # | Method | Path | Description |
|---|--------|------|-------------|
| 128 | `GET` | `/v1/{parent=customers/*}/userinvitations` | List user invitations |
| 129 | `GET` | `/v1/{name=customers/*/userinvitations/*}` | Get a user invitation |
| 130 | `GET` | `/v1/{name=customers/*/userinvitations/*}:isInvitableUser` | Check if user is eligible for invitation |
| 131 | `POST` | `/v1/{name=customers/*/userinvitations/*}:send` | Send a user invitation |
| 132 | `POST` | `/v1/{name=customers/*/userinvitations/*}:cancel` | Cancel a user invitation |

---

## Tier 8 — Resources, Schemas, Groups Settings & Extended APIs

*Completes the surface with resource management, custom schemas, group settings, data transfer, and event subscriptions.*

### 8A. Custom Schemas
| # | Method | Path | Description |
|---|--------|------|-------------|
| 133 | `GET` | `/admin/directory/v1/customer/{customerId}/schemas` | List all custom schemas |
| 134 | `GET` | `/admin/directory/v1/customer/{customerId}/schemas/{schemaKey}` | Get a custom schema |
| 135 | `POST` | `/admin/directory/v1/customer/{customerId}/schemas` | Create a custom schema |
| 136 | `PUT` | `/admin/directory/v1/customer/{customerId}/schemas/{schemaKey}` | Update a custom schema |
| 137 | `PATCH` | `/admin/directory/v1/customer/{customerId}/schemas/{schemaKey}` | Patch a custom schema |
| 138 | `DELETE` | `/admin/directory/v1/customer/{customerId}/schemas/{schemaKey}` | Delete a custom schema |

### 8B. Calendar Resources
| # | Method | Path | Description |
|---|--------|------|-------------|
| 139 | `GET` | `/admin/directory/v1/customer/{customer}/resources/calendars` | List calendar resources |
| 140 | `GET` | `/admin/directory/v1/customer/{customer}/resources/calendars/{calendarResourceId}` | Get a calendar resource |
| 141 | `POST` | `/admin/directory/v1/customer/{customer}/resources/calendars` | Insert a calendar resource |
| 142 | `PUT` | `/admin/directory/v1/customer/{customer}/resources/calendars/{calendarResourceId}` | Update a calendar resource |
| 143 | `PATCH` | `/admin/directory/v1/customer/{customer}/resources/calendars/{calendarResourceId}` | Patch a calendar resource |
| 144 | `DELETE` | `/admin/directory/v1/customer/{customer}/resources/calendars/{calendarResourceId}` | Delete a calendar resource |

### 8C. Buildings
| # | Method | Path | Description |
|---|--------|------|-------------|
| 145 | `GET` | `/admin/directory/v1/customer/{customer}/resources/buildings` | List buildings |
| 146 | `GET` | `/admin/directory/v1/customer/{customer}/resources/buildings/{buildingId}` | Get a building |
| 147 | `POST` | `/admin/directory/v1/customer/{customer}/resources/buildings` | Insert a building |
| 148 | `PUT` | `/admin/directory/v1/customer/{customer}/resources/buildings/{buildingId}` | Update a building |
| 149 | `PATCH` | `/admin/directory/v1/customer/{customer}/resources/buildings/{buildingId}` | Patch a building |
| 150 | `DELETE` | `/admin/directory/v1/customer/{customer}/resources/buildings/{buildingId}` | Delete a building |

### 8D. Features
| # | Method | Path | Description |
|---|--------|------|-------------|
| 151 | `GET` | `/admin/directory/v1/customer/{customer}/resources/features` | List features |
| 152 | `GET` | `/admin/directory/v1/customer/{customer}/resources/features/{featureKey}` | Get a feature |
| 153 | `POST` | `/admin/directory/v1/customer/{customer}/resources/features` | Insert a feature |
| 154 | `PUT` | `/admin/directory/v1/customer/{customer}/resources/features/{featureKey}` | Update a feature |
| 155 | `PATCH` | `/admin/directory/v1/customer/{customer}/resources/features/{featureKey}` | Patch a feature |
| 156 | `DELETE` | `/admin/directory/v1/customer/{customer}/resources/features/{featureKey}` | Delete a feature |
| 157 | `POST` | `/admin/directory/v1/customer/{customer}/resources/features/{oldName}/rename` | Rename a feature |

### 8E. Groups Settings
*(Service endpoint: `https://www.googleapis.com/groups/v1/groups`)*
| # | Method | Path | Description |
|---|--------|------|-------------|
| 158 | `GET` | `/{groupUniqueId}` | Get a group's settings (access, notifications, moderation, etc.) |
| 159 | `PUT` | `/{groupUniqueId}` | Update a group's settings |
| 160 | `PATCH` | `/{groupUniqueId}` | Patch a group's settings |

### 8F. Data Transfer
| # | Method | Path | Description |
|---|--------|------|-------------|
| 161 | `GET` | `/admin/datatransfer/v1/applications` | List applications available for data transfer |
| 162 | `GET` | `/admin/datatransfer/v1/applications/{applicationId}` | Get application info for data transfer |
| 163 | `GET` | `/admin/datatransfer/v1/transfers` | List data transfers |
| 164 | `GET` | `/admin/datatransfer/v1/transfers/{dataTransferId}` | Get a data transfer request |
| 165 | `POST` | `/admin/datatransfer/v1/transfers` | Insert a data transfer request |

### 8G. Workspace Events (Subscriptions)
*(Service endpoint: `https://workspaceevents.googleapis.com`)*
| # | Method | Path | Description |
|---|--------|------|-------------|
| 166 | `POST` | `/v1/subscriptions` | Create a subscription |
| 167 | `GET` | `/v1/subscriptions` | List subscriptions |
| 168 | `GET` | `/v1/{name=subscriptions/*}` | Get a subscription |
| 169 | `PATCH` | `/v1/{subscription.name=subscriptions/*}` | Update or renew a subscription |
| 170 | `DELETE` | `/v1/{name=subscriptions/*}` | Delete a subscription |
| 171 | `POST` | `/v1/{name=subscriptions/*}:reactivate` | Reactivate a suspended subscription |

---

## Summary: Priority Rationale

| Tier | What | Why it's at this priority |
|------|------|--------------------------|
| **1** | Auth + Users + Groups + Members | Every Google Workspace integration needs these. Groups drive authorization decisions. Without auth, nothing else matters. |
| **2** | Org Units + Roles + Role Assignments + Privileges | Org structure and RBAC — foundational for enterprise scenarios. Admin roles govern what users can manage. |
| **3** | Customer + Domains + Domain Aliases | Tenant configuration and domain management — needed for onboarding and multi-domain scenarios. |
| **4** | Chrome OS + Mobile + Cloud Identity Devices | Device fleet management — critical for enterprise device management and security enforcement. |
| **5** | Cloud Identity Groups & Memberships | Advanced group features (transitive membership, search, security settings) — complements basic Directory groups. |
| **6** | Reports + Activity + Usage Reports | Read-only telemetry. Important for compliance but can generate synthetic data in a simulator. |
| **7** | Security + Tokens + ASPs + Verification Codes + User Invitations | Security operations — important for specific scenarios (token auditing, 2SV management, unmanaged account handling). |
| **8** | Schemas + Resources + Groups Settings + Data Transfer + Subscriptions | Long tail. Completes the surface but rarely the first thing a developer needs in a simulator. |

**Total unique endpoints catalogued: ~174** (including pagination, push notifications, and multi-service endpoints)
