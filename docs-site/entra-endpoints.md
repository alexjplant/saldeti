---
title: Entra ID Endpoint Coverage
description: Complete reference of all implemented Microsoft Graph v1.0 API endpoints in saldeti.
---

# Entra ID Endpoint Coverage

This page documents every Microsoft Graph v1.0 API endpoint currently implemented by saldeti. These endpoints accept the same requests and return the same shaped responses as the real Microsoft Graph, making them suitable for local development and integration testing.

## Authentication & Discovery

| Endpoint | Methods | Description |
|----------|---------|-------------|
| `/{tenant}/oauth2/v2.0/token` | POST | OAuth2 token exchange |
| `/{tenant}/v2.0/.well-known/openid-configuration` | GET | OpenID discovery document |

## Batch

| Endpoint | Methods | Description |
|----------|---------|-------------|
| `/v1.0/$batch` | POST | JSON batch requests |

## Users

| Endpoint | Methods | Description |
|----------|---------|-------------|
| `/v1.0/users` | GET, POST | List/create users |
| `/v1.0/users/{id}` | GET, PATCH, DELETE | Get/update/delete user |
| `/v1.0/users/{id}/memberOf` | GET | Groups user belongs to |
| `/v1.0/users/{id}/transitiveMemberOf` | GET | Transitive group membership |
| `/v1.0/users/{id}/manager` | GET | Get user's manager |
| `/v1.0/users/{id}/manager/$ref` | PUT, DELETE | Set/remove manager |
| `/v1.0/users/{id}/directReports` | GET | List direct reports |
| `/v1.0/users/{id}/checkMemberGroups` | POST | Check group membership |
| `/v1.0/users/{id}/getMemberGroups` | POST | Get all group memberships |
| `/v1.0/users/{id}/appRoleAssignments` | GET, POST | List/create user role assignments |
| `/v1.0/users/{id}/appRoleAssignments/{aid}` | DELETE | Delete user role assignment |
| `/v1.0/users/{id}/photo` | GET | Photo metadata (stub) |
| `/v1.0/users/{id}/photo/$value` | GET, PATCH | Photo binary (stub) |
| `/v1.0/users/{id}/changePassword` | POST | Change password (stub) |
| `/v1.0/users/{id}/reprocessLicenseAssignment` | POST | Reprocess licenses (stub) |
| `/v1.0/users/{id}/licenseDetails` | GET | License details (stub) |
| `/v1.0/users/{id}/assignLicense` | POST | Assign/remove licenses |
| `/v1.0/users/delta` | GET | Delta query for users |
| `/v1.0/me` | GET | Get authenticated user |

## Groups

| Endpoint | Methods | Description |
|----------|---------|-------------|
| `/v1.0/groups` | GET, POST | List/create groups |
| `/v1.0/groups/{id}` | GET, PATCH, DELETE | Get/update/delete group |
| `/v1.0/groups/{id}/members` | GET | List direct members |
| `/v1.0/groups/{id}/members/$ref` | POST | Add member |
| `/v1.0/groups/{id}/members/{mid}/$ref` | DELETE | Remove member |
| `/v1.0/groups/{id}/transitiveMembers` | GET | Transitive members |
| `/v1.0/groups/{id}/owners` | GET | List owners |
| `/v1.0/groups/{id}/owners/$ref` | POST | Add owner |
| `/v1.0/groups/{id}/owners/{oid}/$ref` | DELETE | Remove owner |
| `/v1.0/groups/{id}/memberOf` | GET | Groups this group belongs to |
| `/v1.0/groups/{id}/transitiveMemberOf` | GET | Transitive memberOf |
| `/v1.0/groups/{id}/checkMemberGroups` | POST | Check membership |
| `/v1.0/groups/{id}/getMemberGroups` | POST | Get all memberships |
| `/v1.0/groups/{id}/getMemberObjects` | POST | Get all member objects |
| `/v1.0/groups/{id}/appRoleAssignments` | GET, POST | List/create group role assignments |
| `/v1.0/groups/{id}/appRoleAssignments/{aid}` | DELETE | Delete group role assignment |
| `/v1.0/groups/delta` | GET | Delta query for groups |

## Applications

| Endpoint | Methods | Description |
|----------|---------|-------------|
| `/v1.0/applications` | GET, POST | List/create applications |
| `/v1.0/applications/{id}` | GET, PATCH, DELETE | Get/update/delete application |
| `/v1.0/applications(appId='{appId}')` | GET | Get by alternate key (appId) |
| `/v1.0/applications/{id}/addPassword` | POST | Add password credential |
| `/v1.0/applications/{id}/removePassword` | POST | Remove password credential |
| `/v1.0/applications/{id}/addKey` | POST | Add key credential |
| `/v1.0/applications/{id}/removeKey` | POST | Remove key credential |
| `/v1.0/applications/{id}/owners` | GET | List owners |
| `/v1.0/applications/{id}/owners/$ref` | POST | Add owner |
| `/v1.0/applications/{id}/owners/{oid}/$ref` | DELETE | Remove owner |
| `/v1.0/applications/{id}/extensionProperties` | GET, POST | List/create extension properties |
| `/v1.0/applications/{id}/extensionProperties/{extId}` | DELETE | Delete extension property |
| `/v1.0/applications/{id}/setVerifiedPublisher` | POST | Set verified publisher |
| `/v1.0/applications/delta` | GET | Delta query for applications |

## Service Principals

| Endpoint | Methods | Description |
|----------|---------|-------------|
| `/v1.0/servicePrincipals` | GET, POST | List/create service principals |
| `/v1.0/servicePrincipals/{id}` | GET, PATCH, DELETE | Get/update/delete SP |
| `/v1.0/servicePrincipals(appId='{appId}')` | GET | Get by alternate key (appId) |
| `/v1.0/servicePrincipals/{id}/owners` | GET | List owners |
| `/v1.0/servicePrincipals/{id}/owners/$ref` | POST | Add owner |
| `/v1.0/servicePrincipals/{id}/owners/{oid}/$ref` | DELETE | Remove owner |
| `/v1.0/servicePrincipals/{id}/memberOf` | GET | Groups SP belongs to |
| `/v1.0/servicePrincipals/{id}/transitiveMemberOf` | GET | Transitive group membership |
| `/v1.0/servicePrincipals/{id}/appRoleAssignments` | GET, POST | List/create role assignments |
| `/v1.0/servicePrincipals/{id}/appRoleAssignments/{aid}` | DELETE | Delete role assignment |
| `/v1.0/servicePrincipals/{id}/appRoleAssignedTo` | GET, POST | List/create assigned-to |
| `/v1.0/servicePrincipals/{id}/appRoleAssignedTo/{aid}` | DELETE | Delete assigned-to |
| `/v1.0/servicePrincipals/{id}/oauth2PermissionGrants` | GET | List delegated grants |
| `/v1.0/servicePrincipals/{id}/addPassword` | POST | Add password credential |
| `/v1.0/servicePrincipals/{id}/removePassword` | POST | Remove password credential |
| `/v1.0/servicePrincipals/{id}/addKey` | POST | Add key credential |
| `/v1.0/servicePrincipals/{id}/removeKey` | POST | Remove key credential |
| `/v1.0/servicePrincipals/{id}/homeRealmDiscoveryPolicies` | GET | Policy stub (empty) |
| `/v1.0/servicePrincipals/{id}/claimsMappingPolicies` | GET | Policy stub (empty) |
| `/v1.0/servicePrincipals/{id}/tokenIssuancePolicies` | GET | Policy stub (empty) |
| `/v1.0/servicePrincipals/{id}/tokenLifetimePolicies` | GET | Policy stub (empty) |

## OAuth2 Permission Grants

| Endpoint | Methods | Description |
|----------|---------|-------------|
| `/v1.0/oauth2PermissionGrants` | GET, POST | List/create grants |
| `/v1.0/oauth2PermissionGrants/{id}` | GET, PATCH, DELETE | Get/update/delete grant |

## Licensing

| Endpoint | Methods | Description |
|----------|---------|-------------|
| `/v1.0/subscribedSkus` | GET | List available license SKUs |

## Directory Objects

| Endpoint | Methods | Description |
|----------|---------|-------------|
| `/v1.0/directoryObjects/getByIds` | POST | Batch object lookup |


---

::: tip Disclaimer
Saldeti is an independent project, not affiliated with or endorsed by Microsoft Corporation or Google LLC. All trademarks are property of their respective owners.
:::