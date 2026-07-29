package app_test

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/soul-sol/go-exec-format-doctor/internal/app"
	"github.com/soul-sol/go-exec-format-doctor/internal/doctor"
)

func Test_Run_emits_JSON_and_mismatch_exit_when_requested(t *testing.T) {
	t.Parallel()

	// Given
	path := writeELF(t, 0x3e)
	host, err := doctor.NewHost("linux", "arm64")
	require.NoError(t, err)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// When
	exitCode := app.Run(
		[]string{"--json", path},
		app.Streams{Stdout: &stdout, Stderr: &stderr},
		host,
	)

	// Then
	require.Equal(t, 1, exitCode)
	require.Empty(t, stderr.String())
	var report doctor.Report
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &report))
	require.Equal(t, doctor.VerdictMismatch, report.Verdict)
	require.Equal(t, []string{"amd64"}, report.Architectures)
}

func Test_Run_suppresses_paid_link_when_quiet(t *testing.T) {
	t.Parallel()

	// Given
	path := writeELF(t, 0x3e)
	host, err := doctor.NewHost("linux", "arm64")
	require.NoError(t, err)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// When
	exitCode := app.Run(
		[]string{"--quiet", path},
		app.Streams{Stdout: &stdout, Stderr: &stderr},
		host,
	)

	// Then
	require.Equal(t, 1, exitCode)
	require.NotContains(t, stdout.String(), "starter kit")
	require.Contains(t, stdout.String(), "GOOS=linux GOARCH=arm64")
	require.Empty(t, stderr.String())
}

func Test_Run_returns_usage_exit_when_path_is_missing(t *testing.T) {
	t.Parallel()

	// Given
	host, err := doctor.NewHost("linux", "amd64")
	require.NoError(t, err)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// When
	exitCode := app.Run(nil, app.Streams{Stdout: &stdout, Stderr: &stderr}, host)

	// Then
	require.Equal(t, 2, exitCode)
	require.Empty(t, stdout.String())
	require.Contains(t, stderr.String(), "usage:")
}

func writeELF(t *testing.T, machine uint16) string {
	t.Helper()

	header := make([]byte, 20)
	copy(header, []byte{0x7f, 'E', 'L', 'F'})
	header[4] = 2
	header[5] = 1
	binary.LittleEndian.PutUint16(header[18:20], machine)
	path := filepath.Join(t.TempDir(), "probe")
	require.NoError(t, os.WriteFile(path, header, 0o600))

	return path
}
