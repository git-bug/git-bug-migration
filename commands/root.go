// Package commands contains the CLI commands
package commands

// Imported from git-bug and edited accordingly

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	mg1b "github.com/MichaelMure/git-bug-migration/migration1/bug"
	mg1r "github.com/MichaelMure/git-bug-migration/migration1/repository"
)

const rootCommandName = "git-bug-migration"

var repo mg1r.ClockedRepo

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   rootCommandName,
	Short: "A migratory tool for git-bug",
	Long: `git-bug-migration is a tool to migrate git-bug data stored in git repository to a newer format when breaking 
	changes are made in git-bug.
	
	`,

	// For the root command, force the execution of the PreRun
	// even if we just display the help. This is to make sure that we check
	// the repository and give the user early feedback.
	PreRunE: loadRepo,
	RunE:    runMigrateCmd,

	SilenceUsage:      true,
	DisableAutoGenTag: true,

	// Custom bash code to connect the git completion for "git bug migration" to the
	// git-bug-migration completion for "git-bug-migration"
	BashCompletionFunction: `
	_git_bug_migration() {
		__start_git-bug-migration "$@"
	}
	`,
}

// loadRepo is a pre-run function that load the repository for use in a command
func loadRepo(_ *cobra.Command, _ []string) error {
	cwd, err := os.Getwd()

	if err != nil {
		panic(fmt.Errorf("unable to get the current working directory: %q", err))
	}

	repo, err = mg1r.NewGitRepo(cwd, []mg1r.ClockLoader{mg1b.ClockLoader})
	if err == mg1r.ErrNotARepo {
		return fmt.Errorf("%s must be run from within a git repo", rootCommandName)
	} else if err != nil {
		return err
	}

	return nil
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
