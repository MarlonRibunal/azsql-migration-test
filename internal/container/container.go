// Package container manages the lifecycle of a local Azure SQL Database
// Developer container: pull, start, wait-ready, and teardown.
package container

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/MarlonRibunal/azsql-migration-test/internal/config"
)

// Container is a running Azure SQL Database Developer instance.
type Container struct {
	Name       string
	Port       int
	SAPassword string
}

const containerName = "azsql-migration-test"

// Start logs in to the registry (when a registry + password are supplied), pulls
// the image, runs the container, and waits until the engine accepts connections.
//
// Prerequisite: Docker running, and an image configured via --image or
// AZDBDEV_IMAGE. If your registry requires authentication, supply the password
// (--registry-password / AZDBDEV_REGISTRY_PASSWORD) and Start runs `docker login`
// itself; otherwise login is skipped.
func Start(cfg *config.Config) (*Container, error) {
	if cfg.Image == "" {
		return nil, fmt.Errorf("no image configured: set --image or the AZDBDEV_IMAGE environment variable")
	}
	if err := login(cfg); err != nil {
		return nil, err
	}
	if err := run("docker", "pull", "--platform", config.Platform, cfg.Image); err != nil {
		return nil, fmt.Errorf("pull %s: %w (if the registry needs auth, supply --registry / --registry-password, or run docker login yourself)", cfg.Image, err)
	}

	// Remove any stale container left by a previous run.
	_ = exec.Command("docker", "rm", "-f", containerName).Run()

	err := run("docker", "run", "-d",
		"--platform", config.Platform,
		"--name", containerName,
		"-e", "ACCEPT_EULA=Y",
		"-e", "MSSQL_SA_PASSWORD="+cfg.SAPassword,
		// Bind to loopback only so the container (which uses the SA password) is
		// never reachable from the LAN.
		"-p", "127.0.0.1:"+strconv.Itoa(cfg.Port)+":1433",
		cfg.Image,
	)
	if err != nil {
		return nil, fmt.Errorf("run container: %w", err)
	}

	c := &Container{Name: containerName, Port: cfg.Port, SAPassword: cfg.SAPassword}
	if err := c.waitReady(90 * time.Second); err != nil {
		_ = c.Stop()
		return nil, err
	}
	return c, nil
}

// waitReady polls until the engine answers a trivial query.
func (c *Container) waitReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := c.ExecQuery("master", "SELECT 1"); err == nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("container did not become ready within %s", timeout)
}

// ConnString returns a connection string pointing at the local container.
func (c *Container) ConnString(db string) string {
	if db == "" {
		db = "master"
	}
	return fmt.Sprintf(
		"Server=localhost,%d;Database=%s;User Id=sa;Password=%s;TrustServerCertificate=True",
		c.Port, db, c.SAPassword,
	)
}

// sqlcmdPath is the in-container path to sqlcmd. The Developer image ships the
// mssql-tools18 build, which serves a self-signed cert and therefore needs the
// -C flag (trust server certificate) on every invocation.
const sqlcmdPath = "/opt/mssql-tools18/bin/sqlcmd"

// ExecQuery runs a single T-SQL batch inside the container via sqlcmd and
// returns the combined stdout+stderr. The -b flag makes sqlcmd exit non-zero on
// a SQL error, so a non-nil error means the batch failed; -C trusts the
// container's self-signed certificate.
func (c *Container) ExecQuery(db, query string) ([]byte, error) {
	if db == "" {
		db = "master"
	}
	cmd := exec.Command("docker", "exec", c.Name, sqlcmdPath,
		"-S", "localhost", "-U", "sa", "-P", c.SAPassword,
		"-C", "-d", db, "-b", "-Q", query)
	return cmd.CombinedOutput()
}

// identifierRe restricts database names to a safe SQL identifier so the name can
// be interpolated into DDL without injection.
var identifierRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,127}$`)

// EnsureDatabase creates db inside the container if it does not already exist.
func (c *Container) EnsureDatabase(db string) error {
	if !identifierRe.MatchString(db) {
		return fmt.Errorf("invalid database name %q: must match %s", db, identifierRe.String())
	}
	q := fmt.Sprintf("IF DB_ID(N'%s') IS NULL CREATE DATABASE [%s];", db, db)
	if out, err := c.ExecQuery("master", q); err != nil {
		return fmt.Errorf("create database %s: %w\n%s", db, err, out)
	}
	return nil
}

// Stop removes the container.
func (c *Container) Stop() error {
	return exec.Command("docker", "rm", "-f", c.Name).Run()
}

// login authenticates to the preview registry when a password is supplied.
// It is a no-op when the password is empty — the expected state once Developer
// is public/GA, or when the user has already run `docker login` themselves.
//
// The password is piped via stdin (never passed as an argument), so it does not
// appear in the process list.
func login(cfg *config.Config) error {
	if cfg.RegistryPassword == "" || cfg.Registry == "" {
		return nil
	}
	cmd := exec.Command("docker", "login", cfg.Registry, "-u", cfg.RegistryUser, "--password-stdin")
	cmd.Stdin = strings.NewReader(cfg.RegistryPassword)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker login %s as %s: %w", cfg.Registry, cfg.RegistryUser, err)
	}
	return nil
}

// run executes a command, forwarding its output to the user's terminal.
func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
