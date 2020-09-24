package main

import (
	"fmt"
	"io"
	"os"
	stdpath "path"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Env is the environment of a command
type Env struct {
	out out
	err out

	repoPath string
}

func newEnv() *Env {
	return &Env{
		out: out{Writer: os.Stdout},
		err: out{Writer: os.Stderr},
	}
}

type out struct {
	io.Writer
}

func (o out) Printf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(o, format, a...)
}

func (o out) Print(a ...interface{}) {
	_, _ = fmt.Fprint(o, a...)
}

func (o out) Println(a ...interface{}) {
	_, _ = fmt.Fprintln(o, a...)
}

// findRepo is a pre-run function that find the repository
func findRepo(env *Env) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("unable to get the current working directory: %q", err)
		}

		env.repoPath, err = detectGitPath(cwd)
		if err != nil {
			return errors.Wrap(err, "can't find the git repository")
		}

		return nil
	}
}

func detectGitPath(path string) (string, error) {
	// normalize the path
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	for {
		fi, err := os.Stat(stdpath.Join(path, ".git"))
		if err == nil {
			if !fi.IsDir() {
				return "", fmt.Errorf(".git exist but is not a directory")
			}
			return stdpath.Join(path, ".git"), nil
		}
		if !os.IsNotExist(err) {
			// unknown error
			return "", err
		}

		// detect bare repo
		ok, err := isGitDir(path)
		if err != nil {
			return "", err
		}
		if ok {
			return path, nil
		}

		if parent := filepath.Dir(path); parent == path {
			return "", fmt.Errorf(".git not found")
		} else {
			path = parent
		}
	}
}

func isGitDir(path string) (bool, error) {
	markers := []string{"HEAD", "objects", "refs"}

	for _, marker := range markers {
		_, err := os.Stat(stdpath.Join(path, marker))
		if err == nil {
			continue
		}
		if !os.IsNotExist(err) {
			// unknown error
			return false, err
		} else {
			return false, nil
		}
	}

	return true, nil
}
