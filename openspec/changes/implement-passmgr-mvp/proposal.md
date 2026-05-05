# implement-passmgr-mvp Proposal

## Summary

Implement the approved `passmgr` MVP as a local command-line password manager. The MVP provides a single encrypted local vault with initialization, entry CRUD, password generation, and clipboard copy with timed clearing.

## Scope

- Build a Go 1.22+ CLI application named `passmgr`.
- Store all vault data in a local encrypted file.
- Derive encryption keys from the master password with Argon2id.
- Encrypt vault payloads with AES-256-GCM.
- Support MVP commands: `init`, `add`, `get`, `list`, `rm`, `gen`, and `cp`.
- Keep `cmd/passmgr/main.go` thin and place business logic in focused internal packages.

## Files to Create or Modify

- `go.mod`: module and dependency declarations.
- `Makefile`: test/build convenience targets.
- `cmd/passmgr/main.go`: Cobra root command, flags, and subcommand registration.
- `internal/errno/errno.go`: typed application errors and CLI exit codes.
- `internal/config/config.go`: default vault path resolution with `PASSMGR_VAULT_PATH` override.
- `internal/store/store.go`: storage interface.
- `internal/store/file_store.go`: file-backed vault storage with private permissions.
- `internal/clipboard/clipboard.go`: clipboard interface and timed clear helper.
- `internal/clipboard/system_clip.go`: system clipboard adapter.
- `internal/passgen/passgen.go`: secure random password generation.
- `internal/vault/vault.go`: vault domain models and lifecycle.
- `internal/vault/crypto.go`: vault file format encode/decode.
- `pkg/crypto/crypto.go`: Argon2id and AES-256-GCM primitives.
- `tests/integration/cli_test.go`: executable CLI lifecycle tests.
- Package-local test files for every package above.
- `openspec/reports/`: generated test and coverage reports.
- `openspec/changes/implement-passmgr-mvp/verify.md`: final verification record.

## Test Strategy

Unit tests:

- `internal/errno/errno_test.go`: error codes, exit codes, unwrap behavior.
- `internal/config/config_test.go`: default and environment-driven vault paths.
- `pkg/crypto/crypto_test.go`: key derivation, encrypt/decrypt round trips, wrong password, tampering, random salt and nonce generation.
- `internal/store/file_store_test.go`: read/write/exists/path behavior and private file permissions.
- `internal/clipboard/clipboard_test.go`: interface behavior and timed clear using a fake clipboard.
- `internal/passgen/passgen_test.go`: length, character class options, invalid options, and secure randomness constraints.
- `internal/vault/vault_test.go`: init/open/save, wrong password, corrupted file, entry add/get/list/remove, duplicate handling.
- `cmd/passmgr` tests where practical for command wiring and non-interactive command flows.

Integration/E2E tests:

- Build the CLI and run against a temporary vault path.
- `passmgr init` creates an encrypted vault and does not store plaintext secrets.
- `passmgr add` stores an entry.
- `passmgr list` shows entry metadata but not the password.
- `passmgr get` masks the password by default and reveals it only with `--show-password`.
- `passmgr rm` removes an entry after confirmation.
- `passmgr gen` emits a password with requested constraints.
- Wrong master password fails with a non-zero exit code.

## Coverage Targets

- Lines: at least 80%.
- Branches: at least 70%.
- Functions: at least 80%.

For Go coverage, the gate will use `go test ./... -coverprofile=coverage.out` and convert the result into `openspec/reports/coverage-summary.json`. The JSON must record the targets above, even though Go's native tool does not separately report JavaScript-style branch coverage.

## Review Strategy

Each implementation task must receive two-stage review before proceeding:

- Specification compliance review against this proposal and `docs/superpowers/specs/2026-05-05-passmgr-cli-design.md`.
- Code quality review for security, maintainability, errors, and test adequacy.

Critical review findings block the next task.
