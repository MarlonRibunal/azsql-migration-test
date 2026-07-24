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

// runValidate runs the full pass: schema comparison + query replay.
func runValidate(args []string) int {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
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

	diff, err := schema.Deploy(cfg, c)
	if err != nil {
		fmt.Fprintln(os.Stderr, "schema deploy:", err)
		return 1
	}
	fmt.Println("✓ Schema deployed to Azure SQL Database Developer (compatible)")

	var results replay.Results
	if cfg.Queries != "" {
		results, err = replay.Run(cfg, c)
		if err != nil {
			fmt.Fprintln(os.Stderr, "query replay:", err)
			return 1
		}
		fmt.Printf("✓ Query replay complete: %d passed, %d failed\n", results.Passed(), results.Failed())
	}

	if err := report.Write(cfg.ReportDir, diff, results); err != nil {
		fmt.Fprintln(os.Stderr, "write report:", err)
		return 1
	}

	if diff.Breaking() || results.Failed() > 0 {
		fmt.Println("Migration validation FAILED")
		return 1
	}
	fmt.Println("Migration validation completed successfully")
	return 0
}
