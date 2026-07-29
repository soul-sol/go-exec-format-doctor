// Command go-exec-format-doctor diagnoses executable format mismatches.
package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/soul-sol/go-exec-format-doctor/internal/app"
	"github.com/soul-sol/go-exec-format-doctor/internal/doctor"
)

func main() {
	host, err := doctor.NewHost(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		if _, writeErr := fmt.Fprintf(os.Stderr, "host error: %v\n", err); writeErr != nil {
			os.Exit(2)
		}
		os.Exit(2)
	}

	streams := app.Streams{Stdout: os.Stdout, Stderr: os.Stderr}
	os.Exit(app.Run(os.Args[1:], streams, host))
}
