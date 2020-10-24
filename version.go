package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// These variables are initialized externally during the build. See the Makefile.
var GitCommit = ""
var GitLastTag = ""
var GitExactTag = ""

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

func newVersionCommand() *cobra.Command {
	env := newEnv()

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show git-bug-migration version information.",
		Run: func(cmd *cobra.Command, args []string) {
			env.out.Println(version())
		},
	}

	return cmd
}
