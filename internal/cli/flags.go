package cli

import (
	"flag"

	"github.com/MarlonRibunal/azsql-migration-test/internal/config"
)

// bindCommonFlags registers the flags shared by all commands on fs and returns
// the Config they populate after fs.Parse.
func bindCommonFlags(fs *flag.FlagSet) *config.Config {
	cfg := &config.Config{}
	fs.StringVar(&cfg.Source, "source",
		config.EnvOr("AZDBDEV_SOURCE", ""),
		"source connection string (or set AZDBDEV_SOURCE to keep it out of shell history)")
	fs.StringVar(&cfg.Queries, "queries", "", "path to a .sql file of queries to replay")
	fs.StringVar(&cfg.Image, "image", config.Image(), "Azure SQL Database Developer container image")
	fs.IntVar(&cfg.Port, "port", 1433, "host port mapped to the container's 1433")
	fs.StringVar(&cfg.Database, "database", "MigrationValidation",
		"target database inside the container to deploy into and query")
	fs.StringVar(&cfg.SAPassword, "sa-password",
		config.EnvOr("AZDBDEV_SA_PASSWORD", "YourStrong!Passw0rd"),
		"SA password for the local container (or set AZDBDEV_SA_PASSWORD)")
	fs.StringVar(&cfg.ReportDir, "report-dir", "./migration-report", "directory for generated reports")
	fs.BoolVar(&cfg.Keep, "keep", false, "keep the container running after the run")

	// Private-preview registry login. Supply the password at runtime (flag or
	// AZDBDEV_REGISTRY_PASSWORD); leave it empty once Developer is public/GA to
	// skip the login step. Prefer the env var over the flag to keep the password
	// out of shell history and the process list.
	fs.StringVar(&cfg.Registry, "registry",
		config.EnvOr("AZDBDEV_REGISTRY", config.DefaultRegistry),
		"container registry to log in to (empty to skip login)")
	fs.StringVar(&cfg.RegistryUser, "registry-user",
		config.EnvOr("AZDBDEV_REGISTRY_USER", config.DefaultRegistryUser),
		"registry username")
	fs.StringVar(&cfg.RegistryPassword, "registry-password",
		config.EnvOr("AZDBDEV_REGISTRY_PASSWORD", ""),
		"registry password (or set AZDBDEV_REGISTRY_PASSWORD; empty skips login)")
	return cfg
}
