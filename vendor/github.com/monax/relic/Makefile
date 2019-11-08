SHELL := /bin/bash
GO_FILES := $(shell go list -f "{{.Dir}}" ./...)

### Formatting, linting and vetting
.PHONY: check
check:
	@echo "Checking code for formatting style compliance."
	@gofmt -l -d ${GOFILES}
	@gofmt -l ${GOFILES} | read && echo && echo "Your marmot has found a problem with the formatting style of the code." 1>&2 && exit 1 || true
	@go mod tidy

# Just fix it
.PHONY: fix
fix:
	@goimports -l -w ${GO_FILES}

# test relic
.PHONY: test
test: check docs
	@go test ./...

### Release and versioning
.PHONY: version
version:
	@go run ./project/cmd/version/main.go

# Generate full changelog of all release notes
CHANGELOG.md: ./project/releases.go history.go
	@go run ./project/cmd/changelog/main.go > CHANGELOG.md

# Generated release notes for this version
NOTES.md: ./project/releases.go history.go
	@go run ./project/cmd/notes/main.go > NOTES.md

.PHONY: docs
docs: CHANGELOG.md NOTES.md

# Tag a release a push it
.PHONY: tag_release
tag_release: test check docs
	@scripts/tag_release.sh
