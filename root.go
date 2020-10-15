// Package commands contains the CLI commands
package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/MichaelMure/git-bug-migration/migration1"
	"github.com/MichaelMure/git-bug-migration/migration2"
)

const rootCommandName = "git-bug-migration"

func NewRootCommand() *cobra.Command {
	env := newEnv()

	migrations := []Migration{
		&migration1.Migration1{},
		&migration2.Migration2{},
	}

	cmd := &cobra.Command{
		Use:   rootCommandName,
		Short: "A migration tool for git-bug",
		Long: `git-bug-migration is a tool to migrate git-bug data stored in git repository to a newer format when breaking 
changes are made in git-bug.
	
	`,

		Version: version(),

		PreRunE: findRepo(env),
		RunE: func(_ *cobra.Command, _ []string) error {
			return runRootCmd(env, migrations)
		},

		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newVersionCommand())

	return cmd
}

func runRootCmd(env *Env, migrations []Migration) error {
	for i, migration := range migrations {
		if i > 0 {
			env.out.Println()
		}
		env.out.Printf("Migration #%d\n", i+1)
		env.out.Println("Purpose:", migration.Description())
		env.out.Println()

		err := migration.Run(env.repoPath)
		if err != nil {
			env.err.Printf("Error applying migration: %v\n", err)
			os.Exit(1)
		}
		env.out.Println()
	}
	env.out.Println("\nDone!")
	return nil
}
