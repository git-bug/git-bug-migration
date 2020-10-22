package main

import "os"

func main() {
	if err := NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
