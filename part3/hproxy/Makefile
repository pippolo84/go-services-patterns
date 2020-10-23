#
# Variables
#

# Go version
GOVERSION = 1.15

#
# Build targets
#

# local
build: pre-build	## build binaries for current architecure
	go build -race -o bin/server cmd/server/main.go
	go build -race -o bin/hproxy cmd/hproxy/main.go

.PHONY: build

#
# Code quality
#

# test
unit: pre-build		## run unit tests
	go test -v -race ./...

.PHONY: unit

integration: pre-build	## run integration tests
	go test -v -race --tags=integration ./...

.PHONY: integration

# lint
lint: pre-build		## run linter on source code
	golangci-lint run

.PHONY: lint

#
# Build Environment
#

pre-build: envcheck

envcheck:
	@go version | grep -q $(GOVERSION) || printf "\nPlease install go $(GOVERSION) from https://golang.org/dl/\n\n"

.PHONY: pre-build envcheck

#
# Misc
#

clean:
	go clean
	rm bin/server
	rm bin/hproxy

.PHONY: clean
