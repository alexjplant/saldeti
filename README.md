# saldeti

A local Microsoft Graph API simulator for development and testing. saldeti mimics the Microsoft Entra ID (Azure AD) Graph API v1.0 endpoints, allowing you to develop and test applications that integrate with Microsoft Graph without needing a real Azure tenant.

## Features

- **Authentication**: OAuth 2.0 token endpoint (client_credentials, authorization_code, refresh_token)
- **Users**: Full CRUD with OData query support ($filter, $select, $top, $orderby, $count, $search)
- **Groups**: Full CRUD with member/owner management and transitive membership resolution
- **Navigation Properties**: memberOf, transitiveMemberOf, manager, directReports
- **Directory Objects**: getByIds batch lookup, checkMemberGroups, getMemberGroups
- **Delta Queries**: Basic delta query support for users and groups

## Quick Start

```bash
# Build
mise run build

# Run (empty store)
./bin/saldeti -port 9443

# Run with sample seed data (persists changes on shutdown)
./bin/saldeti -port 9443 -seed examples/seed.json -dump snapshot.json

# Get a token (requires seed data or manually created client)
curl -X POST http://localhost:9443/sim-tenant-id/oauth2/v2.0/token \
  -d "grant_type=client_credentials" \
  -d "client_id=sim-client-id" \
  -d "client_secret=sim-client-secret" \
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

By default, the simulator starts with an **empty store** — no users, groups, or clients. This allows you to create exactly the data you need via the API.

To load sample data on startup, pass the `-seed` flag with a path to a JSON seed file:

```bash
./bin/saldeti -port 9443 -seed examples/seed.json
```

A sample `seed.json` file is included in the repository with realistic test data:

- **Client**: `sim-client-id` / `sim-client-secret`, tenant `sim-tenant-id`
- **Admin user**: `admin@saldeti.local` with password `Simulator123!`
- **10 sample users**: Alice Smith, Bob Jones, Charlie Brown, Diana Prince, Eve Wilson, Frank Miller, Grace Lee (disabled), Henry Taylor, Ivan Guest (external), Julia Roberts
- **5 sample groups**: Engineering Team, Marketing Team, All Staff, Leadership (Private), Project Alpha (Unified/M365)
- **Pre-configured memberships and manager hierarchy**

### Seed File Format

The seed JSON file uses this schema:

```json
{
  "clients": [{ "client_id": "...", "client_secret": "...", "tenant_id": "..." }],
  "users": [{ "email": "...", "display_name": "...", "password": "...", "department": "...", "job_title": "..." }],
  "groups": [{ "display_name": "...", "description": "...", "mail_nickname": "..." }],
  "memberships": [{ "user_index": 0, "group_index": 0 }],
  "managers": [{ "user_index": 1, "manager_index": 0 }]
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

### Supported Endpoints

| Endpoint | Methods | Description |
|----------|---------|-------------|
| `/{tenant}/oauth2/v2.0/token` | POST | OAuth2 token exchange |
| `/v1.0/users` | GET, POST | List/create users |
| `/v1.0/users/{id}` | GET, PATCH, DELETE | Get/update/delete user |
| `/v1.0/users/{id}/memberOf` | GET | Groups user belongs to |
| `/v1.0/users/{id}/transitiveMemberOf` | GET | Transitive group membership |
| `/v1.0/users/{id}/manager` | GET | Get user's manager |
| `/v1.0/users/{id}/manager/$ref` | PUT, DELETE | Set/remove manager |
| `/v1.0/users/{id}/directReports` | GET | List direct reports |
| `/v1.0/users/{id}/checkMemberGroups` | POST | Check group membership |
| `/v1.0/users/{id}/getMemberGroups` | POST | Get all group memberships |
| `/v1.0/users/delta` | GET | Delta query for users |
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
| `/v1.0/groups/delta` | GET | Delta query for groups |
| `/v1.0/directoryObjects/getByIds` | POST | Batch object lookup |
| `/v1.0/me` | GET | Get authenticated user |

## Development

```bash
mise run build    # Build binary
mise run test     # Run all tests
mise run lint     # Run go vet
mise run clean    # Clean build artifacts
```

## License

GNU Affero General Public License v3
