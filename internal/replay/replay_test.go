package replay

import (
	"errors"
	"testing"
)

// fakeExec records calls and returns a scripted error per batch.
type fakeExec struct {
	fail map[string]bool
	seen []string
}

func (f *fakeExec) ExecQuery(db, query string) ([]byte, error) {
	f.seen = append(f.seen, query)
	if f.fail[query] {
		return []byte("Msg 208: Invalid object name"), errors.New("exit status 1")
	}
	return []byte("(1 rows affected)"), nil
}

func TestRunRecordsPassAndFail(t *testing.T) {
	ex := &fakeExec{fail: map[string]bool{"SELECT 2;": true}}
	got := run([]string{"SELECT 1;", "SELECT 2;"}, "db", ex)
	if len(got) != 2 {
		t.Fatalf("got %d results, want 2", len(got))
	}
	if !got[0].OK {
		t.Errorf("batch 0 should pass")
	}
	if got[1].OK {
		t.Errorf("batch 1 should fail")
	}
	if got[1].Output == "" {
		t.Errorf("failed batch should capture output")
	}
	if got.Passed() != 1 || got.Failed() != 1 {
		t.Errorf("counts = %d passed / %d failed, want 1/1", got.Passed(), got.Failed())
	}
}

func TestSplitBatches(t *testing.T) {
	script := "SELECT 1;\nGO\n\nSELECT 2;\ngo\n"
	got := SplitBatches(script)
	want := []string{"SELECT 1;", "SELECT 2;"}
	if len(got) != len(want) {
		t.Fatalf("got %d batches, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("batch %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSplitBatchesEmpty(t *testing.T) {
	if got := SplitBatches("\n\nGO\n\n"); len(got) != 0 {
		t.Errorf("expected no batches, got %#v", got)
	}
}

func TestResultsCounts(t *testing.T) {
	r := Results{{OK: true}, {OK: false}, {OK: true}}
	if r.Passed() != 2 {
		t.Errorf("Passed() = %d, want 2", r.Passed())
	}
	if r.Failed() != 1 {
		t.Errorf("Failed() = %d, want 1", r.Failed())
	}
}
