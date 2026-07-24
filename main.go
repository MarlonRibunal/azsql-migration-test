// Command azsql-migration-test validates Azure SQL migrations locally by running
// them against the Azure SQL Database Developer container.
package main

import (
	"os"

	"github.com/MarlonRibunal/azsql-migration-test/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
