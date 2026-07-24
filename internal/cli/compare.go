package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/MarlonRibunal/azsql-migration-test/internal/container"
	"github.com/MarlonRibunal/azsql-migration-test/internal/replay"
	"github.com/MarlonRibunal/azsql-migration-test/internal/report"
	"github.com/MarlonRibunal/azsql-migration-test/internal/schema"
)

// runCompare runs only the schema diff.
func runCompare(args []string) int {
	fs := flag.NewFlagSet("compare", flag.ContinueOnError)
	cfg := bindCommonFlags(fs)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if cfg.Source == "" {
		fmt.Fprintln(os.Stderr, "error: --source is required")
		return 2
	}

	c, err := container.Start(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "container start:", err)
		return 1
	}
	if !cfg.Keep {
		defer c.Stop()
	}

	diff, err := schema.Compare(cfg, c)
	if err != nil {
		fmt.Fprintln(os.Stderr, "schema compare:", err)
		return 1
	}
	fmt.Println("✓ Schema comparison complete")

	if err := report.Write(cfg.ReportDir, diff, replay.Results{}); err != nil {
		fmt.Fprintln(os.Stderr, "write report:", err)
		return 1
	}
	if diff.Breaking() {
		return 1
	}
	return 0
}
