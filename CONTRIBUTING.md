# Contributing to saldeti

Thanks for your interest in contributing to saldeti! This document covers the basics of getting involved.

## Reporting Bugs

Bugs are tracked via [GitHub Issues](https://github.com/alexjplant/saldeti/issues). Before opening a new issue, please [search existing issues](https://github.com/alexjplant/saldeti/issues?q=is%3Aissue) to avoid duplicates. When filing a bug report, use the **Bug Report** template and include:

- A clear description of the problem
- Steps to reproduce
- Expected and actual behavior
- Your environment (OS, Go version, saldeti version)

## Requesting Features

Feature requests are also tracked via [GitHub Issues](https://github.com/alexjplant/saldeti/issues). Use the **Feature Request** template and describe:

- The problem you're trying to solve
- Your proposed solution
- Any alternatives you've considered

## Submitting Pull Requests

1. **Fork** the repository and clone your fork.
2. Create a feature branch from `main`:

   ```bash
   git checkout -b feat/my-feature
   ```

3. Make your changes, keeping commits focused.
4. Ensure the build, tests, and linters all pass (see [Build & Test](#build--test) below).
5. Push your branch and open a Pull Request against `main`.

When you open a PR, the PR template checklist will remind you of the requirements. Please link any related issue (e.g. `Closes #123`).

## Build & Test

saldeti uses [mise](https://mise.jdx.dev/) for task automation. After [installing mise](https://mise.jdx.dev/getting-started.html), the common tasks are:

| Task | Command | Description |
| --- | --- | --- |
| Build | `mise run build` | Build the `bin/saldeti` binary |
| Unit tests | `mise run test` | Run Go unit tests (excludes E2E) |
| Lint | `mise run lint` | Run `go vet ./...` |
| E2E tests | `mise run e2e` | Run Go E2E tests (Entra + Google) |
| UI E2E tests | `mise run ui-e2e` | Run Playwright E2E tests against the UI |

Run `mise tasks` to see the full list of available tasks.

## Code Style

- Format all Go code with `gofmt` (or `go fmt ./...`).
- Run `go vet ./...` (`mise run lint`) before submitting.
- Follow standard Go conventions and idioms.
- Write [Conventional Commits](https://www.conventionalcommits.org/) messages, e.g. `feat(mode): add google support`, `fix(auth): handle token refresh`, `docs: update readme`.

## License

By contributing, you agree that your contributions will be licensed under the AGPL-3.0 license.