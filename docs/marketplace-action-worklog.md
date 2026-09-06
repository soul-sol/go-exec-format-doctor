# GitHub Marketplace action worklog

## Goal and scope

Expose the existing read-only executable-header doctor as a composite GitHub
Action. The action builds the repository command, runs its supported `--json`
and optional `--quiet` modes against one caller-supplied path, exports the
report, and preserves the command's gate exit code.

The action does not execute, modify, upload, or make network requests with the
inspected file. It does not add an expected-platform mode that the command does
not support; compatibility is evaluated against the GitHub runner host.

## Acceptance and validation

- Root `action.yml` has a Marketplace-specific name, one-sentence description,
  branding, documented inputs, and documented outputs.
- README starts with a copyable workflow using
  `soul-sol/go-exec-format-doctor@v1`.
- `.github/workflows/selftest.yml` runs the local action against native AMD64
  and foreign ARM64 Go binaries.
- YAML metadata parses, repository Go tests pass, and local positive/negative
  action equivalents return the expected exit codes and JSON verdicts.

## Source and ledger status

The GitHub clone attempt was blocked by sandbox DNS. Implementation used the
clean local clone whose `origin` is
`https://github.com/soul-sol/go-exec-format-doctor.git` at commit
`df61ebff1c67f25521282b163ada88644143cf96`. Remote freshness could not be
checked in this environment.

Project-Tree sync to `https://project-tree.lifestep.io` was attempted before
implementation and failed because DNS resolution is disabled (`http_status=000`).
This document is the required repository-local fallback record.
