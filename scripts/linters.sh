#!/usr/bin/env bash
echo "golang-ci lint..."
golangci-lint run ./...

echo "gogroup..."
gogroup -order std,other,prefix=github.com/obalunenko/  $(find . -type f -name "*.go" | grep -v "vendor/")