// Package app owns the command-line boundary and output rendering.
package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/soul-sol/go-exec-format-doctor/internal/doctor"
)

// Streams contains the command's output destinations.
type Streams struct {
	Stdout io.Writer
	Stderr io.Writer
}

// Run parses CLI input, inspects one file, and returns a process exit code.
func Run(args []string, streams Streams, host doctor.Host) int {
	flags := flag.NewFlagSet("go-exec-format-doctor", flag.ContinueOnError)
	flags.SetOutput(streams.Stderr)
	jsonOutput := flags.Bool("json", false, "emit a JSON report")
	quiet := flags.Bool("quiet", false, "suppress the optional paid-kit link")

	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 1 {
		return writeFailure(streams.Stderr, "usage: go-exec-format-doctor [--json] [--quiet] FILE")
	}

	target, err := doctor.NewTargetPath(flags.Arg(0))
	if err != nil {
		return writeFailure(streams.Stderr, fmt.Sprintf("input error: %v", err))
	}
	report, err := doctor.Inspect(target, host)
	if err != nil {
		return writeFailure(streams.Stderr, fmt.Sprintf("inspection error: %v", err))
	}
	if *quiet {
		report.FurtherHelpURL = ""
	}

	if *jsonOutput {
		encoder := json.NewEncoder(streams.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return writeFailure(streams.Stderr, fmt.Sprintf("output error: %v", err))
		}
	} else if err := writeHuman(streams.Stdout, report); err != nil {
		return writeFailure(streams.Stderr, fmt.Sprintf("output error: %v", err))
	}

	switch report.Verdict {
	case doctor.VerdictCompatible, doctor.VerdictConditional:
		return 0
	case doctor.VerdictMismatch, doctor.VerdictNotExecutable, doctor.VerdictUnknown:
		return 1
	default:
		return 2
	}
}

func writeFailure(output io.Writer, message string) int {
	if _, err := fmt.Fprintln(output, message); err != nil {
		return 2
	}

	return 2
}

func writeHuman(output io.Writer, report doctor.Report) error {
	targetOS := report.TargetOS
	if targetOS == "" {
		targetOS = "unknown"
	}
	targetArchitectures := strings.Join(report.Architectures, ",")
	if targetArchitectures == "" {
		targetArchitectures = "unknown"
	}

	if _, err := fmt.Fprintf(output, "format: %s\n", report.Format); err != nil {
		return fmt.Errorf("write format: %w", err)
	}
	if _, err := fmt.Fprintf(output, "target: %s/%s\n", targetOS, targetArchitectures); err != nil {
		return fmt.Errorf("write target: %w", err)
	}
	if _, err := fmt.Fprintf(output, "host: %s/%s\n", report.HostOS, report.HostArch); err != nil {
		return fmt.Errorf("write host: %w", err)
	}
	if _, err := fmt.Fprintf(output, "verdict: %s\n", report.Verdict); err != nil {
		return fmt.Errorf("write verdict: %w", err)
	}
	if _, err := fmt.Fprintf(output, "summary: %s\n", report.Summary); err != nil {
		return fmt.Errorf("write summary: %w", err)
	}
	for _, recommendation := range report.Recommendations {
		if _, err := fmt.Fprintf(output, "next: %s\n", recommendation); err != nil {
			return fmt.Errorf("write recommendation: %w", err)
		}
	}
	if report.FurtherHelpURL != "" {
		if _, err := fmt.Fprintf(output, "starter kit: %s\n", report.FurtherHelpURL); err != nil {
			return fmt.Errorf("write further help: %w", err)
		}
	}

	return nil
}
