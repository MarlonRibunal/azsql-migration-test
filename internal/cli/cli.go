// Package cli wires the command-line interface: argument dispatch and the
// per-command handlers (validate, compare, replay).
package cli

import (
	"fmt"
	"os"
	"runtime/debug"
)

// Version is overridden at build time via -ldflags (make build / make release).
var Version = "dev"

// version returns the ldflags value, or falls back to the module version stamped
// by the Go toolchain for `go install module@vX` builds (which do not run
// ldflags), so those users see a real version instead of "dev".
func version() string {
	if Version != "dev" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return Version
}

// Run dispatches args to a subcommand and returns a process exit code.
func Run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	switch args[0] {
	case "validate":
		return runValidate(args[1:])
	case "compare":
		return runCompare(args[1:])
	case "replay":
		return runReplay(args[1:])
	case "version", "--version", "-v":
		fmt.Println("azsql-migration-test", version())
		return 0
	case "help", "-h", "--help":
		usage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", args[0])
		usage()
		return 2
	}
}

func usage() {
	fmt.Print(`azsql-migration-test — validate Azure SQL migrations against Azure SQL Database Developer

Usage:
  azsql-migration-test <command> [flags]

Commands:
  validate   Full pass: schema comparison + query replay
  compare    Schema diff (source vs. target) via sqlpackage
  replay     Run queries against the container, report execution + timing
  version    Print version
  help       Show this help

Run "azsql-migration-test <command> -h" for command flags.
`)
}
