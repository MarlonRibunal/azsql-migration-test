// Package report writes the schema diff and query-replay results to disk.
package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MarlonRibunal/azsql-migration-test/internal/replay"
	"github.com/MarlonRibunal/azsql-migration-test/internal/schema"
)

// Write emits a Markdown summary (and, later, the raw schema-diff XML) into dir.
func Write(dir string, diff schema.Diff, results replay.Results) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("# Migration validation report\n\n")
	b.WriteString("## Schema\n\n")
	fmt.Fprintf(&b, "Changes: %d (breaking: %v)\n", diff.Changes, diff.Breaking())
	if diff.ReportPath != "" {
		fmt.Fprintf(&b, "DeployReport: `%s`\n", filepath.Base(diff.ReportPath))
	}
	b.WriteString("\n## Query replay\n\n")
	fmt.Fprintf(&b, "%d passed, %d failed\n\n", results.Passed(), results.Failed())
	for _, r := range results {
		status := "PASS"
		if !r.OK {
			status = "FAIL"
		}
		fmt.Fprintf(&b, "- [%s] %s (%s)\n", status, firstLine(r.Query), r.Duration)
		if !r.OK && r.Output != "" {
			fmt.Fprintf(&b, "  ```\n  %s\n  ```\n", indent(r.Output))
		}
	}

	return os.WriteFile(filepath.Join(dir, "report.md"), []byte(b.String()), 0o644)
}

// firstLine returns the first line of s, trimmed, for a compact report entry.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

// indent prefixes each line after the first with two spaces so multi-line tool
// output sits inside the report's fenced block.
func indent(s string) string {
	return strings.ReplaceAll(s, "\n", "\n  ")
}
