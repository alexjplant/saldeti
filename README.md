# saldeti

A local API simulator for Microsoft Entra ID (Azure AD) and Google Workspace. saldeti mimics the Microsoft Graph API v1.0 and Google Workspace Admin SDK endpoints so you can develop and test directory-integrated applications without a real cloud tenant.

## Documentation

Full documentation is available at [alexjplant.github.io](https://alexjplant.github.io/saldeti/).

- [CLI Reference](https://alexjplant.github.io/saldeti/cli-reference) -- build flags, seed data, dump options
- [Seed Files](https://alexjplant.github.io/saldeti/seed-files) -- JSON schema and round-trip workflow
- [Entra ID Endpoints](https://alexjplant.github.io/saldeti/entra-endpoints) -- Microsoft Graph API coverage
- [Google Workspace Endpoints](https://alexjplant.github.io/saldeti/google-endpoints) -- Google Admin SDK coverage

## Quick Start

```bash
# Build
mise run build

# Run with Entra ID simulator (default)
./bin/saldeti -port 9443

# Run with the Google Workspace simulator (use -mode google; entra is the default and the two modes are mutually exclusive)
./bin/saldeti -port 9443 -mode google

# Run with seed data and persist changes on shutdown
./bin/saldeti -port 9443 -seed examples/seed.json -dump snapshot.json

# Run with Google Workspace seed data and persist changes on shutdown
./bin/saldeti -port 9443 -mode google -seed examples/google-seed.json -dump snapshot.json

# Get a token (credentials are logged at startup)
curl -X POST http://localhost:9443/<tenant-id>/oauth2/v2.0/token \
  -d "grant_type=client_credentials" \
  -d "client_id=<admin-client-id>" \
  -d "client_secret=<admin-client-secret>" \
  -d "scope=User.Read.All Group.Read.All"
```

Each mode uses its own distinct seed schema: Entra ID uses `examples/seed.json` (schema: `schema/seed.schema.json`) and Google Workspace uses `examples/google-seed.json` (schema: `schema/google-seed.schema.json`). See [Seed Files](https://alexjplant.github.io/saldeti/seed-files) for the file format.

Management UI available at `/ui` (entra mode, default) or `/google-ui/` (google mode) after starting the server.

## Development

```bash
mise run build        # Build binary
mise run test         # Run Go unit tests
mise run test-all     # Run all Go tests including E2E
mise run ui-test      # Run UI unit tests
mise run ui-e2e       # Run Playwright E2E tests
mise run lint         # Run go vet
mise run clean        # Clean build artifacts
```

## Questions

**What is this for?**  
Testing your apps and scripts against a make-believe directory so that you ~~don't write a garbage script that overwrites everybody's phone number with your own then task a 15-year-old helpdesk intern with manually fixing it while you run down the hall to save your job by finding somebody with a recent backup of the domain controller.~~ can perform integration testing in CI without spinning up a tenant.

**Is this vibe-coded?**  
Very yes. GLM-5.1 for orchestration and planning, GLM-4.7 for implementation, Gemini 3 Flash for UI iteration, and DeepSeek for review.

**...but why?**  
Because I wanted it and would rather spend time learning about LLMs and coding harnesses than manually replicating a Microsoft Azure product.

## Disclaimer

Saldeti is an independent, open-source project and is **not affiliated with, endorsed by, sponsored by, or officially connected to** Microsoft Corporation, Google LLC, or Alphabet Inc.

"Microsoft Entra ID", "Azure Active Directory", "Microsoft Graph", and related marks are trademarks of Microsoft Corporation. "Google Workspace", "Google Cloud", "Admin SDK", and related marks are trademarks of Google LLC. All product names, logos, and brands are property of their respective owners.

Saldeti simulates these APIs for local development and testing purposes only. It does not connect to, replicate, or access any real Microsoft or Google cloud services.

## License

[GNU Affero General Public License v3](https://www.gnu.org/licenses/agpl-3.0.en.html). Use it, improve it, don't make money on it.