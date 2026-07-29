// Package doctor inspects executable file headers without executing files.
package doctor

import (
	"errors"
	"fmt"
	"strings"
)

// PaidKitURL is optional further help shown only after a binary mismatch.
const PaidKitURL = "https://soul-sol.github.io/verified-automation-services/go-cross-architecture-ci-kit.html"

var (
	// ErrInvalidHost identifies an empty host OS or architecture.
	ErrInvalidHost = errors.New("doctor: invalid host")
	// ErrInvalidTargetPath identifies an empty or non-regular target path.
	ErrInvalidTargetPath = errors.New("doctor: invalid target path")
)

// Format identifies a file container or executable format.
type Format string

const (
	// FormatUnknown identifies an unrecognized file header.
	FormatUnknown Format = "unknown"
	// FormatELF identifies an Executable and Linkable Format file.
	FormatELF Format = "elf"
	// FormatMachO identifies a single-architecture Mach-O file.
	FormatMachO Format = "mach-o"
	// FormatMachOFat identifies a universal multi-architecture Mach-O file.
	FormatMachOFat Format = "mach-o-universal"
	// FormatPE identifies a Windows Portable Executable file.
	FormatPE Format = "pe"
	// FormatScript identifies a shebang script.
	FormatScript Format = "script"
	// FormatAR identifies an ar archive.
	FormatAR Format = "ar-archive"
	// FormatZIP identifies a ZIP archive.
	FormatZIP Format = "zip-archive"
)

// Verdict describes whether the inspected file can run directly on the host.
type Verdict string

const (
	// VerdictCompatible means the file header matches the current host.
	VerdictCompatible Verdict = "compatible"
	// VerdictMismatch means the target OS or architecture differs from the host.
	VerdictMismatch Verdict = "mismatch"
	// VerdictConditional means a script still depends on its interpreter.
	VerdictConditional Verdict = "conditional"
	// VerdictNotExecutable means the file is a container rather than an executable.
	VerdictNotExecutable Verdict = "not-executable"
	// VerdictUnknown means the tool could not prove compatibility.
	VerdictUnknown Verdict = "unknown"
)

// Host is a parsed operating-system and architecture pair.
type Host struct {
	os   string
	arch string
}

// NewHost parses a host pair from runtime or test input.
func NewHost(rawOS, rawArch string) (Host, error) {
	hostOS := strings.TrimSpace(strings.ToLower(rawOS))
	hostArch := strings.TrimSpace(strings.ToLower(rawArch))
	if hostOS == "" || hostArch == "" {
		return Host{}, fmt.Errorf("os=%q arch=%q: %w", rawOS, rawArch, ErrInvalidHost)
	}

	return Host{os: hostOS, arch: hostArch}, nil
}

// OS returns the parsed host operating system.
func (h Host) OS() string {
	return h.os
}

// Arch returns the parsed host architecture.
func (h Host) Arch() string {
	return h.arch
}

// TargetPath is a non-empty path to an inspection target.
type TargetPath struct {
	raw string
}

// NewTargetPath parses a path supplied at the CLI boundary.
func NewTargetPath(raw string) (TargetPath, error) {
	path := strings.TrimSpace(raw)
	if path == "" {
		return TargetPath{}, ErrInvalidTargetPath
	}

	return TargetPath{raw: path}, nil
}

// String returns the target path.
func (p TargetPath) String() string {
	return p.raw
}

// Report is the stable human and JSON inspection contract.
type Report struct {
	Path            string   `json:"path"`
	Format          Format   `json:"format"`
	TargetOS        string   `json:"target_os"`
	Architectures   []string `json:"architectures"`
	HostOS          string   `json:"host_os"`
	HostArch        string   `json:"host_arch"`
	Verdict         Verdict  `json:"verdict"`
	Summary         string   `json:"summary"`
	Interpreter     string   `json:"interpreter,omitempty"`
	Recommendations []string `json:"recommendations"`
	FurtherHelpURL  string   `json:"further_help_url,omitempty"`
}
