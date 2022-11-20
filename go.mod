module github.com/MichaelMure/git-bug-migration

go 1.14

require (
	github.com/99designs/keyring v1.2.1
	github.com/ProtonMail/go-crypto v0.0.0-20210428141323-04723f9f07d7
	github.com/blang/semver v3.5.1+incompatible
	github.com/blevesearch/bleve v1.0.14
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.9.0
	github.com/go-git/go-billy/v5 v5.3.1
	github.com/go-git/go-git/v5 v5.4.2
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.7.0
	golang.org/x/sys v0.0.0-20220204135822-1c1b9b1eba6a
	golang.org/x/text v0.3.3
)

replace golang.org/x/crypto => golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
