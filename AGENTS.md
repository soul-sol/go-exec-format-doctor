# AGENTS.md

Go 1.23+ read-only executable-header diagnostic CLI.

## Commands

- `task ci` — format check, module verification, vet, lint, nil check, tests
- `task build` — build `bin/go-exec-format-doctor`

## Architecture

- `cmd/go-exec-format-doctor/main.go` — process entrypoint only
- `internal/app/` — CLI parsing, rendering, exit codes
- `internal/doctor/` — typed domain values and header inspection

## Invariants

- Never execute, modify, upload, or make network requests with inspected files.
- Never claim malware or safety assessment.
- Parse CLI input once into `doctor.Host` and `doctor.TargetPath`.
- Errors are wrapped with `%w`; no ignored errors or panic in library code.
- Keep each Go file below 250 pure lines.
