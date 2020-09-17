module github.com/MichaelMure/git-bug-migration

go 1.14

require (
	github.com/99designs/keyring v1.1.5
	github.com/blang/semver v3.5.1+incompatible
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.9.0
	github.com/go-git/go-git/v5 v5.1.0
	github.com/pkg/errors v0.9.1
	github.com/shurcooL/githubv4 v0.0.0-20200915023059-bc5e4feb2971
	github.com/shurcooL/graphql v0.0.0-20181231061246-d48a9a75455f // indirect
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.6.1
	github.com/xanzy/go-gitlab v0.37.0
	golang.org/x/oauth2 v0.0.0-20181106182150-f42d05182288
	golang.org/x/sync v0.0.0-20181221193216-37e7f081c4d4
	golang.org/x/text v0.3.3
)

// Use a forked go-git for now until https://github.com/go-git/go-git/pull/112 is merged
// and released.
replace github.com/go-git/go-git/v5 => github.com/MichaelMure/go-git/v5 v5.1.1-0.20200827115354-b40ca794fe33
