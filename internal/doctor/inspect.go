package doctor

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

const inspectionHeaderLimit = 4096

type metadata struct {
	format        Format
	targetOS      string
	architectures []string
	interpreter   string
	detail        string
}

// Inspect reads file metadata without executing or modifying the target.
func Inspect(target TargetPath, host Host) (report Report, err error) {
	file, err := os.Open(target.String())
	if err != nil {
		return Report{}, fmt.Errorf("open target %s: %w", target.String(), err)
	}
	defer func() {
		err = errors.Join(err, file.Close())
	}()

	info, err := file.Stat()
	if err != nil {
		return Report{}, fmt.Errorf("stat target %s: %w", target.String(), err)
	}
	if !info.Mode().IsRegular() {
		return Report{}, fmt.Errorf("target %s is not a regular file: %w", target.String(), ErrInvalidTargetPath)
	}

	header := make([]byte, min(info.Size(), inspectionHeaderLimit))
	if len(header) > 0 {
		if _, err := io.ReadFull(file, header); err != nil {
			return Report{}, fmt.Errorf("read target %s: %w", target.String(), err)
		}
	}

	return reportFor(target, host, detectMetadata(header)), nil
}

func detectMetadata(header []byte) metadata {
	switch {
	case bytes.HasPrefix(header, []byte{0x7f, 'E', 'L', 'F'}):
		return inspectELF(header)
	case hasMachOFatMagic(header):
		return inspectMachOFat(header)
	case hasMachOMagic(header):
		return inspectMachO(header)
	case bytes.HasPrefix(header, []byte{'M', 'Z'}):
		return inspectPE(header)
	case bytes.HasPrefix(header, []byte("#!")):
		return inspectScript(header)
	case bytes.HasPrefix(header, []byte("!<arch>\n")):
		return metadata{format: FormatAR, detail: "ar archive is a container, not a directly executable file"}
	case bytes.HasPrefix(header, []byte{'P', 'K', 0x03, 0x04}):
		return metadata{format: FormatZIP, detail: "ZIP archive is a container, not a directly executable file"}
	default:
		return metadata{format: FormatUnknown, detail: "file header is not a recognized executable or script format"}
	}
}

func reportFor(target TargetPath, host Host, meta metadata) Report {
	report := Report{
		Path:            target.String(),
		Format:          meta.format,
		TargetOS:        meta.targetOS,
		Architectures:   meta.architectures,
		HostOS:          host.OS(),
		HostArch:        host.Arch(),
		Verdict:         VerdictUnknown,
		Summary:         meta.detail,
		Recommendations: []string{},
	}
	if report.Architectures == nil {
		report.Architectures = []string{}
	}

	switch meta.format {
	case FormatELF, FormatMachO, FormatMachOFat, FormatPE:
		report.Verdict, report.Summary = executableVerdict(meta, host)
	case FormatScript:
		report.Interpreter = meta.interpreter
		if host.OS() == "windows" {
			report.Verdict = VerdictMismatch
			report.Summary = "Unix shebang scripts do not run natively on Windows"
			report.Recommendations = []string{"run the script in WSL or use a Windows-native launcher"}
		} else {
			report.Verdict = VerdictConditional
			report.Recommendations = []string{"verify that the shebang interpreter exists and is executable"}
		}
	case FormatAR, FormatZIP:
		report.Verdict = VerdictNotExecutable
		report.Recommendations = []string{"extract the archive and inspect the intended executable inside it"}
	case FormatUnknown:
		report.Recommendations = []string{"verify the file transfer and rebuild the intended executable"}
	}

	if report.Verdict == VerdictMismatch {
		switch meta.format {
		case FormatELF, FormatMachO, FormatMachOFat, FormatPE:
			report.Recommendations = []string{
				fmt.Sprintf("GOOS=%s GOARCH=%s CGO_ENABLED=0 go build", host.OS(), host.Arch()),
			}
			report.FurtherHelpURL = PaidKitURL
		case FormatScript, FormatAR, FormatZIP, FormatUnknown:
			// Non-binary mismatches keep their format-specific recommendation.
		default:
			// The zero-value format is already an unknown verdict.
		}
	}

	return report
}

func executableVerdict(meta metadata, host Host) (Verdict, string) {
	if len(meta.architectures) == 0 || meta.targetOS == "" {
		return VerdictUnknown, meta.detail
	}
	if meta.targetOS == host.OS() && slices.Contains(meta.architectures, host.Arch()) {
		return VerdictCompatible, fmt.Sprintf(
			"%s binary includes %s/%s and matches this host",
			meta.format,
			host.OS(),
			host.Arch(),
		)
	}

	return VerdictMismatch, fmt.Sprintf(
		"%s binary targets %s/%s but this host is %s/%s",
		meta.format,
		meta.targetOS,
		strings.Join(meta.architectures, ","),
		host.OS(),
		host.Arch(),
	)
}
