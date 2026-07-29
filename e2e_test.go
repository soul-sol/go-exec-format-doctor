package execformatdoctor_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/soul-sol/go-exec-format-doctor/internal/doctor"
)

func Test_E2E_CLI_reports_compatible_for_its_own_binary(t *testing.T) {
	// Given
	binary := buildDoctor(t, runtime.GOOS, runtime.GOARCH)
	command := exec.Command(binary, "--json", binary) //nolint:gosec // Test invokes a binary built into t.TempDir.

	// When
	output, err := command.Output()

	// Then
	require.NoError(t, err)
	var report doctor.Report
	require.NoError(t, json.Unmarshal(output, &report))
	require.Equal(t, doctor.VerdictCompatible, report.Verdict)
	require.Equal(t, runtime.GOOS, report.HostOS)
	require.Contains(t, report.Architectures, runtime.GOARCH)
}

func Test_E2E_CLI_reports_mismatch_for_other_architecture(t *testing.T) {
	// Given
	hostBinary := buildDoctor(t, runtime.GOOS, runtime.GOARCH)
	targetArch := "arm64"
	if runtime.GOARCH == "arm64" {
		targetArch = "amd64"
	}
	foreignBinary := buildDoctor(t, runtime.GOOS, targetArch)
	command := exec.Command(hostBinary, "--json", foreignBinary) //nolint:gosec // Both paths are controlled t.TempDir builds.

	// When
	output, err := command.Output()

	// Then
	var exitError *exec.ExitError
	require.ErrorAs(t, err, &exitError)
	require.Equal(t, 1, exitError.ExitCode())
	var report doctor.Report
	require.NoError(t, json.Unmarshal(output, &report))
	require.Equal(t, doctor.VerdictMismatch, report.Verdict)
	require.Contains(t, report.Architectures, targetArch)
	require.Equal(t, doctor.PaidKitURL, report.FurtherHelpURL)
}

func buildDoctor(t *testing.T, targetOS, targetArch string) string {
	t.Helper()

	output := filepath.Join(t.TempDir(), "doctor-"+targetOS+"-"+targetArch)
	if targetOS == "windows" {
		output += ".exe"
	}
	command := exec.Command( //nolint:gosec // Test runs the Go toolchain with controlled arguments.
		"go",
		"build",
		"-trimpath",
		"-o",
		output,
		"./cmd/go-exec-format-doctor",
	)
	command.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+targetOS, "GOARCH="+targetArch)
	combined, err := command.CombinedOutput()
	require.NoError(t, err, string(combined))

	return output
}
