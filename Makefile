NAME=kafka-dump
BIN_DIR=./bin

# COLORS
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)


TARGET_MAX_CHAR_NUM=20
## Show help
help:
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  ${YELLOW}%-$(TARGET_MAX_CHAR_NUM)s${RESET} ${GREEN}%s${RESET}\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

default: dev

## Dev mode - go run
dev:
	go run main.go

## Compile binary
compile:
	mkdir -p ${BIN_DIR}
	go build -o ${BIN_DIR}/${NAME}

## lint project
lint:
	go vet -composites=false $(go list ./... | grep -v /vendor/)
	gometalinter --vendor --disable=gotype --linter='errcheck:errcheck -blank . :PATH:LINE:COL:MESSAGE'

## Test all packages
test:
	go test -v ./...

## Test coverage
test-cover:
	go test -race -coverpkg=./... -v -coverprofile .testCoverage.out ./...
	gocov convert .testCoverage.out | gocov report

## Release
release: test test-cover compile



