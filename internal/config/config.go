// Package config holds settings shared across the CLI commands.
package config

import "os"

// DefaultImage is the Azure SQL Database Developer container image reference.
//
// This is intentionally empty: the image lives in a registry you must configure.
// Set it with the AZDBDEV_IMAGE environment variable or the --image flag.
const DefaultImage = ""

// Platform is the container platform. The Developer image is published for
// linux/amd64 only, so this is forced (it runs under emulation on arm64 hosts).
const Platform = "linux/amd64"

// DefaultRegistry and DefaultRegistryUser are the registry login defaults.
//
// They are intentionally empty: configure them for your environment via the
// --registry / --registry-user flags or the AZDBDEV_REGISTRY / AZDBDEV_REGISTRY_USER
// environment variables. When both a registry and a password are provided the CLI
// runs `docker login`; otherwise it assumes you are already authenticated (or the
// image is public) and skips login.
const (
	DefaultRegistry     = ""
	DefaultRegistryUser = ""
)

// Config is the resolved set of options for a single command run.
type Config struct {
	Image      string // container image reference
	Port       int    // host port mapped to the container's 1433
	SAPassword string // SA password for the local container
	Source     string // source connection string (schema origin)
	Queries    string // path to a .sql file of queries to replay
	Database   string // target database inside the container to deploy into / query
	ReportDir  string // output directory for generated reports
	Keep       bool   // keep the container running after the run

	// Preview registry login. When RegistryPassword is empty (e.g. at GA, or when
	// the user has already run `docker login`), the login step is skipped.
	Registry         string
	RegistryUser     string
	RegistryPassword string
}

// Image returns the image reference, honoring AZDBDEV_IMAGE when set.
func Image() string {
	if v := os.Getenv("AZDBDEV_IMAGE"); v != "" {
		return v
	}
	return DefaultImage
}

// EnvOr returns the value of key, or def when key is unset/empty.
func EnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
