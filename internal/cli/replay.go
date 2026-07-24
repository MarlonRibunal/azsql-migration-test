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

// runReplay runs only the query replay.
func runReplay(args []string) int {
	fs := flag.NewFlagSet("replay", flag.ContinueOnError)
	cfg := bindCommonFlags(fs)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if cfg.Queries == "" {
		fmt.Fprintln(os.Stderr, "error: --queries is required")
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

	results, err := replay.Run(cfg, c)
	if err != nil {
		fmt.Fprintln(os.Stderr, "query replay:", err)
		return 1
	}
	fmt.Printf("✓ Query replay complete: %d passed, %d failed\n", results.Passed(), results.Failed())

	if err := report.Write(cfg.ReportDir, schema.Diff{}, results); err != nil {
		fmt.Fprintln(os.Stderr, "write report:", err)
		return 1
	}
	if results.Failed() > 0 {
		return 1
	}
	return 0
}
