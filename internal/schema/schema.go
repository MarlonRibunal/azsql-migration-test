// Package schema extracts a source schema and validates it against the local
// Azure SQL Database Developer container using sqlpackage.
package schema

import (
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/MarlonRibunal/azsql-migration-test/internal/config"
	"github.com/MarlonRibunal/azsql-migration-test/internal/container"
)

// Diff is the result of a schema comparison (a sqlpackage DeployReport).
type Diff struct {
	ReportPath string // path to the generated DeployReport XML
	Changes    int    // number of schema operations (items) in the plan
	breaking   bool   // whether the plan drops objects or raises data-loss alerts
}

// Breaking reports whether the diff contains a breaking change.
func (d Diff) Breaking() bool { return d.breaking }

// Compare extracts the source schema, ensures the target database exists, and
// runs a non-destructive DeployReport (source vs. target). It does NOT modify
// the target schema. Used by the `compare` command.
func Compare(cfg *config.Config, c *container.Container) (Diff, error) {
	dacpac, cleanup, err := extract(cfg)
	if err != nil {
		return Diff{}, err
	}
	defer cleanup()
	return deployReport(cfg, c, dacpac)
}

// Deploy extracts the source schema, runs a DeployReport for the artifact, then
// PUBLISHES the schema into the target database. The publish is the real
// compatibility test: if the schema uses a construct the Azure SQL Database
// engine rejects, the publish fails and the migration is not compatible. Used by
// the `validate` command so query replay runs against the real deployed schema.
func Deploy(cfg *config.Config, c *container.Container) (Diff, error) {
	dacpac, cleanup, err := extract(cfg)
	if err != nil {
		return Diff{}, err
	}
	defer cleanup()

	diff, err := deployReport(cfg, c, dacpac)
	if err != nil {
		return Diff{}, err
	}
	if err := publish(cfg, c, dacpac); err != nil {
		return diff, err
	}
	return diff, nil
}

// extract exports the source schema to a temporary .dacpac and returns its path
// plus a cleanup func that removes the temp directory.
func extract(cfg *config.Config) (dacpac string, cleanup func(), err error) {
	if _, err := exec.LookPath("sqlpackage"); err != nil {
		return "", func() {}, fmt.Errorf("sqlpackage not found on PATH: %w", err)
	}
	dir, err := os.MkdirTemp("", "azsql-migration-test-")
	if err != nil {
		return "", func() {}, err
	}
	cleanup = func() { _ = os.RemoveAll(dir) }

	dacpac = filepath.Join(dir, "source.dacpac")
	err = runTool("sqlpackage",
		"/Action:Extract",
		"/SourceConnectionString:"+cfg.Source,
		"/TargetFile:"+dacpac,
		"/p:ExtractAllTableData=false",
	)
	if err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("extract source schema: %w", err)
	}
	return dacpac, cleanup, nil
}

// deployReport ensures the target database exists and generates + parses a
// DeployReport comparing the dacpac to it.
func deployReport(cfg *config.Config, c *container.Container, dacpac string) (Diff, error) {
	if err := c.EnsureDatabase(cfg.Database); err != nil {
		return Diff{}, err
	}
	if err := os.MkdirAll(cfg.ReportDir, 0o755); err != nil {
		return Diff{}, err
	}
	xmlPath := filepath.Join(cfg.ReportDir, "schema-diff.xml")

	err := runTool("sqlpackage",
		"/Action:DeployReport",
		"/SourceFile:"+dacpac,
		"/TargetConnectionString:"+c.ConnString(cfg.Database),
		"/OutputPath:"+xmlPath,
	)
	if err != nil {
		return Diff{}, fmt.Errorf("deploy report: %w", err)
	}
	return parseReport(xmlPath)
}

// publish deploys the dacpac into the target database (the real compat test).
func publish(cfg *config.Config, c *container.Container, dacpac string) error {
	err := runTool("sqlpackage",
		"/Action:Publish",
		"/SourceFile:"+dacpac,
		"/TargetConnectionString:"+c.ConnString(cfg.Database),
	)
	if err != nil {
		return fmt.Errorf("publish schema to container (incompatible with Azure SQL Database?): %w", err)
	}
	return nil
}

// deploymentReport mirrors the subset of the sqlpackage DeployReport XML we care
// about. Field tags match by local name, so the document namespace is ignored.
type deploymentReport struct {
	XMLName xml.Name `xml:"DeploymentReport"`
	Alerts  []struct {
		Name string `xml:"Name,attr"`
	} `xml:"Alerts>Alert"`
	Operations []struct {
		Name  string `xml:"Name,attr"`
		Items []struct {
			Value string `xml:"Value,attr"`
			Type  string `xml:"Type,attr"`
		} `xml:"Item"`
	} `xml:"Operations>Operation"`
}

// parseReport reads a DeployReport XML file and summarizes it into a Diff.
// A change is any operation item; the plan is breaking if it drops an object or
// raises a data-loss alert.
func parseReport(path string) (Diff, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Diff{ReportPath: path}, fmt.Errorf("read deploy report: %w", err)
	}
	var rep deploymentReport
	if err := xml.Unmarshal(raw, &rep); err != nil {
		return Diff{ReportPath: path}, fmt.Errorf("parse deploy report: %w", err)
	}

	changes := 0
	breaking := len(rep.Alerts) > 0
	for _, op := range rep.Operations {
		changes += len(op.Items)
		if op.Name == "Drop" || op.Name == "TableRebuild" {
			breaking = true
		}
	}
	return Diff{ReportPath: path, Changes: changes, breaking: breaking}, nil
}

// runTool executes an external tool, streaming its output to the terminal.
func runTool(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
