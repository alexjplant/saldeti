# saldeti

A local Microsoft Graph API simulator for development and testing. saldeti mimics the Microsoft Entra ID (Azure AD) Graph API v1.0 endpoints, allowing you to develop and test applications that integrate with Microsoft Graph without needing a real Azure tenant.

## Features

- **Management Web UI**: Available at `/ui` after starting the server
- **Authentication**: OAuth 2.0 token endpoint (client_credentials, authorization_code, refresh_token)
- **Users**: Full CRUD with OData query support ($filter, $select, $top, $orderby, $count, $search)
- **Groups**: Full CRUD with member/owner management and transitive membership resolution
- **Applications**: Full CRUD with password/key credential management, owner management, extension properties, and verified publisher support
- **Service Principals**: Full CRUD with password/key credential management, owner management, app role assignments, and OAuth2 permission grant listing
- **App Role Assignments**: Assign and manage app roles on users, groups, and service principals
- **OAuth2 Permission Grants**: Full CRUD for delegated permission grants
- **Navigation Properties**: memberOf, transitiveMemberOf, manager, directReports
- **Directory Objects**: getByIds batch lookup, checkMemberGroups, getMemberGroups
- **Delta Queries**: Delta query support for users, groups, and applications
- **$select & $orderby**: Field projection and sorting support across all list endpoints

A full roadmap is available in [`docs/roadmap.md`](docs/roadmap.md).

## Quick Start

```bash
# Build
mise run build

# Run with empty store (admin client credentials are logged at startup)
./bin/saldeti -port 9443

# Run with sample seed data (persists changes on shutdown)
./bin/saldeti -port 9443 -seed examples/seed.json -dump snapshot.json

# Get a token using the admin client credentials logged at startup
# (With -seed examples/seed.json, the built-in client is sim-client-id / sim-client-secret)
curl -X POST http://localhost:9443/<tenant-id>/oauth2/v2.0/token \
  -d "grant_type=client_credentials" \
  -d "client_id=<admin-client-id>" \
  -d "client_secret=<admin-client-secret>" \
  -d "scope=User.Read.All Group.Read.All"

# List users
curl http://localhost:9443/v1.0/users \
  -H "Authorization: Bearer <token>"

# Create a user
curl -X POST http://localhost:9443/v1.0/users \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"displayName":"Test User","userPrincipalName":"test@saldeti.local"}'
```

## Seed Data

The simulator can start in two modes:

**Without `-seed` flag** (default): the server starts with a **completely empty store** — no users, no groups, no applications, no sample data of any kind. The UI shows empty lists and you create everything yourself via the API or UI.

The only thing automatically created is an **admin OAuth client** (along with its service principal) that the management UI uses to authenticate. Its credentials (client ID, client secret, and tenant ID) are logged at startup. By default these are random UUIDs; you can override them with the `-admin-client-id`, `-admin-client-secret`, and `-admin-tenant-id` flags.

**With `-seed file.json`**: all data from the specified JSON file is loaded — clients, users, groups, applications, memberships, managers, owners, app role assignments, and OAuth2 grants. The sample `examples/seed.json` includes:

- **Client**: `sim-client-id` / `sim-client-secret`, tenant `sim-tenant-id`
- **Admin user**: `admin@saldeti.local` with password `Simulator123!`
- **10 sample users** with inline manager relationships (via `manager_upn` fields)
- **5 sample groups** with inline memberships (via `member_upns`) and nested groups (via `member_group_names`)
- **Application**: "Saldeti Simulator App" with an AppRole
- **App Role Assignment**: One assignment
- **OAuth2 Permission Grant**: One grant

### Seed File Format

The seed JSON file uses this schema:

```json
{
  "clients": [{ "client_id": "...", "client_secret": "...", "tenant_id": "..." }],
  "users": [{ "email": "...", "display_name": "...", "password": "...", "department": "...", "job_title": "..." }],
  "groups": [{ "display_name": "...", "description": "...", "mail_nickname": "..." }],
  "memberships": [{ "user_index": 0, "group_index": 0 }],
  "managers": [{ "user_index": 1, "manager_index": 0 }],
  "ownerships": [{ "user_index": 0, "group_index": 0 }],
  "applications": [{ "display_name": "...", "app_id": "...", "description": "...", "sign_in_audience": "...", "identifier_uris": [], "app_roles": [], "owner_upns": [] }],
  "service_principals": [{ "app_id": "..." }],
  "app_role_assignments": [{ "principal_index": 0, "resource_app_id": "...", "role_value": "..." }],
  "oauth2_grants": [{ "client_app_id": "...", "resource_app_id": "...", "scope": "...", "consent_type": "...", "principal_upn": "..." }]
}
```

See `examples/seed.json` for a complete example.

### Persisting Changes

Use `-dump` to save the current store state on shutdown (Ctrl+C). Combined with `-seed`, this enables a round-trip workflow:

```bash
# Start with seed data and enable dump-on-shutdown
./bin/saldeti -seed examples/seed.json -dump snapshot.json

# Make changes via the API, then Ctrl+C to shut down
# snapshot.json now contains the updated state

# Restart with the updated state
./bin/saldeti -seed snapshot.json -dump snapshot.json
```

The dump uses the same JSON schema as the seed file, so the output can be fed directly back into `-seed`.

## API Coverage

### Authentication & Discovery

| Endpoint | Methods | Description |
|----------|---------|-------------|
| `/{tenant}/oauth2/v2.0/token` | POST | OAuth2 token exchange |
| `/{tenant}/v2.0/.well-known/openid-configuration` | GET | OpenID discovery document |

### Batch

| Endpoint | Methods | Description |
|----------|---------|-------------|
| `/v1.0/$batch` | POST | JSON batch requests |

### Users

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

### Groups

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

### Applications

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

### Service Principals

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

### OAuth2 Permission Grants

| Endpoint | Methods | Description |
|----------|---------|-------------|
| `/v1.0/oauth2PermissionGrants` | GET, POST | List/create grants |
| `/v1.0/oauth2PermissionGrants/{id}` | GET, PATCH, DELETE | Get/update/delete grant |

### Licensing

| Endpoint | Methods | Description |
|----------|---------|-------------|
| `/v1.0/subscribedSkus` | GET | List available license SKUs |

### Directory Objects

| Endpoint | Methods | Description |
|----------|---------|-------------|
| `/v1.0/directoryObjects/getByIds` | POST | Batch object lookup |

## Development

```bash
mise run build        # Build binary
mise run test         # Run Go unit tests
mise run test-all     # Run all Go tests including E2E
mise run ui-test      # Run UI unit tests
mise run ui-e2e       # Run Playwright E2E tests (builds + starts server + runs tests + stops server)
mise run lint         # Run go vet
mise run clean        # Clean build artifacts
```

## Questions
> What is this for?
Testing your apps and scripts against a make-believe directory so that you ~~don't overwrite everybody's phone number with your own then task a 15-year-old helpdesk intern with manually fixing it while you run down the hall to save your job by finding somebody with a recent backup of the domain controller.~~ can perform integration testing in CI without spinning up a tenant.

> Is this vibe-coded?
Very yes. I used GLM-5.1 for orchestration and planning, GLM-4.7 for implementation, Gemini 3 Flash for UI iteration, and DeepSeek for review.

> ...but why?
Because I wanted it and would rather spend time learning about LLMs and coding harnesses than manually replicating a Microsoft Azure product.

> How is this licensed?

[GNU Affero General Public License v3](https://www.gnu.org/licenses/agpl-3.0.en.html). Use it, improve it, don't make money on it.
