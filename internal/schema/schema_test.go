package schema

import "testing"

func TestParseReportCreatesOnly(t *testing.T) {
	d, err := parseReport("testdata/report-creates.xml")
	if err != nil {
		t.Fatal(err)
	}
	if d.Changes != 4 {
		t.Errorf("Changes = %d, want 4", d.Changes)
	}
	if d.Breaking() {
		t.Error("a create-only plan must not be flagged breaking")
	}
}

func TestParseReportBreaking(t *testing.T) {
	d, err := parseReport("testdata/report-breaking.xml")
	if err != nil {
		t.Fatal(err)
	}
	if d.Changes != 2 {
		t.Errorf("Changes = %d, want 2", d.Changes)
	}
	if !d.Breaking() {
		t.Error("a plan with a Drop and a data-loss Alert must be flagged breaking")
	}
}

func TestParseReportMissingFile(t *testing.T) {
	if _, err := parseReport("testdata/does-not-exist.xml"); err == nil {
		t.Error("expected an error for a missing report file")
	}
}
