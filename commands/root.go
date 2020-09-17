// Package commands contains the CLI commands
package commands

// Imported from git-bug and edited accordingly

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	mg1 "github.com/MichaelMure/git-bug-migration/migration01"
	mg1b "github.com/MichaelMure/git-bug-migration/migration01/after/bug"
	mg1r "github.com/MichaelMure/git-bug-migration/migration01/after/repository"
)

const rootCommandName = "git-bug-migration"

var (
	repo mg1r.ClockedRepo

	// These variables are initialized externally during the build. See the Makefile.
	GitCommit   = ""
	GitLastTag  = ""
	GitExactTag = ""

	rootCmdShowVersion bool
)

func runRootCmd(_ *cobra.Command, args []string) error {
	if rootCmdShowVersion {
		fmt.Printf(version())
		return nil
	}

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
	RunE:    runRootCmd,

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

func init() {
	RootCmd.Flags().SortFlags = false

	RootCmd.Flags().BoolVarP(&rootCmdShowVersion, "version", "v", false,
		"Show the version of the migration tool. This will not run the tool.")
}
