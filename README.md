# go-exec-format-doctor

Diagnose `exec format error` and binary architecture mismatches without running
the file.

`go-exec-format-doctor` reads local file headers and reports:

- ELF, Mach-O, universal Mach-O, and PE target architectures
- shebang scripts and their requested interpreter
- ar and ZIP archives that are not directly executable
- truncated or unknown headers
- the current host OS and architecture
- a specific pure-Go rebuild command when a binary does not match the host

It does not execute, modify, upload, or make network requests with the inspected
file. It does not determine whether a file is safe or malicious.

## Install

Download the archive for your platform from
[GitHub Releases](https://github.com/soul-sol/go-exec-format-doctor/releases),
verify it against `checksums.txt`, and place the binary on your `PATH`.

Or build from source with Go 1.23 or newer:

```sh
go install github.com/soul-sol/go-exec-format-doctor/cmd/go-exec-format-doctor@latest
```

## Use

```sh
go-exec-format-doctor ./my-service
```

Example mismatch:

```text
format: elf
target: linux/amd64
host: linux/arm64
verdict: mismatch
summary: elf binary targets linux/amd64 but this host is linux/arm64
next: GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build
```

Machine-readable output:

```sh
go-exec-format-doctor --json ./my-service
```

Suppress the optional further-help link:

```sh
go-exec-format-doctor --quiet ./my-service
```

Exit codes:

| Code | Meaning |
|---:|---|
| `0` | compatible binary or conditionally runnable script |
| `1` | architecture/OS mismatch, archive, or unknown format |
| `2` | invalid CLI input, unreadable path, or output failure |

## What the verdict means

- `compatible`: the file header includes the current OS and architecture.
- `mismatch`: the file header targets another OS or architecture.
- `conditional`: a script has a shebang, but the interpreter is not probed.
- `not-executable`: the file is an archive container.
- `unknown`: the header is truncated, corrupt, or not recognized.

Header compatibility is not a full runtime guarantee. Dynamic libraries,
permissions, mount flags, kernel features, CGO dependencies, and interpreter
availability can still prevent execution.

## Going further

If you want to prevent architecture mismatches in CI, the
[USD 29 Go/Linux Cross-Architecture CI Starter Kit](https://soul-sol.github.io/verified-automation-services/go-cross-architecture-ci-kit.html)
adds tested AMD64, ARM64, and 386 workflows, compile reports, and optional QEMU
smoke tests. The free doctor remains complete and does not require the kit.

## Development

```sh
go test -race -shuffle=on -count=1 ./...
go vet ./...
golangci-lint run ./...
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for scope and verification requirements.

## License

MIT. See [LICENSE](LICENSE).
