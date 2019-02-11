all: help

ifndef GOBIN
export GOBIN := $(GOPATH)/bin
endif

GOFILES := go list -f '{{range .GoFiles}}{{ $$.Dir }}/{{ . }} {{end}}{{range .TestGoFiles}}{{ $$.Dir }}/{{ . }} {{end}}' ./...

define dl
	@curl -sSq -L $(2) -o $(GOBIN)/$(1) && chmod u+x $(GOBIN)/$(1)
endef

define dl_tgz
	@curl -sSq -L $(2) | tar zxf - --strip 1 -C $(GOBIN) --wildcards '*/$(1)'
endef

.PHONY: setup
setup: GOPATH := $(shell mktemp -d)
setup: ## Install required tools
	@echo "==> Installing tools at $(GOBIN) ..."
	@mkdir -p $(GOBIN)
	@ln -s $(PWD)/vendor $(GOPATH)/src
	@$(call dl,dep,https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64)
	@$(call dl_tgz,golangci-lint,https://github.com/golangci/golangci-lint/releases/download/v1.10.2/golangci-lint-1.10.2-linux-amd64.tar.gz)
	@rm -Rf $(GOPATH)

.PHONY: setup-dev
setup-dev: GOPATH := $(shell mktemp -d)
setup-dev: ## Install required development tools
	@echo "==> Installing tools at $(GOBIN) ..."
	@mkdir -p $(GOBIN)
	@ln -s $(PWD)/vendor $(GOPATH)/src
	@$(call dl,swagger,https://github.com/go-swagger/go-swagger/releases/download/0.16.0/swagger_linux_amd64)
	@go install github.com/golang/mock/mockgen
	@go get gopkg.in/src-d/go-license-detector.v2/cmd/license-detector
	@rm -Rf $(GOPATH)

.PHONY: fmt
fmt: ## Format source code
	@go fmt ./...

.PHONY: check
check: ## Perform static code analysis
check: .check-copyright .check-timeutc .check-lint .check-vendor

.PHONY: .check-copyright
.check-copyright:
	@set -e; for f in `$(GOFILES)`; do \
		[[ $$f =~ /scyllaclient/internal/ ]] || \
		[[ $$f =~ /mermaidclient/internal/ ]] || \
		[[ $$f =~ /mock_.*_test[.]go ]] || \
		[ "`head -n 1 $$f`" == "// Copyright (C) 2017 ScyllaDB" ] || \
		(echo $$f; false); \
	done

.PHONY: .check-timeutc
.check-timeutc:
	@set -e; for f in `$(GOFILES)`; do \
		[[ $$f =~ /internal/timeutc/ ]] || \
		[[ $$f =~ /internal/retryablehttp/ ]] || \
		[[ $$f =~ /mermaidclient/internal/ ]] || \
		[[ $$f =~ /scyllaclient/internal/ ]] || \
		[ "`grep 'time.\(Now\|Parse(\|Since\)' $$f`" == "" ] || \
		(echo $$f; false); \
	done

.PHONY: .check-lint
.check-lint:
	@$(GOBIN)/golangci-lint run ./...

.PHONY: .check-vendor
.check-vendor:
	@$(GOBIN)/dep check

.PHONY: test
test: ## Run unit and integration tests
test: unit-test integration-test

.PHONY: unit-test
unit-test: ## Run unit tests
	@echo "==> Running tests (race)..."
	@go test -cover -race ./...

INTEGRATION_TEST_ARGS := -cluster 192.168.100.100 -managed-cluster 192.168.100.11

.PHONY: integration-test
integration-test: ## Run integration tests
	@echo "==> Running integration tests..."
	@go test -cover -race -v -tags integration -run Integration ./internal/cqlping $(INTEGRATION_TEST_ARGS)
	@go test -cover -race -v -tags integration -run Integration ./internal/ssh $(INTEGRATION_TEST_ARGS)
	@go test -cover -race -v -tags integration -run Integration ./scyllaclient $(INTEGRATION_TEST_ARGS)
	@go test -cover -race -v -tags integration -run Integration ./cluster $(INTEGRATION_TEST_ARGS)
	@go test -cover -race -v -tags integration -run Integration ./healthcheck $(INTEGRATION_TEST_ARGS)
	@go test -cover -race -v -tags integration -run Integration ./repair $(INTEGRATION_TEST_ARGS)
	@go test -cover -race -v -tags integration -run Integration ./sched $(INTEGRATION_TEST_ARGS)
	@go test -cover -race -v -tags integration -run Integration ./schema/cql $(INTEGRATION_TEST_ARGS)

.PHONY: dev-server
dev-server: ## Run development server
	@echo "==> Building development server..."
	@go build -race -o ./scylla-manager.dev ./cmd/scylla-manager
	@echo "==> Running development server..."
	@./scylla-manager.dev -c testing/scylla-manager/scylla-manager.yaml; rm -f ./scylla-manager.dev

.PHONY: dev-cli
dev-cli: ## Build development cli binary
	@echo "==> Building development cli..."
	@go build -o ./sctool.dev ./cmd/sctool/

.PHONY: generate
generate:  ## Recreate autogenerated resources
	@echo "==> Generating..."
	@go generate ./...

.PHONY: help
help:
	@awk -F ':|##' '/^[^\t].+?:.*?##/ {printf "\033[36m%-25s\033[0m %s\n", $$1, $$NF}' $(MAKEFILE_LIST)
