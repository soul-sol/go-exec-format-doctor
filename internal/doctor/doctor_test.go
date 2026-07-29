package doctor_test

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/soul-sol/go-exec-format-doctor/internal/doctor"
)

func Test_Inspect_reports_mismatch_when_ELF_arch_differs_from_host(t *testing.T) {
	t.Parallel()

	// Given
	path := writeFixture(t, "probe", elfHeader(0x3e))
	host, err := doctor.NewHost("linux", "arm64")
	require.NoError(t, err)
	target, err := doctor.NewTargetPath(path)
	require.NoError(t, err)

	// When
	report, err := doctor.Inspect(target, host)

	// Then
	require.NoError(t, err)
	require.Equal(t, doctor.FormatELF, report.Format)
	require.Equal(t, []string{"amd64"}, report.Architectures)
	require.Equal(t, doctor.VerdictMismatch, report.Verdict)
	require.Contains(t, report.Recommendations, "GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build")
	require.NotEmpty(t, report.FurtherHelpURL)
}

func Test_Inspect_reports_compatible_when_MachO_matches_host(t *testing.T) {
	t.Parallel()

	// Given
	path := writeFixture(t, "probe", machOHeader(0x0100000c))
	host, err := doctor.NewHost("darwin", "arm64")
	require.NoError(t, err)
	target, err := doctor.NewTargetPath(path)
	require.NoError(t, err)

	// When
	report, err := doctor.Inspect(target, host)

	// Then
	require.NoError(t, err)
	require.Equal(t, doctor.FormatMachO, report.Format)
	require.Equal(t, doctor.VerdictCompatible, report.Verdict)
	require.Empty(t, report.FurtherHelpURL)
}

func Test_Inspect_reports_compatible_when_fat_MachO_contains_host_arch(t *testing.T) {
	t.Parallel()

	// Given
	path := writeFixture(t, "probe", fatMachOHeader(0x01000007, 0x0100000c))
	host, err := doctor.NewHost("darwin", "arm64")
	require.NoError(t, err)
	target, err := doctor.NewTargetPath(path)
	require.NoError(t, err)

	// When
	report, err := doctor.Inspect(target, host)

	// Then
	require.NoError(t, err)
	require.Equal(t, doctor.FormatMachOFat, report.Format)
	require.Equal(t, []string{"amd64", "arm64"}, report.Architectures)
	require.Equal(t, doctor.VerdictCompatible, report.Verdict)
}

func Test_Inspect_reports_PE_target_architecture(t *testing.T) {
	t.Parallel()

	// Given
	path := writeFixture(t, "probe.exe", peHeader(0xaa64))
	host, err := doctor.NewHost("windows", "amd64")
	require.NoError(t, err)
	target, err := doctor.NewTargetPath(path)
	require.NoError(t, err)

	// When
	report, err := doctor.Inspect(target, host)

	// Then
	require.NoError(t, err)
	require.Equal(t, doctor.FormatPE, report.Format)
	require.Equal(t, []string{"arm64"}, report.Architectures)
	require.Equal(t, doctor.VerdictMismatch, report.Verdict)
}

func Test_Inspect_reports_script_as_conditional_on_Unix_host(t *testing.T) {
	t.Parallel()

	// Given
	path := writeFixture(t, "deploy", []byte("#!/usr/bin/env bash\necho ok\n"))
	host, err := doctor.NewHost("linux", "amd64")
	require.NoError(t, err)
	target, err := doctor.NewTargetPath(path)
	require.NoError(t, err)

	// When
	report, err := doctor.Inspect(target, host)

	// Then
	require.NoError(t, err)
	require.Equal(t, doctor.FormatScript, report.Format)
	require.Equal(t, doctor.VerdictConditional, report.Verdict)
	require.Equal(t, "/usr/bin/env bash", report.Interpreter)
	require.Contains(t, report.Summary, "interpreter")
}

func Test_Inspect_reports_script_mismatch_on_Windows_host(t *testing.T) {
	t.Parallel()

	// Given
	path := writeFixture(t, "deploy", []byte("#!/usr/bin/env bash\necho ok\n"))
	host, err := doctor.NewHost("windows", "amd64")
	require.NoError(t, err)
	target, err := doctor.NewTargetPath(path)
	require.NoError(t, err)

	// When
	report, err := doctor.Inspect(target, host)

	// Then
	require.NoError(t, err)
	require.Equal(t, doctor.VerdictMismatch, report.Verdict)
	require.Contains(t, report.Summary, "Windows")
}

func Test_Inspect_reports_archives_as_not_executable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		header []byte
		format doctor.Format
	}{
		{name: "ar", header: []byte("!<arch>\nmember"), format: doctor.FormatAR},
		{name: "zip", header: []byte("PK\x03\x04payload"), format: doctor.FormatZIP},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			path := writeFixture(t, test.name, test.header)
			host, err := doctor.NewHost("linux", "amd64")
			require.NoError(t, err)
			target, err := doctor.NewTargetPath(path)
			require.NoError(t, err)

			// When
			report, err := doctor.Inspect(target, host)

			// Then
			require.NoError(t, err)
			require.Equal(t, test.format, report.Format)
			require.Equal(t, doctor.VerdictNotExecutable, report.Verdict)
		})
	}
}

func Test_Inspect_reports_unknown_when_header_is_truncated(t *testing.T) {
	t.Parallel()

	// Given
	path := writeFixture(t, "broken", []byte{0x7f, 'E', 'L', 'F'})
	host, err := doctor.NewHost("linux", "amd64")
	require.NoError(t, err)
	target, err := doctor.NewTargetPath(path)
	require.NoError(t, err)

	// When
	report, err := doctor.Inspect(target, host)

	// Then
	require.NoError(t, err)
	require.Equal(t, doctor.FormatELF, report.Format)
	require.Equal(t, doctor.VerdictUnknown, report.Verdict)
	require.Contains(t, report.Summary, "truncated")
}

func writeFixture(t *testing.T, name string, data []byte) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, data, 0o600))

	return path
}

func elfHeader(machine uint16) []byte {
	header := make([]byte, 20)
	copy(header, []byte{0x7f, 'E', 'L', 'F'})
	header[4] = 2
	header[5] = 1
	binary.LittleEndian.PutUint16(header[18:20], machine)

	return header
}

func machOHeader(cpu uint32) []byte {
	header := make([]byte, 8)
	copy(header, []byte{0xcf, 0xfa, 0xed, 0xfe})
	binary.LittleEndian.PutUint32(header[4:8], cpu)

	return header
}

func fatMachOHeader(cpus ...uint32) []byte {
	header := make([]byte, 8+(20*len(cpus)))
	binary.BigEndian.PutUint32(header[0:4], 0xcafebabe)
	binary.BigEndian.PutUint32(
		header[4:8],
		uint32(len(cpus)), //nolint:gosec // Test fixtures contain at most two CPU values.
	)
	for index, cpu := range cpus {
		offset := 8 + (index * 20)
		binary.BigEndian.PutUint32(header[offset:offset+4], cpu)
	}

	return header
}

func peHeader(machine uint16) []byte {
	header := make([]byte, 70)
	copy(header, []byte{'M', 'Z'})
	binary.LittleEndian.PutUint32(header[0x3c:0x40], 0x40)
	copy(header[0x40:0x44], []byte{'P', 'E', 0, 0})
	binary.LittleEndian.PutUint16(header[0x44:0x46], machine)

	return header
}
