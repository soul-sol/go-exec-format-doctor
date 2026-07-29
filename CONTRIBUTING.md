# Contributing

## Scope

This project performs local, read-only file-header inspection. Changes must not
execute or modify the inspected file, make network requests, add telemetry, or
claim to assess whether a file is safe.

## Verification

Run:

```sh
go mod verify
gofumpt -l .
go vet ./...
golangci-lint run ./...
nilaway ./...
go test -race -shuffle=on -count=1 ./...
```

New behavior requires a failing test first and must include an observable CLI
test when it changes output or exit status.
