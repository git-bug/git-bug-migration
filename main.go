package main

import (
	"fmt"
	"os"

	"github.com/MichaelMure/git-bug-migration/migration1/bug"
	"github.com/MichaelMure/git-bug-migration/migration1/repository"
)

const rootCommandName = "git-bug-migration"

var repo repository.ClockedRepo

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("unable to get the current working directory: %q", err))
	}

	repo, err = repository.NewGitRepo(cwd, []repository.ClockLoader{bug.ClockLoader})
	if err == repository.ErrNotARepo {
		panic(fmt.Errorf("%s must be run from within a git repo", rootCommandName))
	}

	if err != nil {
		panic(err)
	}

	for streamedBug := range bug.ReadAllLocalBugs(repo) {
		if streamedBug.Err != nil {
			fmt.Print(fmt.Errorf("Got error when reading bug: %q", err))
			continue
		}

		//switch streamedBug.Bug
		fmt.Print(streamedBug.Bug.FirstOp().GetAuthor())
	}
}
