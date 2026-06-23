---
title: Seed Files
description: Reference guide for Entra ID and Google Workspace seed file formats used by Saldeti.
---

# Seed Files

Seed files let you pre-populate the Saldeti simulator with realistic data on startup. Instead of starting from an empty directory, you provide a JSON file that defines users, groups, devices, and other resources — ready for your integration tests to consume immediately.

## Overview

Pass a seed file with the `-seed` flag (the file loaded depends on the active `-mode`):

```sh
saldeti -seed examples/seed.json                       # Entra ID mode (default)
saldeti -mode google -seed examples/google-seed.json   # Google Workspace mode
```

::: tip
The `-seed` flag loads whichever mode is active. Use `-mode entra -seed examples/seed.json` for Entra ID or `-mode google -seed examples/google-seed.json` for Google Workspace.
:::

The simulator parses the file at startup and creates all the declared resources in memory. There is no database — everything lives in process memory, so restarting with a different seed file gives you a completely fresh environment.

You can also capture the current state of a running simulator into a seed file with `-dump`:

```sh
saldeti -seed examples/seed.json -dump state.json
```

On graceful shutdown, the full in-memory state is written to `state.json`.

::: tip
Seed files are plain JSON. You can generate them programmatically, version-control them alongside your tests, or share them across teams for reproducible test environments.
:::

## Entra ID Mode Seed File

The Entra ID seed file follows the schema defined in [`seed.schema.json`](https://raw.githubusercontent.com/saldeti/saldeti/main/schema/seed.schema.json).

### Top-level properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `clients` | array | **Yes** | OAuth2 client credentials. Each entry has `client_id`, `client_secret`, and `tenant_id`. |
| `users` | array | No | Entra ID users. Each user requires `email`, `display_name`, and `password`. |
| `groups` | array | No | Entra ID groups. Each group requires `display_name`. |
| `memberships` | array | No | Explicit group memberships using `user_index` and `group_index`. |
| `ownerships` | array | No | Group ownerships using `user_index` and `group_index`. |
| `managers` | array | No | User-manager relationships using `user_index` and `manager_index`. |
| `applications` | array | No | Application registrations with optional app roles and owner UPNs. |
| `service_principals` | array | No | Service principals linked to application registrations. |
| `app_role_assignments` | array | No | App role assignments linking principals to resources. |
| `oauth2_grants` | array | No | OAuth2 permission grants (delegated or application). |

### Minimal Entra ID example

```json
{
  "clients": [
    {
      "client_id": "sim-client-id",
      "client_secret": "sim-client-secret",
      "tenant_id": "sim-tenant-id"
    }
  ],
  "users": [
    {
      "email": "admin@saldeti.local",
      "display_name": "Admin User",
      "password": "Simulator123!",
      "given_name": "Admin",
      "surname": "User"
    }
  ],
  "groups": [
    {
      "display_name": "Engineering Team",
      "description": "Engineering department",
      "member_upns": ["admin@saldeti.local"]
    }
  ]
}
```

### Cross-referencing resources

Entra ID seed files support two cross-referencing styles:

- **By UPN/email**: Use `member_upns`, `owner_upns`, and `manager_upn` to reference users by their email address.
- **By index**: Use `memberships`, `ownerships`, and `managers` arrays with zero-based `user_index` and `group_index` values for precise control.

::: warning
Index-based references are order-dependent. If you reorder the `users` array, all index values must be updated accordingly.
:::

## Google Workspace Mode Seed File

The Google Workspace seed file follows the schema defined in [`google-seed.schema.json`](https://raw.githubusercontent.com/saldeti/saldeti/main/schema/google-seed.schema.json).

::: tip
Google Workspace seed loading uses the `-seed` flag in google mode. Use `-mode google -seed examples/google-seed.json` to load Google Workspace data on startup.
:::

### Top-level properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `clients` | array | **Yes** | OAuth2 client credentials. Each entry has `client_id` and `client_secret`. |
| `users` | array | No | Google Workspace users. Each user requires `primary_email`. |
| `groups` | array | No | Google groups. Each group requires `email`. |
| `org_units` | array | No | Organizational units. Each requires `name`. |
| `roles` | array | No | Custom admin roles. Each requires `role_name`. |
| `role_assignments` | array | No | Role assignments linking users to roles. |
| `domains` | array | No | Domains associated with the Google Workspace account. |
| `chromeos_devices` | array | No | Chrome OS devices to register. |
| `mobile_devices` | array | No | Mobile devices to register. |
| `group_settings` | array | No | Per-group settings (Groups Settings API). Each requires `group_email`. |

### Minimal Google Workspace example

```json
{
  "clients": [
    {
      "client_id": "sim-client-id.apps.googleusercontent.com",
      "client_secret": "GOCSPX-sim-client-secret"
    }
  ],
  "users": [
    {
      "primary_email": "admin@example.com",
      "given_name": "Admin",
      "family_name": "User",
      "password": "Simulator123!",
      "is_admin": true
    }
  ],
  "groups": [
    {
      "email": "engineering@example.com",
      "name": "Engineering Team",
      "member_emails": ["admin@example.com"]
    }
  ],
  "org_units": [
    {
      "name": "Engineering",
      "description": "Engineering department",
      "parent_org_unit_path": "/"
    }
  ]
}
```

### Cross-referencing resources

Google Workspace seed files use email-based cross-referencing:

- **`member_emails`** in groups references user `primary_email` values.
- **`parent_org_unit_path`** in org units references the path of a parent org unit (e.g., `/Engineering`).
- **`assigned_to_email`** in role assignments references user `primary_email`.
- **`role_id_or_name`** in role assignments references either a system role ID or the `role_name` of a custom role defined in the `roles` array.
- **`org_unit_path`** in users and devices references org unit paths.

::: tip
Org unit paths follow the Google Workspace convention: the root is `/`, and nested paths are `/Parent/Child`. Make sure parent org units are listed before their children.
:::

## JSON Schema Validation

Both seed file formats have machine-readable JSON Schema definitions that you can use to validate your seed files before passing them to the simulator.

### Schema locations

| Schema | URL |
|--------|-----|
| Entra ID | [`seed.schema.json`](https://raw.githubusercontent.com/saldeti/saldeti/main/schema/seed.schema.json) |
| Google Workspace | [`google-seed.schema.json`](https://raw.githubusercontent.com/saldeti/saldeti/main/schema/google-seed.schema.json) |

### Validating with Python

```sh
pip install jsonschema
python -m jsonschema schema/seed.schema.json -i examples/seed.json
python -m jsonschema schema/google-seed.schema.json -i examples/google-seed.json
```

### Validating with Node.js (ajv-cli)

```sh
npx ajv-cli validate -s schema/seed.schema.json -d examples/seed.json
npx ajv-cli validate -s schema/google-seed.schema.json -d examples/google-seed.json
```

### Validating with Go

The project includes a schema generator that also validates both example seed files:

```sh
go run cmd/genschema/main.go
```

This regenerates `schema/seed.schema.json` from the Go types in `internal/entra/seed/schema.go`, `schema/google-seed.schema.json` from the Go types in `internal/google/seed/schema.go`, and validates both `examples/seed.json` and `examples/google-seed.json` against their respective schemas.

### Online validation

Paste your seed file and the corresponding schema into any JSON Schema validator such as [https://www.jsonschemavalidator.net/](https://www.jsonschemavalidator.net/).

---

::: warning Trademark Notice
Saldeti is an independent project and is not affiliated with, endorsed by, or sponsored by Microsoft Corporation or Google LLC. All product names and trademarks are property of their respective owners.
:::