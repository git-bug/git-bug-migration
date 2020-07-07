package commands

import (
	mg1 "github.com/MichaelMure/git-bug-migration/migration1"
	"github.com/spf13/cobra"
)

func runMigrateCmd(_ *cobra.Command, _ []string) error {
	return mg1.Migrate01(repo)
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
