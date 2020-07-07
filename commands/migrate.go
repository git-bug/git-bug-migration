package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	mg1 "github.com/MichaelMure/git-bug-migration/migration1"
)

// These variables are initialized externally during the build. See the Makefile.
var GitCommit = "unset"
var GitLastTag = "unset"
var GitExactTag = "unset"

func runMigrateCmd(_ *cobra.Command, _ []string) error {
	// TODO: might be nicer under a --version flag
	fmt.Printf("%s version %s\n\n", rootCommandName, version())
	return mg1.Migrate01(repo)
}

func version() string {
	if GitExactTag == "undefined" {
		GitExactTag = ""
	}
	version := GitLastTag
	if GitExactTag == "" {
		version = fmt.Sprintf("%s-dev-%.10s", version, GitCommit)
	}
	return version
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate the repository to the newest version.",
	Long: `Migrate the repository from the current version to the newest available, bridging over breaking changes.

	If the repository is already at the latest version, this will leave it as is. If it detects legacy features, it
	will update them accordingly.`,

	PreRunE: loadRepo,
	RunE:    runMigrateCmd,
}
