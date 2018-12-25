NAME=kafka-dump

# COLORS
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)


TARGET_MAX_CHAR_NUM=20


define colored
	@echo '${GREEN}$1${RESET}'
endef

## Show help
help:
	${call colored, help is running...}
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

## dependensies - fetch all dependencies for sripts
dependencies:
	${call colored, dependensies is running...}
	./scripts/get-dependencies.sh

## Dev mode - go run
dev:
	${call colored, dev is running...}
	#docker-compose up&
	go run main.go
.PHONY: dev

## Compile binary
compile:
	${call colored, compile is running...}
	./scripts/compile.sh
.PHONY: compile

## lint project
lint:
	${call colored, lint is running...}
	./scripts/linters.sh
.PHONY: lint

## Test all packages
test:
	${call colored, test is running...}
	./scripts/tests.sh
.PHONY: test

## Test coverage
test-cover:
	${call colored, test-cover is running...}
	go test -race -coverpkg=./... -v -coverprofile .testCoverage.out ./...
	gocov convert .testCoverage.out | gocov report
.PHONY: test-cover

new-version:
	${call colored, new version is running...}
	./scripts/version.sh
.PHONY: new-version


## Release
release:
	${call colored, release is running...}
	./scripts/release.sh
.PHONY: release

.DEFAULT_GOAL := test
