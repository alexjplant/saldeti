# Changelog

All notable changes to saldeti are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Google Workspace **dump-on-shutdown** support (`-dump`), allowing Google mode state to be persisted between runs.
- `group_settings` to the Google seed schema for completeness.
- `healthz` endpoint and improved refresh-token cleanup context in the Entra auth flow.
- Benchmarks and expanded batch test coverage across directory operations.
- Community and governance files: `CONTRIBUTING.md`, `SECURITY.md`, `CODE_OF_CONDUCT.md`, and GitHub issue/PR templates.
- Google Go E2E tests and Google Playwright E2E tests wired into `mise` and CI.

### Changed

- **Mode exclusivity refactor:** replaced the separate `-google` / `-google-seed` flags with a single mutually exclusive `-mode entra|google` flag (`entra` is the default). Each mode now uses its own distinct seed schema.
- Schema distinctness and completeness improvements: preserved `@odata.type` in Graph responses, removed `omitempty` from required model fields, and standardized filter checks and naming.
- Standardized trademark/non-affiliation disclaimers across the README, docs site, and UI footers.
- Corrected seed-schema documentation in the README, CLI reference, and Google endpoints docs.
- Aligned CI docs and fixed the VitePress social link and TypeScript version.

### Fixed

- Eliminated a **refresh-token deadlock** in the Entra authentication flow.
- Replaced a `panic()` with a returned error during UI router setup, preventing a server crash on startup failure.
- Google dump pagination and ChromeOS status round-trip correctness.
- Batch marshalling error handling; completed `TestAssignLicense`.
- `fs.Sub` error handling, `ui-e2e-headed` task, and `RegisterClient` validation.
- HTTPS URL and GitHub org/schema link corrections in docs and README.
- Removed dead code in `DeleteCIMembership`.

### Security

- Removed the JWT signing key from startup logs.
- Fixed OIDC trust headers and documented the authorization code flow.

### Removed

- The `-google` and `-google-seed` command-line flags (superseded by `-mode`).
- Tracked build binary removed from the repository and added to `.gitignore`.

### Deprecated

- Nothing yet.

[Unreleased]: https://github.com/alexjplant/saldeti/commits/main