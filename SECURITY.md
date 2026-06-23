# Security Policy

## Reporting a Vulnerability

If you believe you have found a security vulnerability in saldeti, please report it responsibly:

- Open a **private** [GitHub Security Advisory](https://github.com/alexjplant/saldeti/security/advisories/new), **or**
- Email **security@example.com**

Please **do not** open a public GitHub issue for security-related reports. We aim to acknowledge reports within 72 hours.

## Scope

saldeti is a **local development and testing simulator**. It is **not** production infrastructure and is not intended to be exposed to untrusted networks.

Key scope limitations:

- **No real cloud connections.** saldeti does not contact Microsoft Graph, Google Workspace, or any real cloud service. It only mimics their API responses.
- **No real data.** All directory data is synthetic, loaded from local JSON seed files.
- **Local-only HTTPS.** The server uses a self-signed TLS certificate generated at startup. Clients must disable certificate verification (e.g. `curl -k` or `NODE_TLS_REJECT_UNAUTHORIZED=0`) to connect.
- **No authentication secrets of value.** All client IDs, secrets, and tokens are randomly generated at startup and have no real-world validity.

Because of this, the practical attack surface is limited to the developer's own machine. Reports involving intentional exposure of the simulator to untrusted networks, or use of real credentials/secrets, are out of scope.

## Supported Versions

Only the latest release line receives security updates.

| Version | Supported |
| --- | --- |
| latest | ✅ |
| older | ❌ |