package main

import (
	"github.com/MichaelMure/git-bug-migration/commands"
)

type Migration interface {
	Name() string
	Description() string
	NeedToRun() bool
	Run() error
}

func main() {
	commands.Execute()
}
