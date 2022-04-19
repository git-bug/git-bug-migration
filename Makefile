all: build

GIT_COMMIT:=$(shell git rev-list -1 HEAD)
GIT_LAST_TAG:=$(shell git describe --abbrev=0 --tags)
GIT_EXACT_TAG:=$(shell git name-rev --name-only --tags HEAD)

COMMANDS_PATH:=main
LDFLAGS:=-X '${COMMANDS_PATH}.GitCommit=${GIT_COMMIT}' \
	-X '${COMMANDS_PATH}.GitLastTag=${GIT_LAST_TAG}' \
	-X '${COMMANDS_PATH}.GitExactTag=${GIT_EXACT_TAG}'

build:
	go build -ldflags "$(LDFLAGS)" .

install:
	go install -ldflags "$(LDFLAGS)" .

releases:
	go run github.com/mitchellh/gox@v1.0.1 -ldflags "$(LDFLAGS)" -osarch '!darwin/386' -output "dist/{{.Dir}}_{{.OS}}_{{.Arch}}"

test:
	go test -v -bench=. ./...

.PHONY: build install releases test
