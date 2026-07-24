// Package replay runs a set of T-SQL queries against the local container and
// records pass/fail and timing for each.
package replay

import (
	"os"
	"strings"
	"time"

	"github.com/MarlonRibunal/azsql-migration-test/internal/config"
	"github.com/MarlonRibunal/azsql-migration-test/internal/container"
)

// executor runs one T-SQL batch and returns the tool output. Container satisfies
// this via ExecQuery; splitting it out keeps Run unit-testable.
type executor interface {
	ExecQuery(db, query string) ([]byte, error)
}

// Result is one query batch's outcome.
type Result struct {
	Query    string
	OK       bool
	Duration time.Duration
	Output   string // sqlcmd output (captured on failure for the report)
	Err      error
}

// Results is the full replay set.
type Results []Result

// Passed counts successful batches.
func (r Results) Passed() int {
	n := 0
	for _, x := range r {
		if x.OK {
			n++
		}
	}
	return n
}

// Failed counts failed batches.
func (r Results) Failed() int { return len(r) - r.Passed() }

// Run executes each batch in cfg.Queries against the container's target
// database, capturing pass/fail and timing per batch.
//
// Timing here is informational only — a local container is not a cloud service
// tier, so it is NOT a cloud performance signal. Do not headline it as one.
func Run(cfg *config.Config, c *container.Container) (Results, error) {
	raw, err := os.ReadFile(cfg.Queries)
	if err != nil {
		return nil, err
	}
	return run(SplitBatches(string(raw)), cfg.Database, c), nil
}

// run executes batches against an executor, timing each. It never returns an
// error itself — per-batch failures are recorded on the Result.
func run(batches []string, db string, ex executor) Results {
	results := make(Results, 0, len(batches))
	for _, b := range batches {
		start := time.Now()
		out, err := ex.ExecQuery(db, b)
		results = append(results, Result{
			Query:    b,
			OK:       err == nil,
			Duration: time.Since(start),
			Output:   strings.TrimSpace(string(out)),
			Err:      err,
		})
	}
	return results
}

// SplitBatches splits a T-SQL script into batches on standalone GO separators.
func SplitBatches(s string) []string {
	var out []string
	var cur []string
	flush := func() {
		if b := strings.TrimSpace(strings.Join(cur, "\n")); b != "" {
			out = append(out, b)
		}
		cur = cur[:0]
	}
	for _, line := range strings.Split(s, "\n") {
		if strings.EqualFold(strings.TrimSpace(line), "GO") {
			flush()
			continue
		}
		cur = append(cur, line)
	}
	flush()
	return out
}
