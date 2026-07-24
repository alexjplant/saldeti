---
title: CLI Reference
description: Complete reference for the saldeti command-line flags and usage patterns.
---

# CLI Reference

`saldeti` is a single-binary API simulator for Microsoft Entra ID (Azure AD) and Google Workspace. It exposes realistic OAuth2 token endpoints and Microsoft Graph / Google Workspace API endpoints over HTTPS — perfect for local development and integration testing.

```
Usage: saldeti [options]
```

## Quick Start

### Entra ID mode (default)

Start the simulator with seed data containing users and app registrations:

```sh
saldeti -seed examples/seed.json
```

### Google Workspace mode

Run the Google Workspace API simulator exclusively:

```sh
saldeti -mode google
```

### Daemon mode

Run as a background process:

```sh
saldeti -daemon -seed examples/seed.json
```

### Stop a running daemon

```sh
saldeti -stop
```

### Custom TLS certificate

Provide your own TLS certificate and key instead of the auto-generated self-signed cert:

```sh
saldeti -tls-cert cert.pem -tls-key key.pem
```

### Behind a reverse proxy

When running behind a proxy like nginx or Caddy, set the external base URL:

```sh
saldeti -base-url https://example.com -trust-forwarded-headers
```

## Flags Reference

### General Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-port` | int | `9443` | Port to listen on. |
| `-ui` | bool | `true` | Enable the admin UI. When enabled, a web dashboard is available at `/ui` (Entra mode, default) or `/google-ui/` (Google mode). |
| `-seed` | string | `""` | Path to a JSON seed file. The format depends on the active `-mode` (`seed.schema.json` for Entra ID, `google-seed.schema.json` for Google Workspace). See [Seed Files](/seed-files) for the file format. |
| `-dump` | string | `""` | Path to write a seed JSON file on graceful shutdown. Useful for capturing runtime state. |
| `-domain` | string | `saldeti.local` | Default directory domain. Used for admin user UPN and for seeded users whose username does not contain `@`. |
| `-mode` | string | `entra` | Operating mode. Set to `google` to run the Google Workspace API simulator exclusively, or `entra` (default) for Microsoft Entra ID / Graph. The two modes are mutually exclusive. |
| `-debug` | bool | `false` | Enable debug-level logging. |

### Authentication & Security Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-signing-key` | string | `""` | JWT signing key. If empty, a random 32-byte key is generated on each startup — tokens from a previous run will not validate. |
| `-tls-cert` | string | `""` | Path to a TLS certificate file (PEM). If not set, a self-signed certificate is auto-generated. |
| `-tls-key` | string | `""` | Path to a TLS key file (PEM). If not set, a self-signed certificate is auto-generated. |
| `-base-url` | string | `""` | External base URL (e.g. `https://example.com`). Overrides any proxy header detection. When set, `X-Forwarded-Host` and `X-Forwarded-Proto` headers are ignored. |
| `-trust-forwarded-headers` | bool | `false` | Trust `X-Forwarded-Host` and `X-Forwarded-Proto` headers for base URL detection. Only enable this if the server is behind a trusted reverse proxy. |
| `-admin-client-id` | string | `""` | Admin application client ID. If empty, a random UUID is generated. (Entra mode only; ignored in google mode) |
| `-admin-client-secret` | string | `""` | Admin application client secret. If empty, a random UUID is generated. (Entra mode only; ignored in google mode) |
| `-admin-tenant-id` | string | `""` | Admin application tenant ID. If empty, a random UUID is generated. (Entra mode only; ignored in google mode) |

### Daemon Mode Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-daemon` | bool | `false` | Run as a background daemon. The parent process forks a child, writes its PID, and exits. |
| `-pidfile` | string | `saldeti.pid` | Path to the PID file. Used by both `-daemon` (to write) and `-stop` (to read). |
| `-logfile` | string | `saldeti.log` | Path to the log file. All stdout and stderr from the daemon child process are written here. |
| `-stop` | bool | `false` | Stop a running daemon. Reads the PID file and sends `SIGTERM`. If the process does not stop within 10 seconds, it is force-killed. |

## Flag Interactions

### Admin credentials are all-or-nothing

The three admin credential flags (`-admin-client-id`, `-admin-client-secret`, `-admin-tenant-id`) must all be set or all be left empty. If you set one, you must set all three — otherwise the server will refuse to start.

These flags only apply in entra mode. In google mode, they are ignored.

::: tip
If you leave all three empty, Saldeti generates random UUIDs and prints them to the console at startup. This is ideal for local development where you just need valid credentials quickly.
:::

### Self-signed TLS is automatic

When `-tls-cert` and `-tls-key` are both empty, Saldeti auto-generates an ECDSA/P-256 self-signed certificate valid for 365 days, covering `localhost`, `127.0.0.1`, and `::1`.

::: warning
Browsers will show a certificate warning for self-signed certs. You need to accept the warning (or import the cert into your trust store) before the admin UI will load.
:::

### Base URL overrides forwarded headers

When `-base-url` is set, it takes absolute precedence — the `-trust-forwarded-headers` flag has no effect because the base URL is explicitly known. Use `-trust-forwarded-headers` only when you want Saldeti to detect its own base URL from proxy headers.

### Daemon mode lifecycle

When `-daemon` is set, the parent process:

1. Checks for an existing PID file and verifies the process is not already running.
2. Opens the log file for the child process.
3. Forks a child process with `SALDETI_CHILD=1` in its environment.
4. Writes the child PID to the PID file.
5. Prints the PID, log path, and stop command, then exits.

The child process runs the actual server. All command-line flags are inherited by the child.

### Stopping the daemon

`saldeti -stop` reads the PID file (default `saldeti.pid`), sends `SIGTERM` to the process, and waits up to 10 seconds for it to exit gracefully. If it does not exit, it is force-killed with `SIGKILL`. The PID file is removed after the process exits.

You can specify a custom PID file path if the daemon was started with `-pidfile`:

```sh
saldeti -stop -pidfile /var/run/saldeti.pid
```

---

::: warning Trademark Notice
Saldeti is an independent project and is not affiliated with, endorsed by, or sponsored by Microsoft Corporation or Google LLC. All product names and trademarks are property of their respective owners.
:::