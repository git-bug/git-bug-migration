package main

type Migration interface {
	Description() string
	Run(repoPath string) error
}
