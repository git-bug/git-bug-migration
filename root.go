// Package commands contains the CLI commands
package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/MichaelMure/git-bug-migration/migration1"
	"github.com/MichaelMure/git-bug-migration/migration2"
	"github.com/MichaelMure/git-bug-migration/migration3"
)

const rootCommandName = "git-bug-migration"

type rootOpts struct {
	forReal bool
}

func NewRootCommand() *cobra.Command {
	env := newEnv()
	opts := rootOpts{}

	migrations := []Migration{
		&migration1.Migration1{},
		&migration2.Migration2{},
		&migration3.Migration3{},
	}

	cmd := &cobra.Command{
		Use:   rootCommandName,
		Short: "A migration tool for git-bug",
		Long: `git-bug-migration is a tool to migrate git-bug data stored in git repository to a newer format when breaking 
changes are made in git-bug.

To migrate a repository, go to the corresponding repository and run "git-bug-migration".
	
	`,

		Version: version(),

		PreRunE: findRepo(env),
		RunE: func(_ *cobra.Command, _ []string) error {
			return runRootCmd(env, opts, migrations)
		},

		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	flags := cmd.Flags()
	flags.BoolVar(&opts.forReal, "for-real", false, "Indicate that your really want to run this tool and possibly ruin your data.")

	cmd.AddCommand(newVersionCommand())

	return cmd
}

func runRootCmd(env *Env, opts rootOpts, migrations []Migration) error {
	if !opts.forReal {
		env.err.Println("DISCLAIMER: This tool exist for your convenience to migrate your data and allow git-bug's authors" +
			" to break things and make it better. However, this migration tool is quite crude and experimental. DO NOT TRUST IT BLINDLY.\n\n" +
			"Please make a backup of your .git folder before running it.\n\n" +
			"When done, run this tool again with the --for-real flag.")
		os.Exit(1)
	}

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
