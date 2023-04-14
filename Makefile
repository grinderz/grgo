SHELL := /usr/bin/env bash -o errtrace -o pipefail -o noclobber -o errexit -o nounset

DOCKER_GOLANGCI_LINT_VERSION := 1.52.2
DOCKER_GOLANGCI_LINT_TIMEOUT := 5m

.PHONY: go.generate
go.generate:
	@go generate ./...

.PHONY: go.fmt
go.fmt:
	@gofmt -w -s .

.PHONY: docker.lint
docker.lint:
	docker run -t --rm -v $$(pwd):/app -v ~/.cache/golangci-lint/v$(DOCKER_GOLANGCI_LINT_VERSION):/root/.cache -w /app golangci/golangci-lint:v$(DOCKER_GOLANGCI_LINT_VERSION) golangci-lint run --timeout=$(DOCKER_GOLANGCI_LINT_TIMEOUT)

.PHONY: lint
lint:
	@golangci-lint run --timeout=$(DOCKER_GOLANGCI_LINT_TIMEOUT)

.PHONY: pre-commit
pre-commit:
	@pre-commit run --all-files
